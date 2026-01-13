# Termux 二进制工具

这是一个可以在 Android Termux 中运行的 Minecraft 工具集，包含结构格式转换、地图画生成和存档加密/解密功能。

## 📋 功能

1. **结构格式转换** - 支持多种格式互转（Schematic, BDX, MCStructure 等）
2. **地图画转换** - 将图片转换为 Minecraft 地图画
3. **存档加密/解密** - 网易版世界存档加密解密

## 🚀 快速开始

### 方法 1: 使用预编译二进制（推荐）

1. **下载二进制文件**
   ```bash
   # 从 GitHub Releases 下载 fatalder-termux
   wget https://github.com/你的仓库/releases/latest/download/fatalder-termux
   chmod +x fatalder-termux
   ```

2. **移动到 PATH（可选）**
   ```bash
   mv fatalder-termux ~/../usr/bin/fatalder
   ```

3. **使用**
   ```bash
   fatalder help
   ```

### 方法 2: 自己编译

1. **安装依赖**
   ```bash
   pkg update && pkg upgrade
   pkg install golang git
   ```

2. **克隆项目**
   ```bash
   cd ~
   git clone <你的仓库地址> fatalder-termux
   cd fatalder-termux/Termux脚本
   ```

3. **编译**
   ```bash
   chmod +x build.sh
   ./build.sh
   ```

4. **使用**
   ```bash
   chmod +x fatalder-termux
   ./fatalder-termux help
   ```

## 📖 使用说明

### 结构格式转换

```bash
# 基本用法
fatalder convert <输入文件> <目标格式> [输出文件]

# 示例
fatalder convert input.schematic MCStructure output.mcstructure
fatalder convert input.bdx BDX output.bdx
fatalder c world.mcworld Litematic  # 使用短命令
```

### 地图画转换

```bash
# 基本用法
fatalder mapart <图片文件> <世界文件/目录> [选项]

# 选项:
#   --x <X坐标>       起始 X 坐标（子区块）
#   --y <Y坐标>       起始 Y 坐标（子区块，默认-4）
#   --z <Z坐标>       起始 Z 坐标（子区块）
#   --width <宽度>    地图宽度（地图数量，默认1）
#   --height <高度>   地图高度（地图数量，默认1）
#   --2d              强制2D模式（平面）

# 示例
fatalder mapart image.jpg world.mcworld
fatalder mapart image.png world.mcworld --width 2 --height 2
fatalder m photo.jpg /sdcard/games/com.mojang/minecraftWorlds/World1 --x 0 --y -4 --z 0
```

### 存档加密/解密

```bash
# 加密
fatalder encrypt <世界文件/目录>

# 解密
fatalder decrypt <世界文件/目录>

# 示例
fatalder encrypt world.mcworld
fatalder decrypt /sdcard/games/com.netease/minecraftWorlds/World1
fatalder e world.mcworld  # 使用短命令
fatalder d world.mcworld  # 使用短命令
```

### 列出支持的格式

```bash
fatalder list
# 或
fatalder l
```

### 查看帮助

```bash
fatalder help
# 或
fatalder h
```

## 📱 Termux 路径说明

在 Termux 中访问 Android 文件系统：

```bash
# Android 存储路径
/sdcard/          # 主存储
/storage/emulated/0/  # 主存储（备用路径）

# Minecraft 世界路径示例
/sdcard/games/com.mojang/minecraftWorlds/World1
/sdcard/games/com.netease/minecraftWorlds/World1
```

## 🎯 完整示例

### 示例 1: 转换结构文件

```bash
# 将 Schematic 转换为 MCStructure
fatalder convert \
  /sdcard/Download/building.schematic \
  MCStructure \
  /sdcard/Download/building.mcstructure
```

### 示例 2: 生成地图画

```bash
# 将图片转换为 2x2 地图画
fatalder mapart \
  /sdcard/Download/photo.jpg \
  /sdcard/games/com.mojang/minecraftWorlds/MyWorld \
  --width 2 \
  --height 2 \
  --2d
```

### 示例 3: 解密网易存档

```bash
# 解密网易版世界
fatalder decrypt \
  /sdcard/games/com.netease/minecraftWorlds/World1
```

## ⚙️ 编译选项

如果需要自己编译，可以调整编译参数：

```bash
# 最小体积（推荐）
go build -ldflags="-s -w" -o fatalder-termux main.go

# 包含调试信息
go build -o fatalder-termux main.go

# 指定架构（如果需要）
GOOS=android GOARCH=arm64 go build -o fatalder-termux main.go
```

## 🔧 故障排除

### 问题 1: 权限被拒绝

```bash
chmod +x fatalder-termux
```

### 问题 2: 找不到文件

确保使用绝对路径，例如：
```bash
# 正确
fatalder convert /sdcard/Download/file.schematic MCStructure

# 错误
fatalder convert ~/file.schematic MCStructure
```

### 问题 3: 依赖缺失

```bash
cd Termux脚本
go mod download
```

### 问题 4: 编译失败

确保 Go 版本 >= 1.25:
```bash
go version
pkg upgrade golang  # 如果需要更新
```

## 📦 支持的功能

### 结构格式转换
- Schematic (Java 版)
- SchemV1 / SchemV2
- Litematic
- MCStructure (基岩版)
- MCWorld
- BDX
- Construction
- AxiomBP
- MCFunction
- KBDX
- 以及其他多种格式

### 地图画转换
- 支持 JPG, PNG 等图片格式
- 多地图拼接（1x1 到 NxM）
- 2D/3D 模式
- 自定义位置和尺寸

### 存档加密/解密
- 网易版世界加密
- 网易版世界解密
- 支持 .mcworld 文件和世界目录

## 📝 注意事项

1. **文件路径**: 使用绝对路径，特别是访问 Android 文件系统
2. **权限**: 某些操作可能需要存储权限
3. **性能**: 大文件转换可能需要较长时间
4. **存储空间**: 确保有足够的存储空间
5. **临时文件**: 转换过程会创建临时文件，确保有足够空间

## 🎉 开始使用

```bash
# 查看帮助
fatalder help

# 列出支持的格式
fatalder list

# 开始转换
fatalder convert input.schematic MCStructure
```
