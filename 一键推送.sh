#!/data/data/com.termux/files/usr/bin/bash
# 一键推送代码到 GitHub（优化版）
# 用户名: 85yy88
# 仓库名: k
set -e
# 配置信息（替换邮箱为你的GitHub邮箱）
GITHUB_USER='85yy88'
REPO_NAME='k'
REPO_URL="git@github.com:${GITHUB_USER}/${REPO_NAME}.git"
TERMUX_SCRIPT_PATH="/storage/emulated/0/Download/Termux脚本"
echo "============================================="
echo "  一键推送代码到 GitHub（优化版）"
echo "============================================="
echo ""
# 1. 检查路径
if [ ! -d "$TERMUX_SCRIPT_PATH" ]; then
    echo "错误: 路径不存在: $TERMUX_SCRIPT_PATH"
    echo "请检查路径是否正确"
    exit 1
fi
# 2. 检查并安装 Git
if ! command -v git &> /dev/null; then
    echo "正在安装 Git.."
    pkg install git -y
fi
# 3. 进入项目目录
echo "进入项目目录.."
cd "$TERMUX_SCRIPT_PATH"
# 4. 初始化 Git 仓库（如果未初始化）
if [ ! -d ".git" ]; then
    echo "初始化 Git 仓库.."
    git init
    git remote add origin "$REPO_URL"
fi
# 5. 配置 Git 用户名和邮箱（替换为你的GitHub邮箱）
if ! git config user.name &> /dev/null; then
    echo "配置 Git 用户名和邮箱.."
    git config user.name "$GITHUB_USER"
    git config user.email "你的GitHub邮箱@example.com"
fi
# 6. 拉取远程最新代码（已修正语法错误）
echo "拉取远程最新代码.."
git pull origin main --allow-unrelated-histories || { echo "拉取远程代码失败，将继续尝试推送"; true; }
# 7. 添加所有文件到暂存区
echo "添加文件.."
git add .
# 8. 提交更改
echo "提交更改.."
COMMIT_MSG="自动提交: $(date +"%Y-%m-%d %H:%M:%S")"
git commit -m "$COMMIT_MSG" || echo "提示: 可能没有新更改"
# 9. 设置远程仓库（确保是SSH地址）
echo "设置远程仓库.."
git remote set-url origin "$REPO_URL"
# 10. 推送代码（增加成功/失败提示）
echo "正在推送到 GitHub.."
if git push origin main; then
    echo "✅ 代码推送成功！"
else
    echo "❌ 推送失败，建议检查：1. SSH密钥是否配置 2. 仓库权限是否正确"
    exit 1
fi
