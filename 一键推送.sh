#!/data/data/com.termux/files/usr/bin/bash
# 一键推送代码到 GitHub
# 用户名: 85yy88
# 仓库名: k

set -e

GITHUB_USER="85yy88"
REPO_NAME="k"
REPO_URL="git@github.com:${GITHUB_USER}/${REPO_NAME}.git"
TERMUX_SCRIPT_PATH="/storage/emulated/0/Download/Termux脚本"

echo "========================================"
echo "  一键推送代码到 GitHub"
echo "========================================"
echo ""

# 检查路径
if [ ! -d "$TERMUX_SCRIPT_PATH" ]; then
    echo "错误: 路径不存在: $TERMUX_SCRIPT_PATH"
    echo "请检查路径是否正确"
    exit 1
fi

# 检查并安装 Git
if ! command -v git &> /dev/null; then
    echo "正在安装 Git..."
    pkg update -y
    pkg install -y git
    echo "Git 安装完成！"
    echo ""
fi

# 配置 Git
echo "配置 Git..."
git config --global user.name "85yy88"
git config --global user.email "2694736714@qq.com"

# 进入项目目录
echo "进入项目目录..."
cd "$TERMUX_SCRIPT_PATH"
echo "当前目录: $(pwd)"
echo ""

# 初始化仓库
if [ ! -d ".git" ]; then
    echo "初始化 Git 仓库..."
    git init
    echo ""
fi

# 添加并提交
echo "添加文件..."
git add .
echo ""

echo "提交更改..."
git commit -m "Initial commit: Termux脚本 with all features" 2>/dev/null || {
    echo "提示: 可能没有新更改"
    echo ""
}

# 设置远程仓库
echo "设置远程仓库..."
git remote remove origin 2>/dev/null || true
git remote add origin "$REPO_URL"
echo ""

# 推送
echo "正在推送到 GitHub..."
echo ""

git branch -M main
git push -u origin main

echo ""
echo "✓✓✓ 推送成功！✓✓✓"
echo ""
echo "GitHub Actions 会自动开始编译"
echo "查看编译进度: https://github.com/$GITHUB_USER/$REPO_NAME/actions"
echo ""
