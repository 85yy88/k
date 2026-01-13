#!/bin/bash
# 使用 Docker 编译脚本

echo "=== 使用 Docker 编译 Termux 二进制文件 ==="

# 检查 Docker 是否安装
if ! command -v docker &> /dev/null; then
    echo "错误: 未找到 Docker，请先安装 Docker"
    exit 1
fi

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "当前目录: $SCRIPT_DIR"
echo "开始编译..."

docker run --rm \
  -v "$SCRIPT_DIR:/workspace" \
  -w /workspace \
  -e GOOS=android \
  -e GOARCH=arm64 \
  -e CGO_ENABLED=0 \
  -e GOPROXY=https://goproxy.cn,direct \
  golang:1.25 \
  bash -c "go mod download && go build -ldflags='-s -w' -o fatalder-termux main.go"

if [ -f "$SCRIPT_DIR/fatalder-termux" ]; then
    echo ""
    echo "✓✓✓ 编译成功！✓✓✓"
    echo "文件: $SCRIPT_DIR/fatalder-termux"
    echo "文件大小: $(du -h "$SCRIPT_DIR/fatalder-termux" | cut -f1)"
else
    echo "编译失败"
    exit 1
fi
