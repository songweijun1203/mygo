# mygo

一个开箱即用的 Go 命令行工程。

## 环境要求

- Go 1.27.1 或更高版本

## 快速开始

```bash
go run ./cmd/mygo
go run ./cmd/mygo -name Gopher
```

## 常用命令

```bash
make build  # 构建到 bin/mygo
make run    # 运行程序
make test   # 运行测试
make fmt    # 格式化代码
make clean  # 清理构建产物
```

## 目录结构

```text
.
├── cmd/mygo/          # 程序入口
├── internal/greeting/ # 内部业务包
├── go.mod
└── Makefile
```
