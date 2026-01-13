package main

import (
	"bufio"
	"errors"
	"fmt"
	"image"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/TriM-Organization/bedrock-world-operator/chunk"
	bwo_define "github.com/TriM-Organization/bedrock-world-operator/define"
	"github.com/TriM-Organization/bedrock-world-operator/world"
	"github.com/disintegration/imaging"
	"github.com/mholt/archiver/v3"
	"image/color"
	"image/draw"
	"image/png"
	"sort"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	"github.com/Yeah114/blocks"
	wsdefine "github.com/Yeah114/WaterStructure/define"
	wsmapart "github.com/Yeah114/WaterStructure/utils/map_art"
	wsnetease "github.com/Yeah114/WaterStructure/utils/netease_world"
	wsstructure "github.com/Yeah114/WaterStructure/structure"
)

var selectionRegex = regexp.MustCompile(`@\[(-?\d+),(-?\d+),(-?\d+)\]~\[(-?\d+),(-?\d+),(-?\d+)\]`)

// isCommand 判断参数是否是命令
func isCommand(arg string) bool {
	commands := []string{
		"convert", "c",
		"mapart", "m",
		"encrypt", "e",
		"decrypt", "d",
		"list", "l",
		"parse", "p",
		"quota", "q",
		"help", "h", "-h", "--help",
	}
	for _, cmd := range commands {
		if arg == cmd {
			return true
		}
	}
	return false
}

// handleFileByPath 根据文件路径自动识别并处理
func handleFileByPath(filePath string) {
	// 检查路径是否存在
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "错误: 路径不存在: %s\n", filePath)
		os.Exit(1)
	}

	// 如果是目录，显示文件列表让用户选择
	if info.IsDir() {
		handleDirectory(filePath)
		return
	}

	// 如果是文件，直接处理
	processFileByType(filePath)
}

// handleDirectory 处理目录，显示文件列表
func handleDirectory(dirPath string) {
	// 读取目录下的所有文件
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无法读取目录: %v\n", err)
		os.Exit(1)
	}

	// 过滤出文件（排除目录）
	var files []os.DirEntry
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry)
		}
	}

	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "错误: 目录中没有文件\n")
		os.Exit(1)
	}

	// 显示文件列表
	fmt.Println()
	fmt.Println("=" + strings.Repeat("=", 70) + "=")
	fmt.Printf("目录: %s\n", dirPath)
	fmt.Println("=" + strings.Repeat("=", 70) + "=")
	fmt.Println("文件列表:")
	for i, file := range files {
		fmt.Printf("  %d. %s\n", i+1, file.Name())
	}
	fmt.Println("=" + strings.Repeat("=", 70) + "=")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	var selectedFile string

	// 主循环：选择文件并执行功能
	for {
		if selectedFile == "" {
			// 选择文件
			fmt.Print("请选择文件序号 (输入 q 退出): ")
			choice, err := reader.ReadString('\n')
			if err != nil {
				fmt.Fprintf(os.Stderr, "读取输入失败: %v\n", err)
				os.Exit(1)
			}
			choice = strings.TrimSpace(choice)

			if choice == "q" || choice == "Q" {
				fmt.Println("退出")
				os.Exit(0)
			}

			fileIndex, err := strconv.Atoi(choice)
			if err != nil || fileIndex < 1 || fileIndex > len(files) {
				fmt.Fprintf(os.Stderr, "无效的序号，请重新选择\n")
				continue
			}

			selectedFile = filepath.Join(dirPath, files[fileIndex-1].Name())
			fmt.Printf("\n已选择文件: %s\n", files[fileIndex-1].Name())
		}

		// 根据文件类型显示功能菜单
		ext := strings.ToLower(filepath.Ext(selectedFile))
		if isStructureFile(ext) {
			// 结构文件
			if !showStructureFileOptionsLoop(selectedFile, reader) {
				// 用户选择退出或切换文件
				selectedFile = ""
				continue
			}
		} else if ext == ".mcworld" || strings.HasSuffix(strings.ToLower(selectedFile), ".mcworld") {
			// MCWorld文件
			if !showMCWorldFileOptionsLoop(selectedFile, reader) {
				// 用户选择退出或切换文件
				selectedFile = ""
				continue
			}
		} else if isImageFile(ext) {
			// 图片文件
			fmt.Println("图片文件需要指定世界文件，请使用命令方式:")
			fmt.Printf("  %s mapart %s <世界文件/目录> [选项]\n", os.Args[0], selectedFile)
			selectedFile = ""
			continue
		} else {
			// 未知文件类型，尝试作为结构文件处理
			fmt.Println("未识别的文件类型，尝试作为结构文件处理...")
			if !showStructureFileOptionsLoop(selectedFile, reader) {
				selectedFile = ""
				continue
			}
		}
	}
}

// processFileByType 根据文件类型处理单个文件
func processFileByType(filePath string) {
	ext := strings.ToLower(filepath.Ext(filePath))
	if isStructureFile(ext) {
		showStructureFileOptions(filePath)
	} else if ext == ".mcworld" || strings.HasSuffix(strings.ToLower(filePath), ".mcworld") {
		showMCWorldFileOptions(filePath)
	} else if isImageFile(ext) {
		fmt.Fprintf(os.Stderr, "错误: 地图画转换需要指定世界文件\n")
		fmt.Fprintf(os.Stderr, "用法: %s mapart %s <世界文件/目录> [选项]\n", os.Args[0], filePath)
		os.Exit(1)
	} else {
		fmt.Println("未识别的文件类型，尝试作为结构文件处理...")
		showStructureFileOptions(filePath)
	}
}

// isStructureFile 判断是否是结构文件
func isStructureFile(ext string) bool {
	structureExts := []string{
		".bdx", ".schematic", ".litematic", ".mcstructure",
		".schem", ".nbt", ".schemv1", ".schemv2",
		".tibi", ".bds", ".construction", ".nexus_np",
		".axiom_bp", ".gangban_v3", ".fuhong_v4", ".kbdx",
		".ibimport",
	}
	for _, e := range structureExts {
		if ext == e {
			return true
		}
	}
	return false
}

// isImageFile 判断是否是图片文件
func isImageFile(ext string) bool {
	imageExts := []string{".jpg", ".jpeg", ".png", ".bmp", ".gif", ".webp"}
	for _, e := range imageExts {
		if ext == e {
			return true
		}
	}
	return false
}

// showStructureFileOptions 显示结构文件的可用选项（单次执行）
func showStructureFileOptions(filePath string) {
	reader := bufio.NewReader(os.Stdin)
	showStructureFileOptionsLoop(filePath, reader)
}

// showStructureFileOptionsLoop 显示结构文件的可用选项（循环执行，返回false表示退出或切换文件）
func showStructureFileOptionsLoop(filePath string, reader *bufio.Reader) bool {
	fmt.Println()
	fmt.Println("检测到结构文件，请选择功能：")
	fmt.Println("1. 转换格式")
	fmt.Println("2. 解析文件（生成报告图片）")
	fmt.Println("3. 计算额度")
	fmt.Println("4. 文件优化")
	fmt.Println("5. 切换文件")
	fmt.Println("6. 退出")
	fmt.Print("请选择 (1-6): ")

	choice, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取输入失败: %v\n", err)
		return false
	}
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		// 转换格式
		fmt.Print("请输入目标格式: ")
		targetFormat, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "读取输入失败: %v\n", err)
			return true
		}
		targetFormat = strings.TrimSpace(targetFormat)
		
		fmt.Print("请输入输出文件路径（留空自动生成）: ")
		outputPath, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "读取输入失败: %v\n", err)
			return true
		}
		outputPath = strings.TrimSpace(outputPath)

		useFast := false
		fmt.Print("是否使用快速模式？(y/n，默认n): ")
		fastChoice, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(fastChoice)) == "y" {
			useFast = true
		}

		if err := convertStructure(filePath, targetFormat, outputPath, useFast); err != nil {
			fmt.Fprintf(os.Stderr, "转换失败: %v\n", err)
		} else {
			fmt.Println("✓ 转换完成！")
		}
		return true // 继续当前文件

	case "2":
		// 解析文件
		if err := parseStructureFile(filePath); err != nil {
			fmt.Fprintf(os.Stderr, "解析失败: %v\n", err)
		} else {
			fmt.Println("✓ 图片已生成！")
		}
		return true // 继续当前文件

	case "3":
		// 计算额度
		if err := calculateQuota(filePath); err != nil {
			fmt.Fprintf(os.Stderr, "计算失败: %v\n", err)
		}
		return true // 继续当前文件

	case "4":
		// 文件优化
		handleFileOptimization(filePath, reader)
		return true // 继续当前文件

	case "5":
		// 切换文件
		return false

	case "6":
		// 退出
		fmt.Println("退出")
		os.Exit(0)

	default:
		fmt.Fprintf(os.Stderr, "无效的选择\n")
		return true // 继续当前文件
	}
	return true
}

// showMCWorldFileOptions 显示MCWorld文件的可用选项（单次执行）
func showMCWorldFileOptions(filePath string) {
	reader := bufio.NewReader(os.Stdin)
	showMCWorldFileOptionsLoop(filePath, reader)
}

// showMCWorldFileOptionsLoop 显示MCWorld文件的可用选项（循环执行，返回false表示退出或切换文件）
func showMCWorldFileOptionsLoop(filePath string, reader *bufio.Reader) bool {
	fmt.Println()
	fmt.Println("检测到MCWorld文件，请选择功能：")
	fmt.Println("1. 加密")
	fmt.Println("2. 解密")
	fmt.Println("3. 导出结构方块保存的结构")
	fmt.Println("4. 切换文件")
	fmt.Println("5. 退出")
	fmt.Print("请选择 (1-5): ")

	choice, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取输入失败: %v\n", err)
		return false
	}
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		// 加密
		fmt.Print("请输入输出文件路径（留空自动生成）: ")
		outputPath, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "读取输入失败: %v\n", err)
			return true
		}
		outputPath = strings.TrimSpace(outputPath)

		if err := neteaseCrypt(filePath, outputPath, true); err != nil {
			fmt.Fprintf(os.Stderr, "加密失败: %v\n", err)
		} else {
			fmt.Println("✓ 加密完成！")
		}
		return true // 继续当前文件

	case "2":
		// 解密
		fmt.Print("请输入输出文件路径（留空自动生成）: ")
		outputPath, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "读取输入失败: %v\n", err)
			return true
		}
		outputPath = strings.TrimSpace(outputPath)

		if err := neteaseCrypt(filePath, outputPath, false); err != nil {
			fmt.Fprintf(os.Stderr, "解密失败: %v\n", err)
		} else {
			fmt.Println("✓ 解密完成！")
		}
		return true // 继续当前文件

	case "3":
		// 导出结构方块保存的结构
		handleMCWorldStructureExport(filePath, reader)
		return true // 继续当前文件

	case "4":
		// 切换文件
		return false

	case "5":
		// 退出
		fmt.Println("退出")
		os.Exit(0)

	default:
		fmt.Fprintf(os.Stderr, "无效的选择\n")
		return true // 继续当前文件
	}
	return true
}

func main() {
	// 如果提供了命令行参数，使用命令模式
	if len(os.Args) >= 2 {
		firstArg := os.Args[1]
		if isCommand(firstArg) {
			// 命令模式
			handleCommandMode(firstArg)
			return
		}
	}

	// 否则显示主菜单
	showMainMenu()
}

// handleCommandMode 处理命令模式
func handleCommandMode(command string) {
	switch command {
	case "convert", "c":
		if len(os.Args) < 4 {
			fmt.Fprintf(os.Stderr, "错误: 转换命令需要输入文件和目标格式\n")
			fmt.Fprintf(os.Stderr, "用法: %s convert <输入文件> <目标格式> [输出文件] [--fast]\n", os.Args[0])
			fmt.Fprintf(os.Stderr, "      --fast: 使用快速模式（多线程，适合大文件）\n")
			os.Exit(1)
		}
		inputPath := os.Args[2]
		targetFormat := os.Args[3]
		var outputPath string
		useFast := false
		for i := 4; i < len(os.Args); i++ {
			if os.Args[i] == "--fast" {
				useFast = true
			} else if outputPath == "" {
				outputPath = os.Args[i]
			}
		}
		if err := convertStructure(inputPath, targetFormat, outputPath, useFast); err != nil {
			fmt.Fprintf(os.Stderr, "转换失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✓ 转换完成！")

	case "mapart", "m":
		if len(os.Args) < 4 {
			fmt.Fprintf(os.Stderr, "错误: 地图画命令需要图片文件和世界文件\n")
			fmt.Fprintf(os.Stderr, "用法: %s mapart <图片文件> <世界文件/目录> [输出文件] [选项]\n", os.Args[0])
			fmt.Fprintf(os.Stderr, "选项:\n")
			fmt.Fprintf(os.Stderr, "  --x <X坐标>       起始 X 坐标（子区块，默认0）\n")
			fmt.Fprintf(os.Stderr, "  --y <Y坐标>       起始 Y 坐标（子区块，默认-4）\n")
			fmt.Fprintf(os.Stderr, "  --z <Z坐标>       起始 Z 坐标（子区块，默认0）\n")
			fmt.Fprintf(os.Stderr, "  --width <宽度>    地图宽度（地图数量，默认1）\n")
			fmt.Fprintf(os.Stderr, "  --height <高度>   地图高度（地图数量，默认1）\n")
			fmt.Fprintf(os.Stderr, "  --2d              强制2D模式（平面）\n")
			fmt.Fprintf(os.Stderr, "  --no-ref          禁用参考列\n")
			fmt.Fprintf(os.Stderr, "  --max3d <高度>    最大3D高度（默认0，无限制）\n")
			os.Exit(1)
		}
		imagePath := os.Args[2]
		worldPath := os.Args[3]
		// 解析参数：输出文件和选项
		var outputPath string
		var options []string
		for i := 4; i < len(os.Args); i++ {
			if strings.HasPrefix(os.Args[i], "--") {
				options = append(options, os.Args[i:]...)
				break
			} else if outputPath == "" {
				outputPath = os.Args[i]
			} else {
				options = append(options, os.Args[i:]...)
				break
			}
		}
		if err := convertMapArt(imagePath, worldPath, outputPath, options); err != nil {
			fmt.Fprintf(os.Stderr, "地图画转换失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✓ 地图画转换完成！")

	case "encrypt", "e":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "错误: 加密命令需要世界文件或目录\n")
			fmt.Fprintf(os.Stderr, "用法: %s encrypt <世界文件/目录> [输出文件]\n", os.Args[0])
			os.Exit(1)
		}
		worldPath := os.Args[2]
		var outputPath string
		if len(os.Args) >= 4 {
			outputPath = os.Args[3]
		}
		if err := neteaseCrypt(worldPath, outputPath, true); err != nil {
			fmt.Fprintf(os.Stderr, "加密失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✓ 加密完成！")

	case "decrypt", "d":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "错误: 解密命令需要世界文件或目录\n")
			fmt.Fprintf(os.Stderr, "用法: %s decrypt <世界文件/目录> [输出文件]\n", os.Args[0])
			os.Exit(1)
		}
		worldPath := os.Args[2]
		var outputPath string
		if len(os.Args) >= 4 {
			outputPath = os.Args[3]
		}
		if err := neteaseCrypt(worldPath, outputPath, false); err != nil {
			fmt.Fprintf(os.Stderr, "解密失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✓ 解密完成！")

	case "list", "l":
		listFormats()

	case "parse", "p":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "错误: 解析命令需要文件路径\n")
			fmt.Fprintf(os.Stderr, "用法: %s parse <文件路径>\n", os.Args[0])
			os.Exit(1)
		}
		filePath := os.Args[2]
		if err := parseStructureFile(filePath); err != nil {
			fmt.Fprintf(os.Stderr, "解析失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✓ 解析完成！")

	case "quota", "q":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "错误: 计算额度命令需要文件路径\n")
			fmt.Fprintf(os.Stderr, "用法: %s quota <文件路径>\n", os.Args[0])
			os.Exit(1)
		}
		filePath := os.Args[2]
		if err := calculateQuota(filePath); err != nil {
			fmt.Fprintf(os.Stderr, "计算失败: %v\n", err)
			os.Exit(1)
		}

	case "help", "h", "-h", "--help":
		printUsage()

	default:
		fmt.Fprintf(os.Stderr, "错误: 未知命令 '%s'\n", command)
		printUsage()
		os.Exit(1)
	}
}

// showMainMenu 显示主菜单
func showMainMenu() {
	reader := bufio.NewReader(os.Stdin)
	
	for {
		fmt.Println()
		fmt.Println("=" + strings.Repeat("=", 70) + "=")
		fmt.Println("Minecraft 工具集 - 主菜单")
		fmt.Println("=" + strings.Repeat("=", 70) + "=")
		fmt.Println("1. 地图画转换")
		fmt.Println("2. 存档加密")
		fmt.Println("3. 存档解密")
		fmt.Println("4. 文件功能")
		fmt.Println("5. 退出")
		fmt.Println("=" + strings.Repeat("=", 70) + "=")
		fmt.Print("请选择 (1-5): ")

		choice, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "读取输入失败: %v\n", err)
			os.Exit(1)
		}
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			handleMapArt(reader)
		case "2":
			handleEncrypt(reader)
		case "3":
			handleDecrypt(reader)
		case "4":
			handleFileFunction(reader)
		case "5":
			fmt.Println("退出")
			os.Exit(0)
		default:
			fmt.Fprintf(os.Stderr, "无效的选择，请重新输入\n")
		}
	}
}

// handleMapArt 处理地图画转换
func handleMapArt(reader *bufio.Reader) {
	fmt.Println()
	fmt.Println("地图画转换")
	fmt.Println("请输入图片文件路径（例如: /storage/emulated/0/Download/图片.jpg）:")
	fmt.Print("> ")
	imagePath, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取输入失败: %v\n", err)
		return
	}
	imagePath = strings.TrimSpace(imagePath)

	// 检查文件是否存在
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "错误: 文件不存在: %s\n", imagePath)
		return
	}

	fmt.Println("请输入世界文件/目录路径:")
	fmt.Print("> ")
	worldPath, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取输入失败: %v\n", err)
		return
	}
	worldPath = strings.TrimSpace(worldPath)

	fmt.Println("请输入输出文件路径（留空自动生成）:")
	fmt.Print("> ")
	outputPath, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取输入失败: %v\n", err)
		return
	}
	outputPath = strings.TrimSpace(outputPath)

	// 解析选项
	var options []string
	fmt.Println("请输入选项（留空跳过，格式: --x 0 --y -4 --z 0 --width 1 --height 1 --2d --no-ref --max3d 10）:")
	fmt.Print("> ")
	optionsStr, err := reader.ReadString('\n')
	if err == nil {
		optionsStr = strings.TrimSpace(optionsStr)
		if optionsStr != "" {
			options = strings.Fields(optionsStr)
		}
	}

	if err := convertMapArt(imagePath, worldPath, outputPath, options); err != nil {
		fmt.Fprintf(os.Stderr, "地图画转换失败: %v\n", err)
	} else {
		fmt.Println("✓ 地图画转换完成！")
	}
}

// handleEncrypt 处理存档加密
func handleEncrypt(reader *bufio.Reader) {
	fmt.Println()
	fmt.Println("存档加密")
	fmt.Println("请输入世界文件/目录路径:")
	fmt.Print("> ")
	worldPath, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取输入失败: %v\n", err)
		return
	}
	worldPath = strings.TrimSpace(worldPath)

	fmt.Println("请输入输出文件路径（留空自动生成）:")
	fmt.Print("> ")
	outputPath, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取输入失败: %v\n", err)
		return
	}
	outputPath = strings.TrimSpace(outputPath)

	if err := neteaseCrypt(worldPath, outputPath, true); err != nil {
		fmt.Fprintf(os.Stderr, "加密失败: %v\n", err)
	} else {
		fmt.Println("✓ 加密完成！")
	}
}

// handleDecrypt 处理存档解密
func handleDecrypt(reader *bufio.Reader) {
	fmt.Println()
	fmt.Println("存档解密")
	fmt.Println("请输入世界文件/目录路径:")
	fmt.Print("> ")
	worldPath, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取输入失败: %v\n", err)
		return
	}
	worldPath = strings.TrimSpace(worldPath)

	fmt.Println("请输入输出文件路径（留空自动生成）:")
	fmt.Print("> ")
	outputPath, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取输入失败: %v\n", err)
		return
	}
	outputPath = strings.TrimSpace(outputPath)

	if err := neteaseCrypt(worldPath, outputPath, false); err != nil {
		fmt.Fprintf(os.Stderr, "解密失败: %v\n", err)
	} else {
		fmt.Println("✓ 解密完成！")
	}
}

// handleMCWorldStructureExport 处理MCWorld中结构方块保存的结构导出
func handleMCWorldStructureExport(mcworldPath string, reader *bufio.Reader) {
	fmt.Println()
	fmt.Println("导出结构方块保存的结构")

	// 解压MCWorld
	extractDir, cleanup, err := unarchiveMCWorldToTempDir(mcworldPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "解压MCWorld失败: %v\n", err)
		return
	}
	defer cleanup()

	// 查找structures目录
	structuresDir := filepath.Join(extractDir, "structures")
	structures, err := listMCWorldStructures(structuresDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取结构列表失败: %v\n", err)
		return
	}

	if len(structures) == 0 {
		fmt.Println("未找到保存的结构")
		return
	}

	// 显示结构列表
	fmt.Println()
	fmt.Println("=" + strings.Repeat("=", 70) + "=")
	fmt.Println("结构方块保存的结构列表:")
	for i, name := range structures {
		fmt.Printf("  %d. %s\n", i+1, name)
	}
	fmt.Println("=" + strings.Repeat("=", 70) + "=")
	fmt.Print("请选择结构序号: ")

	choice, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取输入失败: %v\n", err)
		return
	}
	choice = strings.TrimSpace(choice)
	index, err := strconv.Atoi(choice)
	if err != nil || index < 1 || index > len(structures) {
		fmt.Fprintf(os.Stderr, "无效的序号\n")
		return
	}

	selectedStructure := structures[index-1]

	// 输入输出文件路径
	fmt.Println("请输入输出文件路径（留空自动生成）:")
	fmt.Print("> ")
	outputPath, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取输入失败: %v\n", err)
		return
	}
	outputPath = strings.TrimSpace(outputPath)

	// 自动生成输出文件名
	if outputPath == "" {
		baseName := strings.TrimSuffix(filepath.Base(mcworldPath), filepath.Ext(mcworldPath))
		outputPath = filepath.Join(filepath.Dir(mcworldPath), baseName+"_"+selectedStructure+".mcstructure")
	}

	// 确保输出目录存在
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无法创建输出目录: %v\n", err)
		return
	}

	// 复制结构文件
	sourceFile := filepath.Join(structuresDir, selectedStructure+".mcstructure")
	if err := copyFile(sourceFile, outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "导出失败: %v\n", err)
	} else {
		fmt.Printf("✓ 导出完成！输出文件: %s\n", outputPath)
	}
}

// listMCWorldStructures 列出MCWorld中保存的结构
func listMCWorldStructures(structuresDir string) ([]string, error) {
	// 检查structures目录是否存在
	if _, err := os.Stat(structuresDir); os.IsNotExist(err) {
		return []string{}, nil
	}

	// 读取目录
	entries, err := os.ReadDir(structuresDir)
	if err != nil {
		return nil, err
	}

	var structures []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".mcstructure") {
			// 移除.mcstructure扩展名
			name := strings.TrimSuffix(entry.Name(), ".mcstructure")
			structures = append(structures, name)
		}
	}

	// 排序
	for i := 0; i < len(structures)-1; i++ {
		for j := i + 1; j < len(structures); j++ {
			if structures[i] > structures[j] {
				structures[i], structures[j] = structures[j], structures[i]
			}
		}
	}

	return structures, nil
}

// handleMCWorldExport 处理MCWorld导出（按坐标导出）
func handleMCWorldExport(mcworldPath string, reader *bufio.Reader) {
	fmt.Println()
	fmt.Println("MCWorld导出")

	// 输入起始坐标
	fmt.Println("请输入起始坐标 (格式: x y z，例如: 0 -64 0):")
	fmt.Print("> ")
	startCoordStr, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取输入失败: %v\n", err)
		return
	}
	startCoordStr = strings.TrimSpace(startCoordStr)
	startCoords := strings.Fields(startCoordStr)
	if len(startCoords) != 3 {
		fmt.Fprintf(os.Stderr, "错误: 坐标格式不正确，需要3个数字\n")
		return
	}
	startX, err1 := strconv.ParseInt(startCoords[0], 10, 32)
	startY, err2 := strconv.ParseInt(startCoords[1], 10, 32)
	startZ, err3 := strconv.ParseInt(startCoords[2], 10, 32)
	if err1 != nil || err2 != nil || err3 != nil {
		fmt.Fprintf(os.Stderr, "错误: 坐标格式不正确，必须是整数\n")
		return
	}

	// 输入终止坐标
	fmt.Println("请输入终止坐标 (格式: x y z，例如: 15 15 15):")
	fmt.Print("> ")
	endCoordStr, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取输入失败: %v\n", err)
		return
	}
	endCoordStr = strings.TrimSpace(endCoordStr)
	endCoords := strings.Fields(endCoordStr)
	if len(endCoords) != 3 {
		fmt.Fprintf(os.Stderr, "错误: 坐标格式不正确，需要3个数字\n")
		return
	}
	endX, err1 := strconv.ParseInt(endCoords[0], 10, 32)
	endY, err2 := strconv.ParseInt(endCoords[1], 10, 32)
	endZ, err3 := strconv.ParseInt(endCoords[2], 10, 32)
	if err1 != nil || err2 != nil || err3 != nil {
		fmt.Fprintf(os.Stderr, "错误: 坐标格式不正确，必须是整数\n")
		return
	}

	// 显示支持的格式列表
	formats := getSupportedFormats()
	fmt.Println()
	fmt.Println("支持的导出格式:")
	for i, format := range formats {
		fmt.Printf("  %d. %s\n", i+1, format)
	}
	fmt.Print("请选择格式序号: ")
	formatChoice, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取输入失败: %v\n", err)
		return
	}
	formatChoice = strings.TrimSpace(formatChoice)
	formatIndex, err := strconv.Atoi(formatChoice)
	if err != nil || formatIndex < 1 || formatIndex > len(formats) {
		fmt.Fprintf(os.Stderr, "错误: 无效的格式序号\n")
		return
	}
	targetFormat := formats[formatIndex-1]

	// 检查格式是否支持
	targetFactory, ok := wsstructure.StructureNamePool[targetFormat]
	if !ok {
		fmt.Fprintf(os.Stderr, "错误: 不支持的目标格式: %s\n", targetFormat)
		return
	}

	// 输入输出文件路径
	fmt.Println("请输入输出文件路径（留空自动生成）:")
	fmt.Print("> ")
	outputPath, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取输入失败: %v\n", err)
		return
	}
	outputPath = strings.TrimSpace(outputPath)

	// 自动生成输出文件名
	if outputPath == "" {
		baseName := strings.TrimSuffix(filepath.Base(mcworldPath), filepath.Ext(mcworldPath))
		outputPath = filepath.Join(filepath.Dir(mcworldPath), baseName+"_export."+strings.ToLower(targetFormat))
	}

	// 确保输出目录存在
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无法创建输出目录: %v\n", err)
		return
	}

	// 执行导出
	if err := exportFromMCWorld(mcworldPath, outputPath, targetFormat, targetFactory, int32(startX), int32(startY), int32(startZ), int32(endX), int32(endY), int32(endZ)); err != nil {
		fmt.Fprintf(os.Stderr, "导出失败: %v\n", err)
	} else {
		fmt.Printf("✓ 导出完成！输出文件: %s\n", outputPath)
	}
}

// exportFromMCWorld 从MCWorld导出结构文件
func exportFromMCWorld(mcworldPath, outputPath, targetFormat string, targetFactory wsstructure.StructureFunc, startX, startY, startZ, endX, endY, endZ int32) error {
	// 解压MCWorld
	extractDir, cleanup, err := unarchiveMCWorldToTempDir(mcworldPath)
	if err != nil {
		return fmt.Errorf("解压MCWorld失败: %w", err)
	}
	defer cleanup()

	// 打开世界
	bw, err := world.Open(extractDir, nil)
	if err != nil {
		return fmt.Errorf("打开世界失败: %w", err)
	}
	defer func() {
		_ = bw.CloseWorld()
		_ = bw.Close()
	}()

	// 创建输出文件
	outputFile, err := os.OpenFile(outputPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %w", err)
	}
	defer outputFile.Close()

	// 构建坐标
	startPos := wsdefine.BlockPos{startX, startY, startZ}
	endPos := wsdefine.BlockPos{endX, endY, endZ}

	// 导出结构
	targetStruct := targetFactory()
	fmt.Println("正在导出...")
	if err := targetStruct.FromMCWorld(
		bw,
		outputFile,
		startPos,
		endPos,
		func(total int) {
			fmt.Printf("总进度: %d 个子区块\n", total)
		},
		func() {
			// 进度回调（可以在这里添加进度条）
		},
	); err != nil {
		return fmt.Errorf("导出结构失败: %w", err)
	}

	return nil
}

// handleFileFunction 处理文件功能
func handleFileFunction(reader *bufio.Reader) {
	fmt.Println()
	fmt.Println("文件功能")
	fmt.Println("请输入目录路径（例如: /storage/emulated/0/Download）:")
	fmt.Print("> ")
	dirPath, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取输入失败: %v\n", err)
		return
	}
	dirPath = strings.TrimSpace(dirPath)

	// 检查路径是否存在
	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "错误: 路径不存在: %s\n", dirPath)
		return
	}

	// 如果是目录，显示文件列表让用户选择
	if info.IsDir() {
		handleDirectory(dirPath)
		return
	}

	// 如果是文件，直接处理
	processFileByType(dirPath)
}

func printUsage() {
	fmt.Println("Minecraft 工具集 - Termux 版本")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Printf("  %s <命令> [参数...]\n", os.Args[0])
	fmt.Println()
	fmt.Println("命令:")
	fmt.Println("  convert, c    - 转换结构文件格式")
	fmt.Println("                用法: convert <输入文件> <目标格式> [输出文件]")
	fmt.Println()
	fmt.Println("  mapart, m    - 将图片转换为地图画")
	fmt.Println("                用法: mapart <图片文件> <世界文件/目录> [选项]")
	fmt.Println("                选项: --x <X坐标> --y <Y坐标> --z <Z坐标>")
	fmt.Println("                      --width <地图宽度> --height <地图高度>")
	fmt.Println("                      --2d (强制2D模式)")
	fmt.Println()
	fmt.Println("  encrypt, e   - 加密网易版世界存档")
	fmt.Println("                用法: encrypt <世界文件/目录>")
	fmt.Println()
	fmt.Println("  decrypt, d   - 解密网易版世界存档")
	fmt.Println("                用法: decrypt <世界文件/目录>")
	fmt.Println()
	fmt.Println("  parse, p     - 解析结构文件并生成报告图片")
	fmt.Println("                用法: parse <文件路径>")
	fmt.Println("                功能: 统计方块、查找容器、显示物品信息")
	fmt.Println()
	fmt.Println("  quota, q     - 计算结构文件额度")
	fmt.Println("                用法: quota <文件路径>")
	fmt.Println("                功能: 统计方块数量、命令方块数量、NBT方块数量")
	fmt.Println()
	fmt.Println("  list, l      - 列出所有支持的格式")
	fmt.Println()
	fmt.Println("  help, h      - 显示帮助信息")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Printf("  %s convert input.schematic MCStructure output.mcstructure\n", os.Args[0])
	fmt.Printf("  %s convert input.schematic MCStructure output.mcstructure --fast\n", os.Args[0])
	fmt.Printf("  %s mapart image.jpg world.mcworld output.mapart.mcworld --width 2 --height 2\n", os.Args[0])
	fmt.Printf("  %s mapart image.png world.mcworld --2d --no-ref --max3d 10\n", os.Args[0])
	fmt.Printf("  %s encrypt world.mcworld world.encrypted.mcworld\n", os.Args[0])
	fmt.Printf("  %s decrypt world.mcworld world.decrypted.mcworld\n", os.Args[0])
	fmt.Printf("  %s decrypt /sdcard/games/com.netease/minecraftWorlds/World1\n", os.Args[0])
	fmt.Printf("  %s parse /storage/emulated/0/Download/文件.bdx\n", os.Args[0])
	fmt.Printf("  %s quota /storage/emulated/0/Download/文件.bdx\n", os.Args[0])
}

func listFormats() {
	fmt.Println("支持的格式:")
	formats := getSupportedFormats()
	for i, format := range formats {
		fmt.Printf("  %d. %s\n", i+1, format)
	}
}

func getSupportedFormats() []string {
	var formats []string
	for name := range wsstructure.StructureNamePool {
		formats = append(formats, name)
	}
	// 排序
	for i := 0; i < len(formats)-1; i++ {
		for j := i + 1; j < len(formats); j++ {
			if formats[i] > formats[j] {
				formats[i], formats[j] = formats[j], formats[i]
			}
		}
	}
	return formats
}

// convertStructure 转换结构文件格式
func convertStructure(srcPath, targetFormat, destPath string, useFast bool) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("无法打开源文件: %w", err)
	}
	defer srcFile.Close()

	srcStruct, err := wsstructure.StructureFromFile(srcFile)
	if err != nil {
		return fmt.Errorf("无法识别源文件格式: %w", err)
	}
	defer srcStruct.Close()

	fmt.Printf("检测到源格式: %s\n", srcStruct.Name())

	targetFactory, ok := wsstructure.StructureNamePool[targetFormat]
	if !ok {
		return fmt.Errorf("不支持的目标格式: %s\n使用 'list' 命令查看支持的格式", targetFormat)
	}

	if destPath == "" {
		ext := strings.ToLower(filepath.Ext(srcPath))
		baseName := strings.TrimSuffix(srcPath, ext)
		destPath = baseName + "." + strings.ToLower(targetFormat)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("无法创建输出目录: %w", err)
	}

	destFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("无法创建输出文件: %w", err)
	}
	defer destFile.Close()

	// 尝试从 MCWorld 源直接导出（优化路径）
	if handled, err := tryExportFromMCWorldSource(srcPath, targetFormat, targetFactory, destFile); handled {
		if err != nil {
			return err
		}
		fmt.Printf("输出文件: %s\n", destPath)
		return nil
	}

	fmt.Println("开始转换...")

	tmpDir, err := os.MkdirTemp("", "fatalder-convert-*")
	if err != nil {
		return fmt.Errorf("无法创建临时目录: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	worldDir := filepath.Join(tmpDir, "world")
	if err := os.MkdirAll(worldDir, 0755); err != nil {
		return fmt.Errorf("无法创建世界目录: %w", err)
	}

	bedrockWorld, err := world.Open(worldDir, nil)
	if err != nil {
		return fmt.Errorf("无法创建临时世界: %w", err)
	}
	defer func() { _ = bedrockWorld.CloseWorld() }()

	startSubChunkPos := wsdefine.SubChunkPos{0, -4, 0}
	fmt.Println("步骤 1/2: 将源结构转换为临时世界...")
	
	if useFast {
		// 使用快速模式（多线程）
		if err := convertReaderToMCWorldFast(srcStruct, bedrockWorld, bwo_define.SubChunkPos(startSubChunkPos), func(int) {}, func() {}); err != nil {
			return fmt.Errorf("写入世界失败: %w", err)
		}
	} else {
		// 使用标准模式
		if err := srcStruct.ToMCWorld(
			bedrockWorld,
			startSubChunkPos,
			func(int) {},
			func() {},
		); err != nil {
			return fmt.Errorf("写入世界失败: %w", err)
		}
	}

	size := srcStruct.GetSize()
	
	// 如果目标格式是 MCWorld，设置世界名称并直接打包
	if targetFormat == wsstructure.NameMCWorld {
		structureName := strings.TrimSuffix(filepath.Base(srcPath), filepath.Ext(srcPath))
		worldName := fmt.Sprintf("%s@[0,-64,0]~[%d,%d,%d]",
			structureName,
			size.Width-1,
			size.Height-64-1,
			size.Length-1,
		)
		bedrockWorld.LevelDat().LevelName = worldName
		if err := bedrockWorld.CloseWorld(); err != nil {
			return fmt.Errorf("关闭世界失败: %w", err)
		}
		if err := archiveDirAsMCWorld(worldDir, destPath); err != nil {
			return fmt.Errorf("打包MCWorld失败: %w", err)
		}
		fmt.Printf("输出文件: %s\n", destPath)
		fmt.Printf("世界名称: %s\n", worldName)
		return nil
	}
	
	// 其他格式：从临时世界导出
	startBlockPos := wsdefine.BlockPos{
		startSubChunkPos.X() * 16,
		startSubChunkPos.Y() * 16,
		startSubChunkPos.Z() * 16,
	}
	endBlockPos := wsdefine.BlockPos{
		startBlockPos.X() + int32(size.Width) - 1,
		startBlockPos.Y() + int32(size.Height) - 1,
		startBlockPos.Z() + int32(size.Length) - 1,
	}

	fmt.Println("步骤 2/2: 从临时世界导出为目标格式...")
	targetStruct := targetFactory()
	if err := targetStruct.FromMCWorld(
		bedrockWorld,
		destFile,
		startBlockPos,
		endBlockPos,
		func(int) {},
		func() {},
	); err != nil {
		return fmt.Errorf("导出结构失败: %w", err)
	}

	fmt.Printf("输出文件: %s\n", destPath)
	return nil
}

// convertMapArt 将图片转换为地图画
func convertMapArt(imagePath, worldPath, outputPath string, options []string) error {
	img, err := imaging.Open(imagePath)
	if err != nil {
		return fmt.Errorf("无法打开图片: %w", err)
	}

	// 解析选项
	opts := &wsmapart.Options{
		StartSubChunkPos: wsdefine.SubChunkPos{0, -4, 0},
		MapWidth:         1,
		MapHeight:        1,
		Force2D:          false,
		DisableReferenceColumn: false,
		Max3DHeight:      0,
	}

	for i := 0; i < len(options); i++ {
		switch options[i] {
		case "--x":
			if i+1 < len(options) {
				x, _ := strconv.ParseInt(options[i+1], 10, 32)
				opts.StartSubChunkPos = wsdefine.SubChunkPos{int32(x), opts.StartSubChunkPos.Y(), opts.StartSubChunkPos.Z()}
				i++
			}
		case "--y":
			if i+1 < len(options) {
				y, _ := strconv.ParseInt(options[i+1], 10, 32)
				opts.StartSubChunkPos = wsdefine.SubChunkPos{opts.StartSubChunkPos.X(), int32(y), opts.StartSubChunkPos.Z()}
				i++
			}
		case "--z":
			if i+1 < len(options) {
				z, _ := strconv.ParseInt(options[i+1], 10, 32)
				opts.StartSubChunkPos = wsdefine.SubChunkPos{opts.StartSubChunkPos.X(), opts.StartSubChunkPos.Y(), int32(z)}
				i++
			}
		case "--width":
			if i+1 < len(options) {
				w, _ := strconv.Atoi(options[i+1])
				opts.MapWidth = w
				i++
			}
		case "--height":
			if i+1 < len(options) {
				h, _ := strconv.Atoi(options[i+1])
				opts.MapHeight = h
				i++
			}
		case "--2d":
			opts.Force2D = true
		case "--no-ref":
			opts.DisableReferenceColumn = true
		case "--max3d":
			if i+1 < len(options) {
				h, _ := strconv.ParseInt(options[i+1], 10, 32)
				opts.Max3DHeight = int32(h)
				i++
			}
		}
	}

	info, err := os.Stat(worldPath)
	if err != nil {
		return fmt.Errorf("无法访问世界路径: %w", err)
	}

	var worldDir string
	var cleanup func()
	var isTemp bool

	if !info.IsDir() {
		// 是 .mcworld 文件
		worldDir, cleanup, err = unarchiveMCWorldToTempDir(worldPath)
		if err != nil {
			return fmt.Errorf("无法解压世界文件: %w", err)
		}
		isTemp = true
		defer cleanup()
	} else {
		worldDir = worldPath
		isTemp = false
	}

	fmt.Println("正在生成地图画...")
	minPos, maxPos, err := writeMapArtToWorldDir(worldDir, img, opts)
	if err != nil {
		return err
	}

	fmt.Printf("写入范围: (%d,%d,%d) ~ (%d,%d,%d)\n", minPos[0], minPos[1], minPos[2], maxPos[0], maxPos[1], maxPos[2])

	if isTemp {
		// 如果是临时目录，需要重新打包
		if outputPath == "" {
			outputPath = strings.TrimSuffix(worldPath, filepath.Ext(worldPath)) + ".mapart.mcworld"
		}
		if !strings.HasSuffix(strings.ToLower(outputPath), ".mcworld") {
			outputPath += ".mcworld"
		}
		fmt.Printf("正在打包为: %s\n", outputPath)
		if err := archiveDirAsMCWorld(worldDir, outputPath); err != nil {
			return fmt.Errorf("打包失败: %w", err)
		}
		fmt.Printf("地图画已写入: %s\n", outputPath)
	} else {
		fmt.Printf("地图画已写入: %s\n", worldDir)
	}

	return nil
}

func writeMapArtToWorldDir(worldDir string, img image.Image, opts *wsmapart.Options) (minPos [3]int32, maxPos [3]int32, err error) {
	bedrockWorld, err := world.Open(worldDir, nil)
	if err != nil {
		return [3]int32{}, [3]int32{}, fmt.Errorf("无法打开世界: %w", err)
	}
	defer func() { _ = bedrockWorld.CloseWorld() }()

	return wsmapart.GenerateMapArtToWorld(bedrockWorld, img, opts)
}

// neteaseCrypt 网易版世界加密/解密
func neteaseCrypt(worldPath, outputPath string, encrypt bool) error {
	info, err := os.Stat(worldPath)
	if err != nil {
		return fmt.Errorf("无法访问路径: %w", err)
	}

	var dbDir string
	var isTemp bool
	var cleanup func()

	if !info.IsDir() {
		// 是 .mcworld 文件
		worldDir, cleanup, err := unarchiveMCWorldToTempDir(worldPath)
		if err != nil {
			return fmt.Errorf("无法解压世界文件: %w", err)
		}
		defer cleanup()
		isTemp = true

		dbDir = filepath.Join(worldDir, "db")
		if _, err := os.Stat(dbDir); err != nil {
			return fmt.Errorf("缺少 db 目录: %w", err)
		}
	} else {
		// 是世界目录
		if filepath.Base(worldPath) == "db" {
			dbDir = worldPath
		} else {
			dbDir = filepath.Join(worldPath, "db")
		}

		if _, err := os.Stat(filepath.Join(dbDir, "CURRENT")); err != nil {
			return fmt.Errorf("无效的 db 目录: %w", err)
		}
		isTemp = false
	}

	action := "解密"
	if encrypt {
		action = "加密"
	}

	fmt.Printf("正在%s数据库: %s\n", action, dbDir)

	if encrypt {
		err = wsnetease.Encrypt(dbDir, nil)
	} else {
		err = wsnetease.Decrypt(dbDir, nil)
	}

	if err != nil {
		return fmt.Errorf("%s失败: %w", action, err)
	}

	if isTemp {
		// 如果是临时目录，需要重新打包
		if outputPath == "" {
			outputPath = strings.TrimSuffix(worldPath, filepath.Ext(worldPath))
			if encrypt {
				outputPath += ".encrypted.mcworld"
			} else {
				outputPath += ".decrypted.mcworld"
			}
		}
		if !strings.HasSuffix(strings.ToLower(outputPath), ".mcworld") {
			outputPath += ".mcworld"
		}
		fmt.Printf("正在打包为: %s\n", outputPath)
		worldDir := filepath.Dir(dbDir)
		if err := archiveDirAsMCWorld(worldDir, outputPath); err != nil {
			return fmt.Errorf("打包失败: %w", err)
		}
		fmt.Printf("%s完成: %s\n", action, outputPath)
	} else {
		fmt.Printf("%s完成: %s\n", action, dbDir)
	}

	return nil
}

func unarchiveMCWorldToTempDir(mcworldPath string) (string, func(), error) {
	tempDir, err := os.MkdirTemp("", "fatalder-mcworld-*")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() { _ = os.RemoveAll(tempDir) }

	z := archiver.Zip{}
	if err := z.Unarchive(mcworldPath, tempDir); err != nil {
		cleanup()
		return "", nil, err
	}
	return tempDir, cleanup, nil
}

func archiveDirAsMCWorld(worldDir string, outPath string) error {
	entries, err := os.ReadDir(worldDir)
	if err != nil {
		return err
	}
	var inputs []string
	for _, entry := range entries {
		inputs = append(inputs, filepath.Join(worldDir, entry.Name()))
	}

	tmpDir, err := os.MkdirTemp("", "fatalder-mcworld-zip-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	tmpZip := filepath.Join(tmpDir, "world.zip")
	z := archiver.Zip{}
	if err := z.Archive(inputs, tmpZip); err != nil {
		return err
	}

	if err := os.RemoveAll(outPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := os.Rename(tmpZip, outPath); err != nil {
		// 如果重命名失败，尝试复制
		return copyFile(tmpZip, outPath)
	}
	return nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = dstFile.ReadFrom(srcFile)
	return err
}

// tryExportFromMCWorldSource 尝试直接从 MCWorld 源导出（优化路径）
func tryExportFromMCWorldSource(
	structurePath string,
	targetFormat string,
	targetFactory wsstructure.StructureFunc,
	targetFile *os.File,
) (handled bool, err error) {
	ext := strings.ToLower(filepath.Ext(structurePath))
	if ext != ".mcworld" && ext != ".zip" {
		return false, nil
	}

	// 如果目标格式也是 MCWorld，直接复制
	if targetFormat == wsstructure.NameMCWorld {
		src, err := os.Open(structurePath)
		if err != nil {
			return true, err
		}
		defer src.Close()
		if _, err := io.Copy(targetFile, src); err != nil {
			return true, err
		}
		return true, nil
	}

	// 从 MCWorld 提取并转换
	extractDir, cleanup, err := unarchiveMCWorldToTempDir(structurePath)
	if err != nil {
		return false, nil
	}
	defer cleanup()

	bw, err := world.Open(extractDir, nil)
	if err != nil {
		return false, nil
	}
	defer func() {
		_ = bw.CloseWorld()
		_ = bw.Close()
	}()

	// 尝试从文件名或世界名称解析坐标
	startPos, endPos, ok := parseSelectionBounds(structurePath)
	if !ok {
		startPos, endPos, ok = parseSelectionBounds(bw.LevelDat().LevelName)
	}
	if !ok {
		return true, fmt.Errorf("无法从文件名或世界名称中解析坐标信息，请使用完整转换流程")
	}

	targetStruct := targetFactory()
	if err := targetStruct.FromMCWorld(
		bw,
		targetFile,
		startPos,
		endPos,
		func(int) {},
		func() {},
	); err != nil {
		return true, err
	}
	return true, nil
}

// parseSelectionBounds 从字符串解析选择边界坐标
// 格式: @[x1,y1,z1]~[x2,y2,z2]
func parseSelectionBounds(target string) (start wsdefine.BlockPos, end wsdefine.BlockPos, ok bool) {
	matches := selectionRegex.FindStringSubmatch(target)
	if len(matches) != 7 {
		return wsdefine.BlockPos{}, wsdefine.BlockPos{}, false
	}

	parse := func(s string) (int32, bool) {
		v, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return 0, false
		}
		return int32(v), true
	}

	sx, ok1 := parse(matches[1])
	sy, ok2 := parse(matches[2])
	sz, ok3 := parse(matches[3])
	ex, ok4 := parse(matches[4])
	ey, ok5 := parse(matches[5])
	ez, ok6 := parse(matches[6])
	if !(ok1 && ok2 && ok3 && ok4 && ok5 && ok6) {
		return wsdefine.BlockPos{}, wsdefine.BlockPos{}, false
	}

	minPos := wsdefine.BlockPos{minInt32(sx, ex), minInt32(sy, ey), minInt32(sz, ez)}
	maxPos := wsdefine.BlockPos{maxInt32(sx, ex), maxInt32(sy, ey), maxInt32(sz, ez)}
	return minPos, maxPos, true
}

func minInt32(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

func maxInt32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

// convertReaderToMCWorldFast 快速转换模式（多线程批量处理）
func convertReaderToMCWorldFast(reader wsstructure.Structure, bedrockWorld *world.BedrockWorld, startSubChunkPos bwo_define.SubChunkPos, startCallback func(int), progressCallback func()) error {
	if reader == nil {
		return errors.New("reader is nil")
	}
	if bedrockWorld == nil {
		return errors.New("bedrock world is nil")
	}

	size := reader.GetSize()
	xCount := size.GetChunkXCount()
	zCount := size.GetChunkZCount()
	totalChunks := xCount * zCount
	if startCallback != nil {
		startCallback(totalChunks)
	}
	if totalChunks == 0 {
		return nil
	}

	allChunkPos := make([]wsdefine.ChunkPos, 0, totalChunks)
	for x := 0; x < xCount; x++ {
		for z := 0; z < zCount; z++ {
			allChunkPos = append(allChunkPos, wsdefine.ChunkPos{int32(x), int32(z)})
		}
	}

	batchSize := 256
	if totalChunks < batchSize {
		batchSize = totalChunks
	}
	if batchSize <= 0 {
		batchSize = 1
	}

	type batchResult struct {
		positions []wsdefine.ChunkPos
		chunks    map[wsdefine.ChunkPos]*chunk.Chunk
		nbts      map[wsdefine.ChunkPos]map[wsdefine.BlockPos]map[string]any
		err       error
	}

	taskCh := make(chan []wsdefine.ChunkPos)
	resultCh := make(chan batchResult)

	workerCount := runtime.NumCPU()
	if workerCount < 1 {
		workerCount = 1
	}

	var wg sync.WaitGroup
	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func() {
			defer wg.Done()
			for positions := range taskCh {
				chunks, err := reader.GetChunks(positions)
				if err != nil {
					resultCh <- batchResult{positions: positions, err: err}
					continue
				}
				nbts, err := reader.GetChunksNBT(positions)
				if err != nil {
					resultCh <- batchResult{positions: positions, chunks: chunks, err: err}
					continue
				}
				resultCh <- batchResult{positions: positions, chunks: chunks, nbts: nbts}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	go func() {
		for i := 0; i < len(allChunkPos); i += batchSize {
			end := i + batchSize
			if end > len(allChunkPos) {
				end = len(allChunkPos)
			}
			taskCh <- allChunkPos[i:end]
		}
		close(taskCh)
	}()

	chunkOffsetX := startSubChunkPos.X()
	chunkOffsetZ := startSubChunkPos.Z()
	blockYOffset := startSubChunkPos.Y() * 16

	for res := range resultCh {
		if res.err != nil {
			return res.err
		}

		for _, pos := range res.positions {
			chunkData, ok := res.chunks[pos]
			if ok && chunkData != nil {
				chunkData.Compact()
				targetPos := bwo_define.ChunkPos{pos.X() + chunkOffsetX, pos.Z() + chunkOffsetZ}
				if err := bedrockWorld.SaveChunk(bwo_define.DimensionIDOverworld, targetPos, chunkData); err != nil {
					return err
				}
			}
			if progressCallback != nil {
				progressCallback()
			}
		}

		for cpos, blockMap := range res.nbts {
			if len(blockMap) == 0 {
				continue
			}
			list := make([]map[string]any, 0, len(blockMap))
			absChunkX := (cpos.X() + chunkOffsetX) * 16
			absChunkZ := (cpos.Z() + chunkOffsetZ) * 16
			for bpos, n := range blockMap {
				if n == nil {
					continue
				}
				m := make(map[string]any, len(n)+3)
				for k, v := range n {
					m[k] = v
				}
				m["x"] = absChunkX + bpos.X()
				m["y"] = blockYOffset + bpos.Y() + 64
				m["z"] = absChunkZ + bpos.Z()
				list = append(list, m)
			}
			if len(list) == 0 {
				continue
			}
			targetPos := bwo_define.ChunkPos{cpos.X() + chunkOffsetX, cpos.Z() + chunkOffsetZ}
			if err := bedrockWorld.SaveNBT(bwo_define.DimensionIDOverworld, targetPos, list); err != nil {
				return err
			}
		}
	}

	return nil
}

// handleFileOptimization 处理文件优化菜单
func handleFileOptimization(filePath string, reader *bufio.Reader) {
	for {
		fmt.Println()
		fmt.Println("=" + strings.Repeat("=", 70) + "=")
		fmt.Println("文件优化功能")
		fmt.Println("=" + strings.Repeat("=", 70) + "=")
		fmt.Println("1. 删除方块")
		fmt.Println("2. 替换方块")
		fmt.Println("3. 添加拒绝方块")
		fmt.Println("4. 返回")
		fmt.Println("=" + strings.Repeat("=", 70) + "=")
		fmt.Print("请选择 (1-4): ")

		choice, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "读取输入失败: %v\n", err)
			return
		}
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			// 删除方块
			handleDeleteBlocks(filePath, reader)
		case "2":
			// 替换方块
			handleReplaceBlocks(filePath, reader)
		case "3":
			// 添加拒绝方块
			handleAddDenyBlocks(filePath, reader)
		case "4":
			// 返回
			return
		default:
			fmt.Fprintf(os.Stderr, "无效的选择\n")
		}
	}
}

// handleDeleteBlocks 处理删除方块
func handleDeleteBlocks(filePath string, reader *bufio.Reader) {
	fmt.Println()
	fmt.Println("删除方块")
	fmt.Println("提示：可以在解析图片中找到方块名字")
	fmt.Print("请输入要删除的方块名字（例如: minecraft:stone）: ")
	blockName, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取输入失败: %v\n", err)
		return
	}
	blockName = strings.TrimSpace(blockName)
	if blockName == "" {
		fmt.Fprintf(os.Stderr, "方块名字不能为空\n")
		return
	}

	fmt.Print("请输入输出文件路径（留空覆盖原文件）: ")
	outputPath, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取输入失败: %v\n", err)
		return
	}
	outputPath = strings.TrimSpace(outputPath)
	if outputPath == "" {
		outputPath = filePath
	}

	count, err := deleteBlocksInFile(filePath, blockName, outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "删除失败: %v\n", err)
	} else {
		fmt.Printf("✓ 删除完成！共删除 %d 个方块\n", count)
	}
}

// handleReplaceBlocks 处理替换方块
func handleReplaceBlocks(filePath string, reader *bufio.Reader) {
	fmt.Println()
	fmt.Println("替换方块")
	fmt.Println("提示：可以在解析图片中找到方块名字")
	fmt.Print("请输入被替换的方块名字（例如: minecraft:stone）: ")
	oldBlockName, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取输入失败: %v\n", err)
		return
	}
	oldBlockName = strings.TrimSpace(oldBlockName)
	if oldBlockName == "" {
		fmt.Fprintf(os.Stderr, "方块名字不能为空\n")
		return
	}

	fmt.Print("请输入替换成的方块名字（例如: minecraft:air）: ")
	newBlockName, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取输入失败: %v\n", err)
		return
	}
	newBlockName = strings.TrimSpace(newBlockName)
	if newBlockName == "" {
		fmt.Fprintf(os.Stderr, "方块名字不能为空\n")
		return
	}

	fmt.Print("请输入输出文件路径（留空覆盖原文件）: ")
	outputPath, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取输入失败: %v\n", err)
		return
	}
	outputPath = strings.TrimSpace(outputPath)
	if outputPath == "" {
		outputPath = filePath
	}

	count, err := replaceBlocksInFile(filePath, oldBlockName, newBlockName, outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "替换失败: %v\n", err)
	} else {
		fmt.Printf("✓ 替换完成！共替换 %d 个方块\n", count)
	}
}

// handleAddDenyBlocks 处理添加拒绝方块
func handleAddDenyBlocks(filePath string, reader *bufio.Reader) {
	fmt.Println()
	fmt.Println("添加拒绝方块")
	fmt.Println("说明：将在建筑最底下按xz坐标生成拒绝方块，建筑整体上移一格")

	fmt.Print("请输入输出文件路径（留空覆盖原文件）: ")
	outputPath, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取输入失败: %v\n", err)
		return
	}
	outputPath = strings.TrimSpace(outputPath)
	if outputPath == "" {
		outputPath = filePath
	}

	err = addDenyBlocksToFile(filePath, outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "添加失败: %v\n", err)
	} else {
		fmt.Println("✓ 添加完成！")
	}
}

// deleteBlocksInFile 删除文件中的指定方块
func deleteBlocksInFile(filePath, blockName, outputPath string) (int, error) {
	// 转换为air
	return replaceBlocksInFile(filePath, blockName, "minecraft:air", outputPath)
}

// replaceBlocksInFile 替换文件中的方块
func replaceBlocksInFile(filePath, oldBlockName, newBlockName, outputPath string) (int, error) {
	// 打开源文件
	srcFile, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("无法打开源文件: %w", err)
	}
	defer srcFile.Close()

	// 读取结构
	reader, err := wsstructure.StructureFromFile(srcFile)
	if err != nil {
		return 0, fmt.Errorf("无法识别文件格式: %w", err)
	}
	defer reader.Close()

	// 获取目标方块的RuntimeID
	newRuntimeID, found := blocks.BlockStrToRuntimeID(newBlockName)
	if !found {
		return 0, fmt.Errorf("无法识别方块: %s", newBlockName)
	}

	// 获取旧方块的RuntimeID（用于匹配）
	oldRuntimeID, found := blocks.BlockStrToRuntimeID(oldBlockName)
	if !found {
		return 0, fmt.Errorf("无法识别方块: %s", oldBlockName)
	}

	// 创建临时MCWorld
	tempDir, err := os.MkdirTemp("", "fatalder-optimize-*")
	if err != nil {
		return 0, fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(tempDir)

	worldDir := filepath.Join(tempDir, "world")
	if err := os.MkdirAll(worldDir, 0755); err != nil {
		return 0, fmt.Errorf("创建世界目录失败: %w", err)
	}

	bedrockWorld, err := world.Open(worldDir, nil)
	if err != nil {
		return 0, fmt.Errorf("打开世界失败: %w", err)
	}

	// 将结构写入临时世界
	size := reader.GetSize()
	if err := reader.ToMCWorld(bedrockWorld, bwo_define.SubChunkPos{0, -4, 0}, func(int) {}, func() {}); err != nil {
		bedrockWorld.CloseWorld()
		return 0, fmt.Errorf("写入世界失败: %w", err)
	}

	// 使用原始结构数据来查找需要替换的方块位置
	// 写入MCWorld时，结构从SubChunkPos{0, -4, 0}开始，即区块(0,0)的Y=-64位置
	xCount := size.GetChunkXCount()
	zCount := size.GetChunkZCount()
	allChunkPos := make([]wsdefine.ChunkPos, 0, xCount*zCount)
	for x := 0; x < xCount; x++ {
		for z := 0; z < zCount; z++ {
			allChunkPos = append(allChunkPos, wsdefine.ChunkPos{int32(x), int32(z)})
		}
	}

	chunks, err := reader.GetChunks(allChunkPos)
	if err != nil {
		bedrockWorld.CloseWorld()
		return 0, fmt.Errorf("获取区块失败: %w", err)
	}

	// 替换方块
	// MCWorld坐标：结构写入到区块(0,0)的Y=-64位置
	replacedCount := 0
	for cpos, c := range chunks {
		// 结构内的相对区块坐标
		structChunkX := cpos.X()
		structChunkZ := cpos.Z()
		// 转换为MCWorld的绝对区块坐标（写入时从(0,0)开始）
		worldChunkX := structChunkX
		worldChunkZ := structChunkZ
		
		for localX := uint8(0); localX < 16; localX++ {
			for localZ := uint8(0); localZ < 16; localZ++ {
				for y := int16(-64); y < int16(size.Height-64); y++ {
					blockRuntimeID := c.Block(localX, y, localZ, 0)
					if blockRuntimeID == oldRuntimeID {
						// 计算MCWorld中的绝对坐标
						worldX := int32(worldChunkX*16) + int32(localX)
						worldZ := int32(worldChunkZ*16) + int32(localZ)
						worldY := int16(y) // Y坐标从-64开始
						
						if err := bedrockWorld.SetBlock(worldX, worldY, worldZ, newRuntimeID); err != nil {
							bedrockWorld.CloseWorld()
							return replacedCount, fmt.Errorf("设置方块失败: %w", err)
						}
						replacedCount++
					}
				}
			}
		}
	}

	// 保存世界（已经关闭过了，这里需要重新关闭）
	if err := bedrockWorld.CloseWorld(); err != nil {
		return replacedCount, fmt.Errorf("保存世界失败: %w", err)
	}

	// 从MCWorld导出回原格式
	ext := strings.ToLower(filepath.Ext(filePath))
	targetFormat := strings.TrimPrefix(ext, ".")
	if targetFormat == "" || targetFormat == "mcworld" || targetFormat == "zip" {
		targetFormat = "MCStructure" // 默认使用MCStructure
	}

	return exportFromMCWorld(worldDir, filePath, outputPath, targetFormat, replacedCount)
}

// addDenyBlocksToFile 添加拒绝方块
func addDenyBlocksToFile(filePath, outputPath string) error {
	// 打开源文件
	srcFile, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("无法打开源文件: %w", err)
	}
	defer srcFile.Close()

	// 读取结构
	reader, err := wsstructure.StructureFromFile(srcFile)
	if err != nil {
		return fmt.Errorf("无法识别文件格式: %w", err)
	}
	defer reader.Close()

	// 获取拒绝方块的RuntimeID
	denyRuntimeID, found := blocks.BlockStrToRuntimeID("minecraft:deny")
	if !found {
		return fmt.Errorf("无法识别拒绝方块")
	}

	// 创建临时MCWorld
	tempDir, err := os.MkdirTemp("", "fatalder-deny-*")
	if err != nil {
		return fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(tempDir)

	worldDir := filepath.Join(tempDir, "world")
	if err := os.MkdirAll(worldDir, 0755); err != nil {
		return fmt.Errorf("创建世界目录失败: %w", err)
	}

	bedrockWorld, err := world.Open(worldDir, nil)
	if err != nil {
		return fmt.Errorf("打开世界失败: %w", err)
	}

	// 将结构写入临时世界（上移一格）
	size := reader.GetSize()
	offset := reader.GetOffsetPos()
	// 上移一格：从-3开始而不是-4
	if err := reader.ToMCWorld(bedrockWorld, bwo_define.SubChunkPos{0, -3, 0}, func(int) {}, func() {}); err != nil {
		bedrockWorld.CloseWorld()
		return fmt.Errorf("写入世界失败: %w", err)
	}

	// 在底部添加拒绝方块层
	xCount := size.GetChunkXCount()
	zCount := size.GetChunkZCount()
	allChunkPos := make([]wsdefine.ChunkPos, 0, xCount*zCount)
	for x := 0; x < xCount; x++ {
		for z := 0; z < zCount; z++ {
			allChunkPos = append(allChunkPos, wsdefine.ChunkPos{int32(x), int32(z)})
		}
	}

	chunks, err := reader.GetChunks(allChunkPos)
	if err != nil {
		bedrockWorld.CloseWorld()
		return fmt.Errorf("获取区块失败: %w", err)
	}

	// 找到最底部的Y坐标
	minY := int16(0)
	for _, c := range chunks {
		for localX := uint8(0); localX < 16; localX++ {
			for localZ := uint8(0); localZ < 16; localZ++ {
				for y := int16(-64); y < int16(size.Height-64); y++ {
					blockRuntimeID := c.Block(localX, y, localZ, 0)
					if blockRuntimeID != blocks.AIR_RUNTIMEID {
						if y < minY {
							minY = y
						}
						break
					}
				}
			}
		}
	}

	// 在minY-1位置添加拒绝方块层
	denyY := minY - 1 + offset.Y()
	for cpos := range chunks {
		chunkWorldX := cpos.X() * 16
		chunkWorldZ := cpos.Z() * 16
		for localX := uint8(0); localX < 16; localX++ {
			for localZ := uint8(0); localZ < 16; localZ++ {
				worldX := int32(chunkWorldX) + int32(localX) + offset.X()
				worldZ := int32(chunkWorldZ) + int32(localZ) + offset.Z()
				if err := bedrockWorld.SetBlock(worldX, denyY, worldZ, denyRuntimeID); err != nil {
					bedrockWorld.CloseWorld()
					return fmt.Errorf("设置拒绝方块失败: %w", err)
				}
			}
		}
	}

	// 保存世界
	if err := bedrockWorld.CloseWorld(); err != nil {
		return fmt.Errorf("保存世界失败: %w", err)
	}

	// 从MCWorld导出回原格式
	ext := strings.ToLower(filepath.Ext(filePath))
	targetFormat := strings.TrimPrefix(ext, ".")
	if targetFormat == "" || targetFormat == "mcworld" || targetFormat == "zip" {
		targetFormat = "MCStructure"
	}

	_, err = exportFromMCWorld(worldDir, filePath, outputPath, targetFormat, 0)
	return err
}

// exportFromMCWorld 从MCWorld导出结构
func exportFromMCWorld(worldDir, originalPath, outputPath, targetFormat string, operationCount int) (int, error) {
	// 打开临时世界
	bedrockWorld, err := world.Open(worldDir, nil)
	if err != nil {
		return operationCount, fmt.Errorf("打开世界失败: %w", err)
	}
	defer bedrockWorld.CloseWorld()

	// 获取结构范围
	size := bedrockWorld.Bounds()
	startPos := wsdefine.BlockPos{size[0][0], size[0][1], size[0][2]}
	endPos := wsdefine.BlockPos{size[1][0], size[1][1], size[1][2]}

	// 创建目标格式的结构
	targetFactory, ok := wsstructure.StructureNamePool[targetFormat]
	if !ok {
		return operationCount, fmt.Errorf("不支持的目标格式: %s", targetFormat)
	}

	// 创建输出文件
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return operationCount, fmt.Errorf("创建输出文件失败: %w", err)
	}
	defer outputFile.Close()

	// 导出结构
	targetStruct := targetFactory()
	if err := targetStruct.FromMCWorld(bedrockWorld, outputFile, startPos, endPos, func(int) {}, func() {}); err != nil {
		return operationCount, fmt.Errorf("导出结构失败: %w", err)
	}

	return operationCount, nil
}
