package judge

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	seccomp "github.com/seccomp/libseccomp-golang"
)

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
