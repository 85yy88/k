# GitHub Actions 自动编译说明

## 🎯 功能

上传代码到 GitHub 后，GitHub Actions 会自动编译生成 `fatalder-termux` 文件。

## 📋 使用方法

### 1. 上传代码到 GitHub

使用以下任一方法：
- GitHub Mobile App（推荐）
- GitHub 网页版
- Git 命令行

### 2. 自动编译

代码上传后，GitHub Actions 会自动：
- 检测代码推送
- 开始编译流程
- 编译 ARM64 版本（Termux 使用）
- 生成可执行文件

### 3. 下载编译好的文件

1. 在仓库页面点击 **"Actions"** 标签页
2. 点击最新的 workflow run（绿色 ✓ 表示成功）
3. 滚动到底部，在 **"Artifacts"** 部分
4. 点击 **"fatalder-termux"** 下载

## ⏱️ 编译时间

通常需要 **2-5 分钟**，取决于：
- 依赖下载速度
- GitHub Actions 服务器负载

## 📱 在 Termux 中使用

下载后：
```bash
# 1. 传输到手机
# 2. 在 Termux 中：
chmod +x fatalder-termux
./fatalder-termux
```

## 🔄 手动触发编译

如果需要手动触发编译：
1. 在仓库页面点击 "Actions"
2. 选择 "Build Termux Binary"
3. 点击 "Run workflow"
4. 点击 "Run workflow" 按钮

## ✅ 编译状态

- ✅ 绿色：编译成功
- ❌ 红色：编译失败（查看日志）
- 🟡 黄色：编译中

## 📝 注意事项

1. 首次编译可能需要更长时间（下载依赖）
2. 编译产物保留 30 天
3. 确保 `Termux脚本` 目录结构正确
4. 确保 `go.mod` 和 `main.go` 文件存在
