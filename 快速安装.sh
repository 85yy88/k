#!/bin/bash
# 快速安装脚本 - 一键在 Termux 中编译

echo "开始安装..."

# 安装 Go
pkg install -y golang 2>/dev/null || true

# 设置代理
export GOPROXY="https://goproxy.cn,direct"

# 编译
cd "$(dirname "$0")"
go mod download
go build -ldflags="-s -w" -o fatalder-termux main.go
chmod +x fatalder-termux

echo "完成！运行: ./fatalder-termux"
