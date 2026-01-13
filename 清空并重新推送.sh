#!/data/data/com.termux/files/usr/bin/bash
# 清空 GitHub 仓库并重新推送

set -e

echo "=== 清空 GitHub 仓库并重新推送 ==="
echo ""

# 检查参数
if [ -z "$1" ]; then
    echo "使用方法: bash 清空并重新推送.sh https://github.com/YOUR_USERNAME/REPO_NAME.git"
    echo ""
    read -p "请输入仓库地址: " REPO_URL
else
    REPO_URL="$1"
fi

# 提取仓库名称
REPO_NAME=$(basename "$REPO_URL" .git)
TEMP_DIR="/data/data/com.termux/files/usr/tmp/github-$REPO_NAME-$$"

echo "仓库地址: $REPO_URL"
echo "临时目录: $TEMP_DIR"
echo ""

# 清理临时目录
rm -rf "$TEMP_DIR"
mkdir -p "$TEMP_DIR"

# 克隆仓库
echo "正在克隆仓库..."
cd "$TEMP_DIR"
git clone "$REPO_URL" repo
cd repo

# 删除所有文件（保留 .git）
echo "正在清空仓库..."
git rm -rf . 2>/dev/null || true
git commit -m "Clear repository" || {
    echo "提示: 可能仓库已经是空的"
}

# 推送清空操作
echo "正在推送清空操作..."
git push origin main || git push origin master

# 复制新文件
echo "正在复制新文件..."
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cp -r "$SCRIPT_DIR"/* . 2>/dev/null || true
cp -r "$SCRIPT_DIR"/.[^.]* . 2>/dev/null || true

# 排除 .git 目录
rm -rf .git 2>/dev/null || true

# 添加并提交
echo "正在添加文件..."
git add .
git commit -m "Initial commit: Termux脚本 with all features"

# 推送
echo "正在推送新文件..."
git push origin main || git push origin master

# 清理
cd ~
rm -rf "$TEMP_DIR"

echo ""
echo "✓✓✓ 完成！✓✓✓"
echo ""
echo "仓库已清空并重新推送"
echo "GitHub Actions 会自动开始编译"
echo ""
