package judge

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/crazyfrankie/judge-go/constant"
)

// StdCheck 检查用户输出和标准输出是否一致
func StdCheck(userOutputPath string) (int, error) {
	// 打开用户输出文件
	userFile, err := os.Open(userOutputPath)
	if err != nil {
		return constant.Fail, fmt.Errorf("failed to open user output file: %w", err)
	}
	defer userFile.Close()

	// 创建读取器
	reader := bufio.NewReader(userFile)

	// 逐行读取并比较
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return constant.Fail, fmt.Errorf("error reading file: %w", err)
		}

		// 去除行尾空白字符
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 分割用户输出和标准输出
		parts := strings.Split(line, " ")
		if len(parts) != 2 {
			return constant.Fail, &constant.ContentErr{
				Msg: fmt.Sprintf("invalid line format: %s", line),
			}
		}

		// 比较用户输出和标准输出
		if parts[0] != parts[1] {
			return constant.Fail, &constant.ContentErr{
				Msg: fmt.Sprintf("output mismatch: got %s, want %s", parts[0], parts[1]),
			}
		}
	}

	return constant.Success, nil
}
