# GitHub Actions 配置文件说明

## ⚠️ 重要：文件位置要求

**GitHub Actions 的 workflow 文件必须放在以下位置：**

```
仓库根目录/
└── .github/
    └── workflows/
        └── build.yml  ← 必须在这里！
```

## 📁 文件结构说明

### 如果只上传 `Termux脚本` 文件夹

那么文件结构应该是：
```
Termux脚本/
├── .github/
│   └── workflows/
│       └── build.yml  ← 必须在这里
├── main.go
├── go.mod
└── ...其他文件
```

### 如果上传整个项目

那么文件结构应该是：
```
fatalder/
├── .github/
│   └── workflows/
│       └── build.yml  ← 在根目录
├── Termux脚本/
│   ├── main.go
│   └── ...
└── ...其他文件夹
```

## ✅ 检查方法

在 GitHub 仓库页面：

1. **点击仓库中的文件列表**
2. **查看是否有 `.github` 文件夹**
3. **点击 `.github` → `workflows`**
4. **应该能看到 `build.yml` 文件**

如果没有看到，说明文件位置不对！

## 🔧 如何创建正确的文件结构

### 方法一：在 GitHub 网页上创建

1. 在仓库页面点击 **"Add file"** → **"Create new file"**
2. 文件名输入：`.github/workflows/build.yml`
   - **注意**：输入 `.github` 时，GitHub 会自动创建文件夹
3. 粘贴 workflow 内容
4. 点击 **"Commit new file"**

### 方法二：确保本地文件结构正确

确保你的文件结构是：
```
Termux脚本/
├── .github/
│   └── workflows/
│       └── build.yml
├── main.go
└── ...
```

## 📝 Workflow 文件内容

确保 `build.yml` 文件内容正确：

```yaml
name: Build Termux Binary

on:
  workflow_dispatch:  # 手动触发
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

## ✅ 验证 Workflow 是否正确

1. **在仓库页面点击 "Actions" 标签页**
2. **应该能看到 "Build Termux Binary" workflow**
3. **如果看不到，说明文件位置不对**

## 🚨 常见错误

### ❌ 错误：文件放在错误位置

```
Termux脚本/
├── build.yml  ← 错误！GitHub 不会识别
└── ...
```

### ✅ 正确：文件放在正确位置

```
Termux脚本/
├── .github/
│   └── workflows/
│       └── build.yml  ← 正确！
└── ...
```

## 💡 总结

**必须**：`.github/workflows/build.yml`  
**不能**：放在其他位置

GitHub 只会识别 `.github/workflows/` 目录下的 `.yml` 或 `.yaml` 文件作为 workflow 配置。
