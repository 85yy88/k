#!/data/data/com.termux/files/usr/bin/bash
# 在 Termux 中推送代码到 GitHub 的脚本

set -e

echo "=== Termux 推送代码到 GitHub ==="
echo ""

# 检查 Git 是否安装
if ! command -v git &> /dev/null; then
    echo "正在安装 Git..."
    pkg update -y
    pkg install -y git
    echo "Git 安装完成！"
fi

echo "Git 版本: $(git --version)"
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

# 配置 Git（如果还没配置）
if [ -z "$(git config --global user.name)" ]; then
    echo "配置 Git 用户信息..."
    read -p "请输入你的名字: " USER_NAME
    read -p "请输入你的邮箱: " USER_EMAIL
    git config --global user.name "$USER_NAME"
    git config --global user.email "$USER_EMAIL"
    echo "Git 配置完成！"
    echo ""
fi

# 初始化仓库（如果还没有）
if [ ! -d ".git" ]; then
    echo "初始化 Git 仓库..."
    git init
    echo ""
fi

# 添加所有文件
echo "添加文件到 Git..."
git add .
echo ""

# 提交
echo "提交更改..."
git commit -m "Initial commit: Termux脚本 with all features" 2>/dev/null || {
    echo "提示: 可能没有新更改"
    echo ""
}

echo ""
echo "========================================"
echo "准备完成！"
echo "========================================"
echo ""
echo "下一步操作："
echo ""
echo "1. 在 GitHub 上创建新仓库："
echo "   https://github.com/new"
echo "   ⚠️  不要勾选任何初始化选项！"
echo ""
echo "2. 创建仓库后，执行以下命令（替换 YOUR_USERNAME 和 REPO_NAME）："
echo ""
echo "   git remote add origin https://github.com/YOUR_USERNAME/REPO_NAME.git"
echo "   git branch -M main"
echo "   git push -u origin main"
echo ""
echo "3. 推送后，GitHub Actions 会自动编译"
echo "   在仓库的 Actions 标签页查看编译进度"
echo ""
