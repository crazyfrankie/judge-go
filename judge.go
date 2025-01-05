package judge

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/crazyfrankie/judge-go/constant"
	"golang.org/x/sys/unix"
)

func BaseRun(limit *Limit, cgroupPath, userOutputPath, execFilePath string, execArgs, envs []string, syscallRule []bool, uid, gid int, chroot string) (Result, error) {
	// Lock the current thread to ensure that the current thread is not migrated during execution
	runtime.LockOSThread()
	// Unlock at the end of the function
	defer runtime.UnlockOSThread()

	var result Result
	var status unix.WaitStatus
	var ru unix.Rusage

	// start time
	startTime := time.Now().UnixMilli()

	// Ensure that only one process can create child processes
	var ForkLock sync.RWMutex
	ForkLock.Lock()

	// Create child process
	pid, _, _ := unix.Syscall(unix.SYS_FORK, 0, 0, 0)

	if pid < 0 {
		// Fork error
		return result, errors.New("fork failure")
	} else {
		// Child process
		var errs []string
		// Set limitation and isolation
		if err := limitAndIsolate(cgroupPath, limit); err != nil {
			panic(err)
		}
		defer func() {
			// Delete Cgroup
			if err := os.RemoveAll(cgroupPath); err != nil {
				errs = append(errs, fmt.Sprintf("failed to clean up cgroup: %v", err))
			}
		}()

		// if err := setProcUser(uid, gid); err != nil {
		// 	panic(err)
		// }

		// // Change the root directory of the process
		// if chroot != "" {
		// 	if err := unix.Chroot(chroot); err != nil {
		// 		panic(err)
		// 	}
		// }

		// Open user's output file
		useroutFile, er := os.OpenFile(userOutputPath, os.O_RDWR, 0644)
		if er != nil {
			return result, fmt.Errorf("openning inputfile error: %v", er)
		}
		defer func() {
			closeErr := useroutFile.Close()
			if closeErr != nil {
				errs = append(errs, fmt.Sprintf("failed to close input file: %v", closeErr))
			}
		}()

		// Apply seccomp for syscall restrictions
		if err := applySeccomp(syscallRule); err != nil {
			panic(err)
		}

		// Exec run
		cmd := exec.Command(execArgs[0], append(execArgs[1:], execFilePath)...)
		// set args
		cmd.Env = envs
		cmd.Stdout = useroutFile
		err := cmd.Run()
		if err != nil {
			return result, err
		}

		// You will never arrive here
		unix.Exit(-1)
	}
	// Release the parent process's ForkLock lock
	ForkLock.Unlock()

	// Parent

	// Set real time limit
	stop := make(chan struct{})
	// If a time limit is set, a timer is created
	// Determine if the child process exits after a period of time
	if limit.RealTime != 0 {
		ticker := time.NewTicker(time.Millisecond * time.Duration(limit.RealTime))
		go func() {
			defer ticker.Stop()

			select {
			case <-ticker.C:
				// The child process timed out
				// If the child process is still running, send the SIGKILL signal to terminate it
				ret, _ := unix.Wait4(int(pid), &status, unix.WNOHANG, &ru)
				if ret == 0 {
					_ = unix.Kill(int(pid), unix.SIGKILL)
				}
			case <-stop:
				return
			}
		}()
	}

	// Open the /proc/[pid]/status file of the child process for memory usage
	fd, err := unix.Open("/proc/"+strconv.Itoa(int(pid))+"/status", unix.O_RDONLY, 600)
	if err != nil {
		return result, err
	}
	defer unix.Close(fd)

	var regs unix.PtraceRegs
	for {
		// Wait for the child process to pause and obtain the status
		if _, err := unix.Wait4(int(pid), &status, unix.WSTOPPED, &ru); err != nil {
			return result, err
		}

		// If the child process exits, the loop exits
		if status.Exited() {
			break
		}

		// If the child process pause signal is not SIGTRAP, the pause is not caused by ptrace
		if status.StopSignal() != unix.SIGTRAP {
			_, _, _ = unix.Syscall(unix.SYS_PTRACE, uintptr(unix.PTRACE_KILL), 0, 0)
			_, _ = unix.Wait4(int(pid), nil, 0, nil)
			result.ReFlag = true
			break
		}

		// Gets register information for the child process
		if err := unix.PtraceGetRegs(int(pid), &regs); err != nil {
			return result, err
		}

		// Gets the memory usage of the child process
		if ms, err := MemoryUsage(fd); err != nil {
			return result, err
		} else if ms.VMData > result.MemoryUsed {
			result.MemoryUsed = ms.VMData
		}

		// Let the child continue executing and wait for the next system call
		if err := unix.PtraceSyscall(int(pid), 0); err != nil {
			return result, err
		}
	}

	// Stop the timer to end the time limit check
	stop <- struct{}{}

	// cpu time used
	// user time + system time
	result.CpuTimeUsed = int(ru.Utime.Sec*1000) + int(ru.Utime.Usec/1000) +
		int(ru.Stime.Sec*1000) + int(ru.Stime.Usec/1000)

	// real time used
	endTime := time.Now().UnixMilli()
	result.RealTimeUsed = int(endTime - startTime)

	// Records the signal received by the process
	result.Signal = int(status.StopSignal())

	return result, nil
}

func StdCheck(userOutputPath, stdOutputPath string) (int, error) {
	var errs []string

	answerFile, err := os.Open(userOutputPath)
	if err != nil {
		return constant.Fail, fmt.Errorf("opening answer file error: %v", err)
	}
	defer func() {
		err := answerFile.Close()
		if err != nil {
			errs = append(errs, err.Error())
		}
	}()

	stdFile, err := os.Open(stdOutputPath)
	if err != nil {
		return constant.Fail, fmt.Errorf("opening answer file error: %v", err)
	}
	defer func() {
		err := stdFile.Close()
		if err != nil {
			errs = append(errs, err.Error())
		}
	}()

	answerReader := bufio.NewReader(answerFile)
	stdReader := bufio.NewReader(stdFile)

	for {
		ans, err1 := answerReader.ReadByte()
		out, err2 := stdReader.ReadByte()

		// If both files arrive at EOF and the contents are the same, the loop redirects
		if err1 == io.EOF && err2 == io.EOF {
			break
		}

		// If one file reaches the EOF and another file still has content, an inconsistent content error is returned
		if err1 == io.EOF && err2 != io.EOF {
			return constant.Fail, &constant.ContentErr{Msg: fmt.Sprintf("one file ended before the other: err1: %v,err2: %v", err1, err2)}
		}
		if err2 == io.EOF && err1 != io.EOF {
			return constant.Fail, &constant.ContentErr{Msg: fmt.Sprintf("one file ended before the other: err1: %v,err2: %v", err1, err2)}
		}

		// If other read errors occur, return the system error directly
		if err1 != nil || err2 != nil {
			return constant.Fail, &constant.SystemErr{Msg: fmt.Sprintf("error reading files: err1: %v, err2: %v", err1, err2)}
		}

		// If the bytes are different, the contents are inconsistent, and an error is returned
		if ans != out {
			return constant.Fail, &constant.ContentErr{Msg: "content mismatch"}
		}
	}

	if len(errs) > 0 {
		return constant.Fail, &constant.SystemErr{Msg: "close files error"}
	}

	return constant.Success, nil
}

func AllowSyscall(syscallRule []bool, syscallId uint64) bool {
	return syscallRule[int(syscallId)]
}
