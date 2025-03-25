package judge

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/crazyfrankie/judge-go/constant"
)

// StdCheck 检查用户输出和标准输出是否一致
func StdCheck(userOutputPath, stdOutputPath string) (int, error) {
	// 打开用户输出文件
	userFile, err := os.Open(userOutputPath)
	if err != nil {
		return constant.Fail, fmt.Errorf("failed to open user output file: %w", err)
	}
	defer userFile.Close()

	// 打开标准输出文件
	stdFile, err := os.Open(stdOutputPath)
	if err != nil {
		return constant.Fail, fmt.Errorf("failed to open standard output file: %w", err)
	}
	defer stdFile.Close()

	// 创建读取器
	userReader := bufio.NewReader(userFile)
	stdReader := bufio.NewReader(stdFile)

	// 逐行比较
	for {
		userLine, err1 := userReader.ReadBytes('\n')
		stdLine, err2 := stdReader.ReadBytes('\n')

		// 处理EOF
		if err1 == io.EOF && err2 == io.EOF {
			break
		}
		if err1 == io.EOF || err2 == io.EOF {
			return constant.Fail, &constant.ContentErr{
				Msg: "output length mismatch",
			}
		}
		if err1 != nil || err2 != nil {
			return constant.Fail, &constant.SystemErr{
				Msg: fmt.Sprintf("error reading files: %v, %v", err1, err2),
			}
		}

		// 去除行尾空白字符
		userLine = bytes.TrimSpace(userLine)
		stdLine = bytes.TrimSpace(stdLine)

		// 比较内容
		if !bytes.Equal(userLine, stdLine) {
			return constant.Fail, &constant.ContentErr{
				Msg: "content mismatch",
			}
		}
	}

	return constant.Success, nil
}
