.PHONY: all build test clean

# 默认目标
all: build

# 构建项目
build:
	go build -o bin/judge-go ./example

# 运行测试
test:
	go test -v ./...

# 清理构建文件
clean:
	rm -rf bin/
	go clean

# 安装依赖
deps:
	go mod download
	go mod tidy

# 运行示例
run: build
	./bin/judge-go 