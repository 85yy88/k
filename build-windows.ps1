# Windows PowerShell 交叉编译脚本 - 编译 Termux ARM64 版本

Write-Host "=== 编译 Termux ARM64 二进制文件 ===" -ForegroundColor Green

# 检查 Go 是否安装
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "错误: 未找到 Go，请先安装 Go: https://golang.org/dl/" -ForegroundColor Red
    exit 1
}

Write-Host "Go 版本: $(go version)" -ForegroundColor Cyan

# 进入脚本目录
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $scriptDir

# 下载依赖
Write-Host ""
Write-Host "正在下载依赖..." -ForegroundColor Yellow
go mod download

# 设置环境变量用于交叉编译
$env:GOOS = "android"
$env:GOARCH = "arm64"
$env:CGO_ENABLED = "0"

# 编译 ARM64 版本（Termux 默认架构）
Write-Host ""
Write-Host "正在编译 ARM64 版本..." -ForegroundColor Yellow
go build -ldflags="-s -w" -o fatalder-termux main.go

if (Test-Path "fatalder-termux") {
    Write-Host ""
    Write-Host "✓ 编译成功！" -ForegroundColor Green
    Write-Host "二进制文件: $(Resolve-Path fatalder-termux)" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "使用方法:" -ForegroundColor Yellow
    Write-Host "  1. 将 fatalder-termux 文件传输到 Android 设备"
    Write-Host "  2. 在 Termux 中执行:"
    Write-Host "     chmod +x fatalder-termux"
    Write-Host "     ./fatalder-termux"
    Write-Host ""
    Write-Host "  或者移动到 PATH:"
    Write-Host "     mv fatalder-termux ~/../usr/bin/fatalder"
    Write-Host "     fatalder"
} else {
    Write-Host "错误: 编译失败" -ForegroundColor Red
    exit 1
}
