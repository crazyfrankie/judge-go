package judge

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

// Judge 表示一个评测实例
type Judge struct {
	config *Config
	result *Result
	mu     sync.Mutex
	done   chan struct{}
}

// Config 评测配置
type Config struct {
	// 资源限制
	Limits struct {
		CPU    time.Duration
		Memory int64
		Stack  int64
		Output int64
	}
	// 执行配置
	Exec struct {
		Path string
		Args []string
		Env  []string
	}
	// 安全配置
	Security struct {
		UID      int
		GID      int
		Chroot   string
		Syscalls []bool
	}
	// 文件配置
	Files struct {
		UserOutput string
		StdOutput  string
		CgroupPath string
	}
}

// NewJudge 创建一个新的评测实例
func NewJudge(config *Config) *Judge {
	return &Judge{
		config: config,
		result: &Result{},
		done:   make(chan struct{}),
	}
}

// Run 执行评测
func (j *Judge) Run(ctx context.Context) (*Result, error) {
	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(ctx, j.config.Limits.CPU)
	defer cancel()

	// 创建命令
	cmd := exec.CommandContext(ctx, j.config.Exec.Path, j.config.Exec.Args...)
	cmd.Env = j.config.Exec.Env

	// 设置输出
	outputFile, err := os.Create(j.config.Files.UserOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()
	cmd.Stdout = outputFile

	// 启动进程
	if err := j.startProcess(cmd); err != nil {
		return nil, err
	}

	// 等待进程结束
	if err := j.waitProcess(cmd); err != nil {
		return nil, err
	}

	return j.result, nil
}

// startProcess 启动进程并设置资源限制
func (j *Judge) startProcess(cmd *exec.Cmd) error {
	// 设置资源限制
	if err := j.setupResourceLimits(cmd); err != nil {
		return err
	}

	// 设置安全限制
	if err := j.setupSecurity(cmd); err != nil {
		return err
	}

	// 启动进程
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}

	return nil
}

// waitProcess 等待进程结束并收集结果
func (j *Judge) waitProcess(cmd *exec.Cmd) error {
	startTime := time.Now().UnixMilli()

	// 启动goroutine监控内存使用
	go j.monitorMemory(cmd.Process.Pid)

	// 等待进程结束
	err := cmd.Wait()
	endTime := time.Now().UnixMilli()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			j.result.Signal = exitErr.ExitCode()
			return nil
		}
		return fmt.Errorf("process error: %w", err)
	}

	// 更新时间统计
	j.result.RealTimeUsed = endTime - startTime

	// 获取资源使用情况
	if rusage, ok := cmd.ProcessState.SysUsage().(*syscall.Rusage); ok {
		j.result.CpuTimeUsed = int64(rusage.Utime.Sec)*1000 +
			int64(rusage.Utime.Usec)/1000 +
			int64(rusage.Stime.Sec)*1000 +
			int64(rusage.Stime.Usec)/1000
	}

	return nil
}

// setupResourceLimits 设置资源限制
func (j *Judge) setupResourceLimits(cmd *exec.Cmd) error {
	// 设置cgroup
	if err := j.setupCgroup(); err != nil {
		return err
	}

	// 设置栈大小限制
	if j.config.Limits.Stack > 0 {
		if err := syscall.Setrlimit(syscall.RLIMIT_STACK, &syscall.Rlimit{
			Cur: uint64(j.config.Limits.Stack),
			Max: uint64(j.config.Limits.Stack),
		}); err != nil {
			return fmt.Errorf("failed to set stack limit: %w", err)
		}
	}

	// 设置输出大小限制
	if j.config.Limits.Output > 0 {
		if err := syscall.Setrlimit(syscall.RLIMIT_FSIZE, &syscall.Rlimit{
			Cur: uint64(j.config.Limits.Output),
			Max: uint64(j.config.Limits.Output),
		}); err != nil {
			return fmt.Errorf("failed to set output size limit: %w", err)
		}
	}

	return nil
}

// setupSecurity 设置安全限制
func (j *Judge) setupSecurity(cmd *exec.Cmd) error {
	// 设置seccomp
	if err := j.setupSeccomp(); err != nil {
		return err
	}

	// 设置用户和组ID
	if j.config.Security.UID != 0 || j.config.Security.GID != 0 {
		cmd.SysProcAttr.Credential = &syscall.Credential{
			Uid: uint32(j.config.Security.UID),
			Gid: uint32(j.config.Security.GID),
		}
	}

	return nil
}

// monitorMemory 监控内存使用
func (j *Judge) monitorMemory(pid int) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-j.done:
			return
		case <-ticker.C:
			if usage, err := getMemoryUsage(pid); err == nil {
				j.mu.Lock()
				if usage > j.result.MemoryUsed {
					j.result.MemoryUsed = usage
				}
				j.mu.Unlock()
			}
		}
	}
}

// Check 检查输出结果
func (j *Judge) Check() (int, error) {
	return StdCheck(j.config.Files.UserOutput, j.config.Files.StdOutput)
}

// Close 清理资源
func (j *Judge) Close() error {
	close(j.done)
	return os.RemoveAll(j.config.Files.CgroupPath)
}

func AllowSyscall(syscallRule []bool, syscallId uint64) bool {
	return syscallRule[int(syscallId)]
}

// 修改时间相关的类型转换
func (j *Judge) updateTimes(startTime, endTime int64, ru *syscall.Rusage) {
	j.result.CpuTimeUsed = int64(ru.Utime.Sec)*1000 + int64(ru.Utime.Usec)/1000 +
		int64(ru.Stime.Sec)*1000 + int64(ru.Stime.Usec)/1000
	j.result.RealTimeUsed = endTime - startTime
}
