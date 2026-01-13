#!/bin/bash

# Termux 二进制编译脚本

set -e

echo "=== 编译 Termux 二进制文件 ==="

# 检查 Go 是否安装
if ! command -v go &> /dev/null; then
    echo "错误: 未找到 Go，请先安装:"
    echo "  pkg install golang"
    exit 1
fi

echo "Go 版本: $(go version)"

# 进入脚本目录
cd "$(dirname "$0")"

# 下载依赖
echo ""
echo "正在下载依赖..."
go mod download

# 设置 Go 代理（使用国内镜像，避免网络问题）
export GOPROXY="https://goproxy.cn,direct"

# 编译 ARM64 版本（Termux 默认架构）
echo ""
echo "正在编译 ARM64 版本..."
GOOS=android GOARCH=arm64 go build -ldflags="-s -w" -o fatalder-termux main.go

if [ -f "fatalder-termux" ]; then
    echo ""
    echo "✓ 编译成功！"
    echo "二进制文件: $(pwd)/fatalder-termux"
    echo ""
    echo "使用方法:"
    echo "  chmod +x fatalder-termux"
    echo "  ./fatalder-termux help"
    echo ""
    echo "或者移动到 PATH:"
    echo "  mv fatalder-termux ~/../usr/bin/fatalder"
    echo "  fatalder help"
else
    echo "错误: 编译失败"
    exit 1
fi
