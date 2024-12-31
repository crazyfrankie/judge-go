package judge

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/crazyfrankie/judge-go/constant"
)

func BaseRun(codeFilename, inputFileName string) error {
	var errs []string

	// 编译用户代码
	cmd := exec.Command("go", "build", codeFilename)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("compiling code error: %v", err)
	}

	// 删除用户代码的可执行文件
	defer func() {
		er := os.Remove("main")
		if er != nil {
			errs = append(errs, er.Error())
		}
	}()

	// 打开文件准备写入用户代码输出结果
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
	// 重定向程序的输出到文件
	runCmd.Stdout = inputFile
	// 执行用户代码可执行文件
	err = runCmd.Run()
	if err != nil {
		return fmt.Errorf("running user code error: %v", err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("cleaning up error: %v", errs)
	}

	return nil
}

func StdCheck(userInputName, stdInputName string) (int, error) {
	var errs []string

	answerFile, err := os.Open(userInputName)
	if err != nil {
		return constant.Fail, fmt.Errorf("opening answer file error: %v", err)
	}
	defer func() {
		err := answerFile.Close()
		if err != nil {
			errs = append(errs, err.Error())
		}
	}()

	stdFile, err := os.Open(stdInputName)
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

		// 如果两个文件都到达EOF，内容一致则跳出循环
		if err1 == io.EOF && err2 == io.EOF {
			break
		}

		// 如果一个文件到达EOF而另一个文件还有内容，返回内容不一致错误
		if err1 == io.EOF && err2 != io.EOF {
			return constant.Fail, &constant.ContentErr{Msg: fmt.Sprintf("one file ended before the other: err1: %v,err2: %v", err1, err2)}
		}
		if err2 == io.EOF && err1 != io.EOF {
			return constant.Fail, &constant.ContentErr{Msg: fmt.Sprintf("one file ended before the other: err1: %v,err2: %v", err1, err2)}
		}

		// 如果发生了其他读取错误，直接返回系统错误
		if err1 != nil || err2 != nil {
			return constant.Fail, &constant.SystemErr{Msg: fmt.Sprintf("error reading files: err1: %v, err2: %v", err1, err2)}
		}

		// 如果字节不同，说明内容不一致，返回错误
		if ans != out {
			return constant.Fail, &constant.ContentErr{Msg: "content mismatch"}
		}
	}

	if len(errs) > 0 {
		return constant.Fail, &constant.SystemErr{Msg: "close files error"}
	}

	return constant.Success, nil
}
