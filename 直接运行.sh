#!/data/data/com.termux/files/usr/bin/bash
# 直接运行 Go 源码（不编译，每次运行都会编译）

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# 检查 Go 是否安装
if ! command -v go &> /dev/null; then
    echo "正在安装 Go..."
    pkg update -y
    pkg install -y golang
fi

# 设置代理
export GOPROXY="https://goproxy.cn,direct"

# 直接运行（第一次会下载依赖并编译，之后会使用缓存）
go run main.go "$@"
