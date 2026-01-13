#!/data/data/com.termux/files/usr/bin/bash

# Termux 安装和编译脚本
# 使用方法：在 Termux 中执行: bash install-termux.sh

set -e

echo "=== Termux 安装脚本 ==="
echo ""

# 检查并安装 Go
if ! command -v go &> /dev/null; then
    echo "正在安装 Go..."
    pkg update -y
    pkg install -y golang
    echo "Go 安装完成！"
else
    echo "Go 已安装: $(go version)"
fi

echo ""

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "当前目录: $(pwd)"
echo ""

# 检查必要文件
if [ ! -f "go.mod" ]; then
    echo "错误: 未找到 go.mod 文件"
    exit 1
fi

if [ ! -f "main.go" ]; then
    echo "错误: 未找到 main.go 文件"
    exit 1
fi

# 设置 Go 代理（使用国内镜像，如果网络有问题）
echo "设置 Go 代理..."
export GOPROXY="https://goproxy.cn,direct"
# 如果上面不行，可以尝试：
# export GOPROXY="https://mirrors.aliyun.com/goproxy/,direct"
# 或者：
# export GOPROXY="direct"

echo "下载依赖..."
go mod download || {
    echo "使用备用代理重试..."
    export GOPROXY="https://mirrors.aliyun.com/goproxy/,direct"
    go mod download
}

echo ""
echo "开始编译..."
go build -ldflags="-s -w" -o fatalder-termux main.go

if [ -f "fatalder-termux" ]; then
    chmod +x fatalder-termux
    echo ""
    echo "✓✓✓ 编译成功！✓✓✓"
    echo ""
    echo "文件路径: $(pwd)/fatalder-termux"
    echo "文件大小: $(du -h fatalder-termux | cut -f1)"
    echo ""
    echo "使用方法:"
    echo "  ./fatalder-termux"
    echo ""
    echo "或者移动到 PATH:"
    echo "  mv fatalder-termux ~/../usr/bin/fatalder"
    echo "  fatalder"
    echo ""
else
    echo "编译失败"
    exit 1
fi
