#!/bin/bash
# 修复 Git 推送问题

cd /storage/emulated/0/Download/Termux脚本

echo "=== 修复 Git 推送问题 ==="
echo ""

# 1. 检查 Git 状态
echo "1. 检查 Git 状态..."
git status

echo ""
echo "2. 处理子模块修改..."

# 2. 检查子模块状态
if [ -d "modules/WaterStructure" ]; then
    cd modules/WaterStructure
    if [ -d ".git" ]; then
        echo "   WaterStructure 是一个 Git 子模块"
        echo "   检查子模块状态..."
        git status
        echo ""
        echo "   添加子模块的所有更改..."
        git add .
        echo "   提交子模块更改..."
        git commit -m "Update WaterStructure module" || echo "   子模块可能没有新更改"
    fi
    cd ../..
fi

echo ""
echo "3. 添加所有文件..."
git add .

echo ""
echo "4. 提交更改..."
git commit -m "Update: Fix module paths and configurations" || echo "   可能没有新更改需要提交"

echo ""
echo "5. 拉取远程更改（使用 rebase）..."
git pull --rebase origin main || {
    echo "   拉取失败，尝试强制同步..."
    echo "   警告：这将覆盖本地更改"
    read -p "   是否继续？(y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        git fetch origin
        git reset --hard origin/main
    else
        echo "   已取消"
        exit 1
    fi
}

echo ""
echo "6. 推送到 GitHub..."
git push origin main

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ 推送成功！"
else
    echo ""
    echo "❌ 推送失败"
    echo "   请检查："
    echo "   1. SSH 密钥是否配置正确"
    echo "   2. 仓库权限是否正确"
    echo "   3. 网络连接是否正常"
fi
