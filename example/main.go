package main

import (
	"context"
	"fmt"
	"log"
	"time"

	judge "github.com/crazyfrankie/judge-go"
)

func main() {
	// 创建评测配置
	config := &judge.Config{
		Limits: struct {
			CPU    time.Duration
			Memory int64
			Stack  int64
			Output int64
		}{
			CPU:    2 * time.Second,
			Memory: 128 * 1024 * 1024,
			Stack:  8 * 1024 * 1024,
			Output: 10 * 1024 * 1024,
		},
		Exec: struct {
			Path string
			Args []string
			Env  []string
		}{
			Path: "/usr/bin/python3",
			Args: []string{"test.py"},
			Env:  []string{"PATH=/usr/local/bin:/usr/bin:/bin"},
		},
		Files: struct {
			UserOutput string
			CgroupPath string
		}{
			UserOutput: "user_output.txt",
			CgroupPath: "cgroup",
		},
	}

	// 创建评测实例
	j := judge.NewJudge(config)

	// 运行评测
	result, err := j.Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// 检查结果
	_, err = j.Check()
	if err != nil {
		log.Fatal(err)
	}

	// 输出结果
	fmt.Printf("评测状态: %s\n", result.GetStatus())
	fmt.Printf("CPU时间: %dms\n", result.CpuTimeUsed)
	fmt.Printf("实际时间: %dms\n", result.RealTimeUsed)
	fmt.Printf("内存使用: %d bytes\n", result.MemoryUsed)
}
