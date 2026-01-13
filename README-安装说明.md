# Termux 安装说明

## 方法一：在 Termux 中直接编译（推荐）

### 步骤：

1. **将整个 `Termux脚本` 文件夹传输到 Android 设备**
   - 可以通过 USB、网络共享、云盘等方式

2. **在 Termux 中打开文件夹**
   ```bash
   cd ~/Termux脚本
   # 或者你存放的路径
   ```

3. **运行安装脚本**
   ```bash
   bash install-termux.sh
   ```

   脚本会自动：
   - 检查并安装 Go（如果需要）
   - 下载依赖
   - 编译生成 `fatalder-termux` 可执行文件

4. **运行程序**
   ```bash
   ./fatalder-termux
   ```

## 方法二：手动编译

如果自动脚本有问题，可以手动执行：

```bash
# 1. 安装 Go（如果还没安装）
pkg update
pkg install golang

# 2. 进入项目目录
cd ~/Termux脚本

# 3. 设置代理（如果网络有问题）
export GOPROXY="https://goproxy.cn,direct"

# 4. 下载依赖
go mod download

# 5. 编译
go build -ldflags="-s -w" -o fatalder-termux main.go

# 6. 添加执行权限
chmod +x fatalder-termux

# 7. 运行
./fatalder-termux
```

## 方法三：使用预编译文件（如果已编译）

如果你已经在 Windows 上成功编译了 `fatalder-termux` 文件：

1. 将 `fatalder-termux` 文件传输到 Android 设备
2. 在 Termux 中：
   ```bash
   chmod +x fatalder-termux
   ./fatalder-termux
   ```

## 网络问题解决

如果下载依赖时遇到网络问题，可以尝试：

```bash
# 使用国内代理
export GOPROXY="https://goproxy.cn,direct"

# 或者阿里云镜像
export GOPROXY="https://mirrors.aliyun.com/goproxy/,direct"

# 或者直接模式（不使用代理）
export GOPROXY="direct"
```

## 安装到系统 PATH（可选）

编译完成后，可以将程序安装到系统 PATH，方便在任何位置使用：

```bash
mv fatalder-termux ~/../usr/bin/fatalder
fatalder  # 现在可以在任何位置直接使用
```
