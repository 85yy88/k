# 临时编译脚本
$ErrorActionPreference = "Stop"

# 设置工作目录为脚本所在目录
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $scriptDir

Write-Host "当前工作目录: $(Get-Location)" -ForegroundColor Cyan
Write-Host "检查文件..." -ForegroundColor Yellow

if (-not (Test-Path "go.mod")) {
    Write-Host "错误: 未找到 go.mod 文件" -ForegroundColor Red
    exit 1
}

if (-not (Test-Path "main.go")) {
    Write-Host "错误: 未找到 main.go 文件" -ForegroundColor Red
    exit 1
}

Write-Host "设置编译环境..." -ForegroundColor Yellow
$env:GOOS = "android"
$env:GOARCH = "arm64"
$env:CGO_ENABLED = "0"

Write-Host "下载依赖..." -ForegroundColor Yellow
go mod tidy
if ($LASTEXITCODE -ne 0) {
    Write-Host "依赖下载失败" -ForegroundColor Red
    exit 1
}

Write-Host "开始编译..." -ForegroundColor Yellow
go build -ldflags="-s -w" -o fatalder-termux main.go
if ($LASTEXITCODE -ne 0) {
    Write-Host "编译失败" -ForegroundColor Red
    exit 1
}

if (Test-Path "fatalder-termux") {
    $f = Get-Item "fatalder-termux"
    Write-Host ""
    Write-Host "✓✓✓ 编译成功！✓✓✓" -ForegroundColor Green
    Write-Host ""
    Write-Host "文件路径: $($f.FullName)" -ForegroundColor Cyan
    Write-Host "文件大小: $([math]::Round($f.Length/1MB,2)) MB" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "下一步操作:" -ForegroundColor Yellow
    Write-Host "  1. 将 fatalder-termux 文件传输到 Android 设备"
    Write-Host "  2. 在 Termux 中执行: chmod +x fatalder-termux"
    Write-Host "  3. 运行: ./fatalder-termux"
    Write-Host ""
} else {
    Write-Host "编译失败：未找到输出文件" -ForegroundColor Red
    exit 1
}
