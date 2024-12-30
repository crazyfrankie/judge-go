package judge

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
)

const (
	Fail = iota
	Success
)

func BaseRun(codeFilename, inputFileName string) error {
	cmd := exec.Command("go", "build", codeFilename)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error compiling code: %v", err)
	}
	defer os.Remove("main")

	inputFile, err := os.Open(inputFileName)
	if err != nil {
		return fmt.Errorf("error openning inputfile: %v", err)
	}
	defer inputFile.Close()

	runCmd := exec.Command("./main")
	runCmd.Stdout = inputFile

	err = runCmd.Run()
	if err != nil {
		return fmt.Errorf("error running user code: %v", err)
	}

	return nil
}

func StdCheck(userInputName, stdInputName string) int {
	answerFile, err := os.Open(userInputName)
	if err != nil {
		fmt.Println("Error opening answer file:", err)
		return Fail
	}
	defer answerFile.Close()

	stdFile, err := os.Open(stdInputName)
	if err != nil {
		fmt.Println("Error opening answer file:", err)
		return Fail
	}
	defer stdFile.Close()

	answerReader := bufio.NewReader(answerFile)
	stdReader := bufio.NewReader(stdFile)

	for {
		ans, err1 := answerReader.ReadByte()
		out, err2 := stdReader.ReadByte()

		// 检查是否有任何文件结束
		if err1 != nil || err2 != nil {
			if err1 != err2 || ans != out { // 如果两个文件有不同的结束位置或内容不同
				return Fail
			}
			break
		}

		// 如果字符不同
		if ans != out {
			return Fail
		}
	}

	return Success
}
