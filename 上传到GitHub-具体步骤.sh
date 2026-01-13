#!/data/data/com.termux/files/usr/bin/bash
# 上传 Termux脚本 到 GitHub 的具体步骤
# 用户名: 85yy88
# 仓库名: k
# 路径: /storage/emulated/0/Download/Termux脚本

set -e

echo "========================================"
echo "  上传 Termux脚本 到 GitHub"
echo "========================================"
echo ""

# 配置信息
GITHUB_USER="85yy88"
REPO_NAME="k"
REPO_URL="https://github.com/${GITHUB_USER}/${REPO_NAME}.git"
TERMUX_SCRIPT_PATH="/storage/emulated/0/Download/Termux脚本"

echo "仓库地址: $REPO_URL"
echo "本地路径: $TERMUX_SCRIPT_PATH"
echo ""

# 检查路径是否存在
if [ ! -d "$TERMUX_SCRIPT_PATH" ]; then
    echo "错误: 路径不存在: $TERMUX_SCRIPT_PATH"
    echo ""
    echo "请检查路径是否正确，或者使用以下命令查看："
    echo "  ls /storage/emulated/0/Download/"
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

echo "Git 版本: $(git --version)"
echo ""

# 进入项目目录
cd "$TERMUX_SCRIPT_PATH"
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

# 配置 Git
echo "配置 Git..."
git config --global user.name "85yy88"
git config --global user.email "2694736714@qq.com"
echo "Git 配置完成！"
echo ""

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
    echo "提示: 可能没有新更改，或已经提交过了"
    echo ""
}

# 检查远程仓库
if git remote | grep -q "^origin$"; then
    echo "远程仓库已存在，更新地址..."
    git remote set-url origin "$REPO_URL"
else
    echo "添加远程仓库..."
    git remote add origin "$REPO_URL"
fi

echo ""
echo "========================================"
echo "  准备完成！"
echo "========================================"
echo ""
echo "下一步："
echo ""
echo "1. 确保 GitHub 仓库已创建："
echo "   https://github.com/$GITHUB_USER/$REPO_NAME"
echo ""
echo "2. 推送代码（执行以下命令）："
echo ""
echo "   git branch -M main"
echo "   git push -u origin main"
echo ""
echo "或者直接执行："
echo "   bash 一键推送.sh"
echo ""
