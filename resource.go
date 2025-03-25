package judge

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	seccomp "github.com/seccomp/libseccomp-golang"
)

// setupCgroup 设置cgroup限制
func (j *Judge) setupCgroup() error {
	// 创建cgroup目录
	if err := os.MkdirAll(j.config.Files.CgroupPath, 0755); err != nil {
		return fmt.Errorf("failed to create cgroup directory: %w", err)
	}

	// 设置CPU限制
	if j.config.Limits.CPU > 0 {
		cpuLimit := j.config.Limits.CPU.Microseconds()
		if err := os.WriteFile(
			filepath.Join(j.config.Files.CgroupPath, "cpu.max"),
			[]byte(fmt.Sprintf("%d 100000", cpuLimit)),
			0644,
		); err != nil {
			return fmt.Errorf("failed to set cpu limit: %w", err)
		}
	}

	// 设置内存限制
	if j.config.Limits.Memory > 0 {
		if err := os.WriteFile(
			filepath.Join(j.config.Files.CgroupPath, "memory.max"),
			[]byte(strconv.FormatInt(j.config.Limits.Memory*1024, 10)),
			0644,
		); err != nil {
			return fmt.Errorf("failed to set memory limit: %w", err)
		}
	}

	return nil
}

// setupSeccomp 设置seccomp限制
func (j *Judge) setupSeccomp() error {
	if j.config.Security.Syscalls == nil {
		j.config.Security.Syscalls = make([]bool, 512)
		for i := range j.config.Security.Syscalls {
			j.config.Security.Syscalls[i] = true
		}
	}

	filter, err := seccomp.NewFilter(seccomp.ActAllow)
	if err != nil {
		return fmt.Errorf("error creating seccomp filter: %w", err)
	}

	for i, allowed := range j.config.Security.Syscalls {
		if !allowed {
			err = filter.AddRule(seccomp.ScmpSyscall(i), seccomp.ActErrno)
			if err != nil {
				return fmt.Errorf("error adding rule to seccomp filter for syscall %d: %w", i, err)
			}
		}
	}

	err = filter.Load()
	if err != nil {
		return fmt.Errorf("error loading seccomp filter: %w", err)
	}

	return nil
}

// getMemoryUsage 获取进程内存使用情况
func getMemoryUsage(pid int) (int64, error) {
	statusFile := fmt.Sprintf("/proc/%d/status", pid)
	data, err := os.ReadFile(statusFile)
	if err != nil {
		return 0, fmt.Errorf("failed to read status file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "VmRSS:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				memory, err := strconv.ParseInt(fields[1], 10, 64)
				if err != nil {
					return 0, fmt.Errorf("failed to parse memory value: %w", err)
				}
				return memory * 1024, nil // 转换为字节
			}
		}
	}

	return 0, fmt.Errorf("memory usage not found")
}
