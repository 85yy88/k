package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	wsdefine "github.com/Yeah114/WaterStructure/define"
	wsstructure "github.com/Yeah114/WaterStructure/structure"
)

// calculateQuota 计算结构文件的额度
// 1. 统计普通方块、NBT方块、命令方块的数量
// 2. 让用户输入三种方块的额度（单价）
// 3. 计算总额度
func calculateQuota(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("无法打开文件: %w", err)
	}
	defer file.Close()

	structure, err := wsstructure.StructureFromFile(file)
	if err != nil {
		return fmt.Errorf("无法识别文件格式: %w", err)
	}
	defer structure.Close()

	size := structure.GetSize()
	fmt.Printf("文件: %s\n", filePath)
	fmt.Printf("尺寸: %d × %d × %d (宽×高×长)\n", size.Width, size.Height, size.Length)
	fmt.Println()

	// 统计所有非空气方块
	fmt.Println("正在统计方块数量...")
	totalBlockCount, err := structure.CountNonAirBlocks()
	if err != nil {
		return fmt.Errorf("统计方块数量失败: %w", err)
	}

	// 统计命令方块和其他NBT方块
	fmt.Println("正在统计NBT方块数量...")
	commandBlockCount, otherNBTCount := countNBTBlocks(structure)

	// 计算普通方块数量（总方块数 - 命令方块数 - NBT方块数）
	normalBlockCount := totalBlockCount - commandBlockCount - otherNBTCount

	// 显示统计结果
	fmt.Println()
	fmt.Println("=" + strings.Repeat("=", 60) + "=")
	fmt.Println("方块数量统计")
	fmt.Println("=" + strings.Repeat("=", 60) + "=")
	fmt.Printf("普通方块数量:      %d\n", normalBlockCount)
	fmt.Printf("NBT方块数量:       %d\n", otherNBTCount)
	fmt.Printf("命令方块数量:      %d\n", commandBlockCount)
	fmt.Printf("总方块数量:        %d (不含空气)\n", totalBlockCount)
	fmt.Println("=" + strings.Repeat("=", 60) + "=")
	fmt.Println()

	// 读取用户输入的额度
	reader := bufio.NewReader(os.Stdin)

	// 输入普通方块额度
	fmt.Print("请输入普通方块额度（单价）: ")
	normalQuotaStr, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("读取输入失败: %w", err)
	}
	normalQuotaStr = strings.TrimSpace(normalQuotaStr)
	normalQuota, err := strconv.ParseFloat(normalQuotaStr, 64)
	if err != nil {
		return fmt.Errorf("无效的普通方块额度: %v", err)
	}

	// 输入NBT方块额度
	fmt.Print("请输入NBT方块额度（单价）: ")
	nbtQuotaStr, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("读取输入失败: %w", err)
	}
	nbtQuotaStr = strings.TrimSpace(nbtQuotaStr)
	nbtQuota, err := strconv.ParseFloat(nbtQuotaStr, 64)
	if err != nil {
		return fmt.Errorf("无效的NBT方块额度: %v", err)
	}

	// 输入命令方块额度
	fmt.Print("请输入命令方块额度（单价）: ")
	commandQuotaStr, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("读取输入失败: %w", err)
	}
	commandQuotaStr = strings.TrimSpace(commandQuotaStr)
	commandQuota, err := strconv.ParseFloat(commandQuotaStr, 64)
	if err != nil {
		return fmt.Errorf("无效的命令方块额度: %v", err)
	}

	// 计算总额度
	normalCost := float64(normalBlockCount) * normalQuota
	nbtCost := float64(otherNBTCount) * nbtQuota
	commandCost := float64(commandBlockCount) * commandQuota
	totalCost := normalCost + nbtCost + commandCost

	// 显示计算结果
	fmt.Println()
	fmt.Println("=" + strings.Repeat("=", 60) + "=")
	fmt.Println("额度计算结果")
	fmt.Println("=" + strings.Repeat("=", 60) + "=")
	fmt.Printf("普通方块: %d × %.2f = %.2f\n", normalBlockCount, normalQuota, normalCost)
	fmt.Printf("NBT方块:  %d × %.2f = %.2f\n", otherNBTCount, nbtQuota, nbtCost)
	fmt.Printf("命令方块: %d × %.2f = %.2f\n", commandBlockCount, commandQuota, commandCost)
	fmt.Println(strings.Repeat("-", 62))
	fmt.Printf("总额度: %.2f\n", totalCost)
	fmt.Println("=" + strings.Repeat("=", 60) + "=")

	return nil
}

// countNBTBlocks 统计命令方块和其他NBT方块数量
// 参考源码/control/task/build_task.go 中的实现
func countNBTBlocks(structure wsstructure.Structure) (commandBlockCount, otherNBTCount int) {
	size := structure.GetSize()
	xCount := size.GetChunkXCount()
	zCount := size.GetChunkZCount()

	// 遍历所有区块
	for cx := 0; cx < xCount; cx++ {
		for cz := 0; cz < zCount; cz++ {
			chunksNBT, err := structure.GetChunksNBT([]wsdefine.ChunkPos{{int32(cx), int32(cz)}})
			if err != nil {
				continue
			}

			chunkNBT := chunksNBT[wsdefine.ChunkPos{int32(cx), int32(cz)}]
			for _, nbt := range chunkNBT {
				// 判断是否是命令方块（通过NBT中的Command字段）
				if _, hasCommand := nbt["Command"]; hasCommand {
					commandBlockCount++
				} else {
					// 其他NBT方块
					otherNBTCount++
				}
			}
		}
	}

	return commandBlockCount, otherNBTCount
}

