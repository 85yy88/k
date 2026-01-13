#!/data/data/com.termux/files/usr/bin/bash
# 一键推送代码到 GitHub（需要先创建仓库）

set -e

echo "=== 一键推送代码到 GitHub ==="
echo ""

# 检查参数
if [ -z "$1" ]; then
    echo "使用方法: bash 一键推送GitHub.sh https://github.com/YOUR_USERNAME/REPO_NAME.git"
    echo ""
    echo "或者分步执行："
    echo "  1. bash 一键推送GitHub.sh"
    echo "  2. 输入仓库地址"
    exit 1
fi

REPO_URL="$1"

# 检查 Git
if ! command -v git &> /dev/null; then
    echo "正在安装 Git..."
    pkg install -y git
fi

# 配置 Git（如果还没配置）
if [ -z "$(git config --global user.name)" ]; then
    read -p "请输入你的名字: " USER_NAME
    read -p "请输入你的邮箱: " USER_EMAIL
    git config --global user.name "$USER_NAME"
    git config --global user.email "$USER_EMAIL"
fi

# 初始化仓库
if [ ! -d ".git" ]; then
    git init
fi

# 添加并提交
git add .
git commit -m "Initial commit: Termux脚本 with all features" 2>/dev/null || echo "没有新更改"

# 添加远程仓库
git remote remove origin 2>/dev/null || true
git remote add origin "$REPO_URL"

# 推送
echo ""
echo "正在推送到 GitHub..."
git branch -M main
git push -u origin main

echo ""
echo "✓✓✓ 推送成功！✓✓✓"
echo ""
echo "GitHub Actions 会自动开始编译"
echo "在仓库的 Actions 标签页查看编译进度"
echo ""
