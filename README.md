# Judge-Go

Judge-Go 是一个用 Go 语言实现的在线评测系统，用于运行和评测用户提交的代码。

## 特性

- 支持多种编程语言（通过配置执行器）
- 资源限制（CPU时间、内存、栈大小、输出大小）
- 安全隔离（cgroup、namespace、seccomp）
- 精确的结果比对
- 详细的评测报告

## 安装

```bash
go get github.com/crazyfrankie/judge-go
```

## 使用方法

### 基本使用

```go
package main

import (
    "context"
    "github.com/crazyfrankie/judge-go/judge"
)

func main() {
    // 创建评测配置
    config := &judge.Config{
        Limits: struct {
            CPU     time.Duration
            Memory  int64
            Stack   int64
            Output  int64
        }{
            CPU:     2 * time.Second,
            Memory:  128 * 1024 * 1024,
            Stack:   8 * 1024 * 1024,
            Output:  10 * 1024 * 1024,
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
            UserOutput   string
            StdOutput    string
            CgroupPath   string
        }{
            UserOutput: "user_output.txt",
            StdOutput:  "test_output.txt",
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
    status, err := j.Check()
    if err != nil {
        log.Fatal(err)
    }

    // 输出结果
    fmt.Printf("评测状态: %s\n", result.GetStatus())
    fmt.Printf("CPU时间: %dms\n", result.CpuTimeUsed)
    fmt.Printf("实际时间: %dms\n", result.RealTimeUsed)
    fmt.Printf("内存使用: %d bytes\n", result.MemoryUsed)
}
```

### 配置说明

#### 资源限制

- `CPU`: CPU时间限制
- `Memory`: 内存限制
- `Stack`: 栈大小限制
- `Output`: 输出大小限制

#### 执行配置

- `Path`: 执行器路径
- `Args`: 执行器参数
- `Env`: 环境变量

#### 安全配置

- `UID`: 用户ID
- `GID`: 组ID
- `Chroot`: 根目录
- `Syscalls`: 允许的系统调用列表

#### 文件配置

- `UserOutput`: 用户输出文件路径
- `StdOutput`: 标准输出文件路径
- `CgroupPath`: cgroup路径

## 注意事项

1. 需要root权限来设置cgroup和namespace
2. 建议在容器环境中运行
3. 确保系统支持cgroup v2
4. 注意配置适当的系统调用限制

## 贡献

欢迎提交Issue和Pull Request！

## 许可证

MIT License 