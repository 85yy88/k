# 修复 GitHub Actions 自动编译

## 问题诊断

如果上传后没有自动编译，可能的原因：

### 1. 检查 Actions 是否启用

1. 在 GitHub 仓库页面
2. 点击 **"Settings"**（设置）
3. 左侧菜单找到 **"Actions"** → **"General"**
4. 确保 **"Allow all actions and reusable workflows"** 已启用
5. 点击 **"Save"**

### 2. 检查 Workflow 文件位置

Workflow 文件必须在正确的位置：
- ✅ 正确：`.github/workflows/build.yml`（仓库根目录）
- ❌ 错误：`Termux脚本/.github/workflows/build.yml`（如果只上传了 Termux脚本 文件夹，这个位置也可以）

### 3. 检查分支名称

确保代码推送到 `main` 或 `master` 分支。

### 4. 手动触发编译

如果自动触发不工作，可以手动触发：

1. 在仓库页面点击 **"Actions"** 标签页
2. 左侧选择 **"Build Termux Binary"**
3. 点击 **"Run workflow"**
4. 选择分支（main 或 master）
5. 点击 **"Run workflow"** 按钮

---

## 快速修复步骤

### 方法一：确保文件结构正确

如果只上传了 `Termux脚本` 文件夹，确保包含：
```
Termux脚本/
├── .github/
│   └── workflows/
│       └── build.yml  ← 这个文件必须存在
├── main.go
├── go.mod
└── ...其他文件
```

### 方法二：在 GitHub 网页上创建 Workflow

1. 在仓库页面点击 **"Add file"** → **"Create new file"**
2. 文件名输入：`.github/workflows/build.yml`
3. 复制以下内容：

```yaml
name: Build Termux Binary

on:
  workflow_dispatch:
  push:
    branches: [ main, master ]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25'
      
      - name: Build for Android ARM64
        env:
          GOOS: android
          GOARCH: arm64
          CGO_ENABLED: 0
        run: |
          if [ -f "main.go" ]; then
            echo "Building in current directory"
          elif [ -f "Termux脚本/main.go" ]; then
            cd Termux脚本
          else
            echo "Error: main.go not found"
            exit 1
          fi
          go mod download
          go build -ldflags="-s -w" -o fatalder-termux main.go
      
      - name: Upload artifact
        uses: actions/upload-artifact@v3
        with:
          name: fatalder-termux
          path: |
            fatalder-termux
            Termux脚本/fatalder-termux
          retention-days: 30
```

4. 点击 **"Commit new file"**

### 方法三：检查 Actions 标签页

1. 在仓库页面点击 **"Actions"** 标签页
2. 如果看到 **"Build Termux Binary"**，说明 workflow 已配置
3. 如果没有看到，说明 workflow 文件位置不对或格式错误

---

## 验证 Workflow 是否工作

1. 在仓库页面点击 **"Actions"** 标签页
2. 应该能看到 **"Build Termux Binary"** workflow
3. 点击它，然后点击 **"Run workflow"** 手动触发
4. 如果成功，会看到编译进度
5. 编译完成后，在 Artifacts 部分下载文件

---

## 如果还是不行

1. **检查仓库设置**：
   - Settings → Actions → General
   - 确保 Actions 已启用

2. **检查文件路径**：
   - 确保 `.github/workflows/build.yml` 文件存在
   - 文件内容格式正确（YAML 格式）

3. **查看错误日志**：
   - 在 Actions 标签页点击失败的 workflow
   - 查看错误信息

4. **手动触发测试**：
   - 使用 workflow_dispatch 手动触发
   - 这样可以测试 workflow 是否配置正确
