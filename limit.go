package judge

import (
	"golang.org/x/sys/unix"
)

type Limit struct {
	CpuTime    int
	Memory     int
	Stack      int
	OutputSize int
}

// SetProcLimit set process resource limit
func SetProcLimit(limit *Limit) error {
	var rl unix.Rlimit

	// cpu time limit (MS)
	if limit.CpuTime != 0 {
		rl.Cur = uint64(limit.CpuTime/1000 + 1)
		rl.Max = uint64(limit.CpuTime)
		if err := unix.Setrlimit(unix.RLIMIT_CPU, &rl); err != nil {
			return err
		}
	}

	// memory limit (KB)
	if limit.Memory != 0 {
		rl.Cur = uint64(limit.Memory * 1024)
		rl.Max = rl.Cur * 2
		if err := unix.Setrlimit(unix.RLIMIT_DATA, &rl); err != nil {
			return err
		}

		rl.Cur = rl.Cur * 2
		rl.Max = rl.Cur
		if err := unix.Setrlimit(unix.RLIMIT_AS, &rl); err != nil {
			return err
		}
	}

	// stack limit (KB)
	if limit.Stack != 0 {
		rl.Cur = uint64(limit.Stack * 1024)
		rl.Max = rl.Cur
		if err := unix.Setrlimit(unix.RLIMIT_STACK, &rl); err != nil {
			return err
		}
	}

	// outputsize limit (B)
	if limit.OutputSize != 0 {
		rl.Cur = uint64(limit.OutputSize)
		rl.Max = rl.Cur
		if err := unix.Setrlimit(unix.RLIMIT_FSIZE, &rl); err != nil {
			return err
		}
	}

	return nil
}

func SetProUser(uid, gid int) error {
	if uid != 0 {
		if err := unix.Setuid(uid); err != nil {
			return err
		}
	}

	if gid != 0 {
		if err := unix.Setgid(gid); err != nil {
			return err
		}
	}

	return nil
}

