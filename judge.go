package judge

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/crazyfrankie/judge-go/constant"
	"golang.org/x/sys/unix"
)

func BaseRun(codeFilename, inputFileName string) error {
	var errs []string

	// Compile user's code
	cmd := exec.Command("go", "build", codeFilename)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("compiling code error: %v", err)
	}

	// Delete user's executables
	defer func() {
		er := os.Remove("main")
		if er != nil {
			errs = append(errs, er.Error())
		}
	}()

	// Open file ready to write user code output results
	inputFile, er := os.OpenFile(inputFileName, os.O_RDWR, 0644)
	if er != nil {
		return fmt.Errorf("openning inputfile error: %v", err)
	}
	defer func() {
		closeErr := inputFile.Close()
		if closeErr != nil {
			errs = append(errs, fmt.Sprintf("failed to close input file: %v", closeErr))
		}
	}()

	runCmd := exec.Command("./main")
	// Redirects the program's output to a file
	runCmd.Stdout = inputFile
	// Execute user code executables
	err = runCmd.Run()
	if err != nil {
		return fmt.Errorf("running user code error: %v", err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("cleaning up error: %v", errs)
	}

	return nil
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

func SetProcStream(inputFilePath, outputFilePath, errorFilePath string) error {
	// set input stream
	if inputFilePath != "" {
		// open input file
		if fd, err := unix.Open(inputFilePath, unix.O_RDONLY, 0666); err != nil {
			return err
		} else {
			// dup pipe
			if err := unix.Dup2(fd, unix.Stdin); err != nil {
				return err
			}
		}
	}

	// set output stream
	if outputFilePath != "" {
		// open file
		if fd, err := unix.Open(outputFilePath, unix.O_WRONLY|unix.O_CREAT, 0666); err != nil {
			return err
		} else {
			// dup pipe
			if err := unix.Dup2(fd, unix.Stdout); err != nil {
				return err
			}
		}
	}

	// set error stream
	if errorFilePath != "" {
		if fd, err := unix.Open(errorFilePath, unix.O_WRONLY|unix.O_CREAT, 0666); err != nil {
			return err
		} else {
			if err := unix.Dup2(fd, unix.Stderr); err != nil {
				return err
			}
		}
	}

	return nil
}
