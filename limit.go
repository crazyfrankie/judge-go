package judge

import (
	"fmt"
	"os"
	"strconv"
	"syscall"

	seccomp "github.com/seccomp/libseccomp-golang"
	"golang.org/x/sys/unix"
)

type Limit struct {
	CpuTime    int
	Memory     int
	Stack      int
	OutputSize int
}

func createCgroup(cgroupPath string, limit *Limit) error {
	var rl unix.Rlimit

	// create a Cgroups
	if err := os.MkdirAll(cgroupPath, 0755); err != nil {
		return fmt.Errorf("failed to create cgroup directory: %v", err)
	}

	// cpu time limit (MS)
	if limit.CpuTime > 0 {
		cpuLimit := limit.CpuTime * 1000
		if err := os.WriteFile(cgroupPath+"/cpu.cfs_quota_us", []byte(strconv.Itoa(cpuLimit)), 0644); err != nil {
			return fmt.Errorf("failed to set cpu limit: %v", err)
		}
	}

	// memory limit (KB)
	if limit.Memory > 0 {
		if err := os.WriteFile(cgroupPath+"/memory.limit_in_bytes", []byte(strconv.Itoa(limit.Memory*1024)), 0644); err != nil {
			return fmt.Errorf("failed to set memory limit: %v", err)
		}
	}

	// outputsize limit (B)
	if limit.OutputSize > 0 {
		rl.Cur = uint64(limit.OutputSize)
		rl.Max = rl.Cur
		if err := unix.Setrlimit(unix.RLIMIT_FSIZE, &rl); err != nil {
			return fmt.Errorf("failed to set file size limit: %v", err)
		}
	}

	// stack limit (KB)
	if limit.Stack != 0 {
		rl.Cur = uint64(limit.Stack * 1024)
		rl.Max = rl.Cur
		if err := unix.Setrlimit(unix.RLIMIT_STACK, &rl); err != nil {
			return fmt.Errorf("failed to set stack limit: %v", err)
		}
	}

	return nil
}

func addToCgroup(cgroupPath string, pid int) error {
	return os.WriteFile(cgroupPath+"/cgroup.procs", []byte(strconv.Itoa(pid)), 0644)
}

func enterNamespace() error {
	// create a new PID Namespace
	if err := syscall.Unshare(syscall.CLONE_NEWPID); err != nil {
		return fmt.Errorf("failed to unshare PID namespace: %v", err)
	}

	// create a new Network Namespace
	if err := syscall.Unshare(syscall.CLONE_NEWNET); err != nil {
		return fmt.Errorf("failed to unshare Network namespace: %v", err)
	}

	// create a new Mount Namespace
	if err := syscall.Unshare(syscall.CLONE_NEWNS); err != nil {
		return fmt.Errorf("failed to unshare Mount namespace: %v", err)
	}

	return nil
}

func setProUser(uid, gid int) error {
	// set user id
	if uid != 0 {
		if err := unix.Setuid(uid); err != nil {
			return err
		}
	}

	// set user group id
	if gid != 0 {
		if err := unix.Setgid(gid); err != nil {
			return err
		}
	}

	return nil
}

func limitSysCall() error {
	sec, err := seccomp.NewFilter(seccomp.ActAllow)
	if err != nil {
		return err
	}

	err = sec.AddArch(seccomp.ArchNative)
	if err != nil {
		return err
	}

	// read write mmap fork execve ptrace openat
	syscallsBlock := []int32{0, 1, 9, 41, 56, 57, 59, 101, 257}

	for _, syscall := range syscallsBlock {
		if err := sec.AddRule(seccomp.ScmpSyscall(syscall), seccomp.ActKillThread); err != nil {
			return err
		}
	}

	if err := sec.Load(); err != nil {
		return fmt.Errorf("failed to load seccomp filter: %v", err)
	}

	return nil
}

func limitAndIsolate(cgroupPath string, limit *Limit, uid int, gid int) error {
	// create a Cgroup to limit processes
	if err := createCgroup(cgroupPath, limit); err != nil {
		return fmt.Errorf("failed to create cgroup: %v", err)
	}

	// get current PID
	pid := os.Getpid()

	// add current proc in Cgroup
	if err := addToCgroup(cgroupPath, pid); err != nil {
		return fmt.Errorf("failed to add process to cgroup: %v", err)
	}

	// use Namespace isolate processes
	if err := enterNamespace(); err != nil {
		return fmt.Errorf("failed to enter namespace: %v", err)
	}

	// set user and group
	if err := setProUser(uid, gid); err != nil {
		return fmt.Errorf("failed to set process user: %v", err)
	}

	return nil
}
