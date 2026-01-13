package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/TriM-Organization/bedrock-world-operator/chunk"
	bwo_define "github.com/TriM-Organization/bedrock-world-operator/define"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	"github.com/Yeah114/blocks"
	wsdefine "github.com/Yeah114/WaterStructure/define"
	wsstructure "github.com/Yeah114/WaterStructure/structure"
)

// parseStructureFile 解析结构文件并生成图片报告
func parseStructureFile(filePath string) error {
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
	offset := structure.GetOffsetPos()

	fmt.Println("正在解析文件...")

	// 获取所有区块
	xCount := size.GetChunkXCount()
	zCount := size.GetChunkZCount()
	allChunkPos := make([]wsdefine.ChunkPos, 0, xCount*zCount)
	for x := 0; x < xCount; x++ {
		for z := 0; z < zCount; z++ {
			allChunkPos = append(allChunkPos, wsdefine.ChunkPos{int32(x), int32(z)})
		}
	}

	fmt.Println("正在读取方块数据...")
	chunks, err := structure.GetChunks(allChunkPos)
	if err != nil {
		return fmt.Errorf("读取方块数据失败: %w", err)
	}

	fmt.Println("正在读取NBT数据...")
	chunksNBT, err := structure.GetChunksNBT(allChunkPos)
	if err != nil {
		return fmt.Errorf("读取NBT数据失败: %w", err)
	}

	// 统计方块
	blockCounts := make(map[string]int)
	containers := []ContainerInfo{}

	fmt.Println("正在分析方块和容器...")
	for chunkPos, chunkData := range chunks {
		if chunkData == nil {
			continue
		}

		chunkX := int(chunkPos.X())
		chunkZ := int(chunkPos.Z())

		// 遍历区块内的方块
		for subChunkY := 0; subChunkY < 24; subChunkY++ {
			subChunk := chunkData.SubChunk(bwo_define.DimensionIDOverworld, subChunkY)
			if subChunk == nil {
				continue
			}

			for localX := 0; localX < 16; localX++ {
				for localZ := 0; localZ < 16; localZ++ {
					worldX := chunkX*16 + localX
					worldY := subChunkY*16 - 64
					worldZ := chunkZ*16 + localZ

					// 检查是否在结构范围内
					if worldX < 0 || worldX >= size.Width ||
						worldY < -64 || worldY >= size.Height-64 ||
						worldZ < 0 || worldZ >= size.Length {
						continue
					}

					runtimeID := subChunk.BlockRuntimeID(localX, 0, localZ, 0)
					if runtimeID == 0 { // 空气
						continue
					}

					// 获取方块名称
					block, found := blocks.RuntimeIDToBlock(runtimeID)
					if !found {
						blockCounts["未知方块"]++
						continue
					}

					blockName := block.ShortName()
					if !strings.Contains(blockName, ":") {
						blockName = "minecraft:" + blockName
					}

					blockCounts[blockName]++

					// 检查是否是容器
					blockPos := wsdefine.BlockPos{
						int32(worldX),
						int32(worldY),
						int32(worldZ),
					}

					if chunkNBTs, ok := chunksNBT[chunkPos]; ok {
						if nbtData, ok := chunkNBTs[blockPos]; ok {
							containerInfo := parseContainer(blockName, worldX, worldY, worldZ, nbtData)
							if containerInfo != nil {
								containers = append(containers, *containerInfo)
							}
						}
					}
				}
			}
		}
	}

	// 生成图片
	outputPath := strings.TrimSuffix(filePath, filepath.Ext(filePath)) + "_解析报告.png"
	if err := generateReportImage(outputPath, filePath, size, offset, blockCounts, containers); err != nil {
		return fmt.Errorf("生成图片失败: %w", err)
	}

	fmt.Printf("✓ 图片已生成: %s\n", outputPath)
	return nil
}

type ContainerInfo struct {
	BlockName  string
	X, Y, Z    int
	CustomName string
	Items      []ItemInfo
}

type ItemInfo struct {
	Name         string
	Count        int
	Slot         int
	Enchantments []EnchantmentInfo
	CustomName   string
}

type EnchantmentInfo struct {
	ID    string
	Level int
}

func parseContainer(blockName string, x, y, z int, nbtData map[string]any) *ContainerInfo {
	// 检查是否是容器
	isShulker := strings.Contains(blockName, "shulker_box")
	isContainer := strings.Contains(blockName, "chest") ||
		strings.Contains(blockName, "barrel") ||
		isShulker ||
		strings.Contains(blockName, "hopper") ||
		strings.Contains(blockName, "dispenser") ||
		strings.Contains(blockName, "dropper") ||
		strings.Contains(blockName, "furnace") ||
		strings.Contains(blockName, "smoker") ||
		strings.Contains(blockName, "blast_furnace")

	if !isContainer {
		return nil
	}

	container := &ContainerInfo{
		BlockName: blockName,
		X:         x,
		Y:         y,
		Z:         z,
		Items:     []ItemInfo{},
	}

	// 获取自定义名称
	if customName, ok := nbtData["CustomName"].(string); ok {
		container.CustomName = customName
	}

	// 获取物品列表
	itemsRaw, ok := nbtData["Items"]
	if !ok {
		return container
	}

	var itemsList []any
	switch v := itemsRaw.(type) {
	case []any:
		itemsList = v
	case map[string]any:
		itemsList = []any{v}
	default:
		return container
	}

	for _, itemRaw := range itemsList {
		itemMap, ok := itemRaw.(map[string]any)
		if !ok {
			continue
		}

		item := ItemInfo{}

		// 物品名称
		if name, ok := itemMap["Name"].(string); ok {
			item.Name = name
		} else if name, ok := itemMap["id"].(string); ok {
			item.Name = name
		}

		// 数量
		if count, ok := itemMap["Count"].(byte); ok {
			item.Count = int(count)
		} else if count, ok := itemMap["Count"].(int); ok {
			item.Count = count
		}

		// 槽位
		if slot, ok := itemMap["Slot"].(byte); ok {
			item.Slot = int(slot)
		} else if slot, ok := itemMap["Slot"].(int); ok {
			item.Slot = slot
		}

		// 自定义名称
		if customName, ok := itemMap["CustomName"].(string); ok {
			item.CustomName = customName
		}

		// 附魔
		if enchantments, ok := itemMap["Enchantments"].([]any); ok {
			for _, enchRaw := range enchantments {
				enchMap, ok := enchRaw.(map[string]any)
				if !ok {
					continue
				}
				ench := EnchantmentInfo{}
				if id, ok := enchMap["id"].(string); ok {
					ench.ID = id
				}
				if level, ok := enchMap["lvl"].(int16); ok {
					ench.Level = int(level)
				} else if level, ok := enchMap["lvl"].(int); ok {
					ench.Level = level
				}
				item.Enchantments = append(item.Enchantments, ench)
			}
		}

		if item.Name != "" {
			container.Items = append(container.Items, item)
		}
	}

	return container
}

func generateReportImage(outputPath, filePath string, size wsdefine.Size, offset wsdefine.Offset, blockCounts map[string]int, containers []ContainerInfo) error {
	padding := 40
	lineHeight := 25
	tableRowHeight := 22
	containerItemHeight := 18

	// 计算高度
	headerHeight := 120
	blockTableHeight := 50 + len(blockCounts)*tableRowHeight
	if blockTableHeight < 100 {
		blockTableHeight = 100
	}

	containerHeight := 50
	for _, c := range containers {
		containerHeight += 50 // 容器标题
		containerHeight += len(c.Items) * containerItemHeight
		if len(c.Items) == 0 {
			containerHeight += containerItemHeight
		}
		containerHeight += 20 // 间距
	}
	if len(containers) == 0 {
		containerHeight = 50 + 30
	}

	totalHeight := headerHeight + blockTableHeight + containerHeight + padding*3
	width := 1400

	// 创建图片
	img := image.NewRGBA(image.Rect(0, 0, width, totalHeight))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{245, 245, 250, 255}}, image.Point{}, draw.Src)

	// 绘制函数
	drawText := func(x, y int, text string, clr color.Color) {
		point := fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6(y * 64)}
		d := &font.Drawer{
			Dst:  img,
			Src:  image.NewUniform(clr),
			Face: basicfont.Face7x13,
			Dot:  point,
		}
		d.DrawString(text)
	}

	y := padding

	// 标题
	drawText(width/2-100, y, "结构文件解析报告", color.RGBA{0, 0, 0, 255})
	y += 35

	// 文件信息
	fileName := filepath.Base(filePath)
	if len(fileName) > 60 {
		fileName = fileName[:57] + "..."
	}
	drawText(padding, y, "文件: "+fileName, color.RGBA{60, 60, 60, 255})
	y += lineHeight

	// 尺寸信息
	sizeText := fmt.Sprintf("尺寸: %d × %d × %d (宽×高×长)", size.Width, size.Height, size.Length)
	drawText(padding, y, sizeText, color.RGBA{60, 60, 60, 255})
	y += lineHeight

	// 偏移信息
	if offset.X() != 0 || offset.Y() != 0 || offset.Z() != 0 {
		offsetText := fmt.Sprintf("偏移: (%d, %d, %d)", offset.X(), offset.Y(), offset.Z())
		drawText(padding, y, offsetText, color.RGBA{60, 60, 60, 255})
		y += lineHeight
	}

	y += 20

	// 方块统计表格标题
	drawText(padding, y, "方块统计", color.RGBA{0, 0, 0, 255})
	y += 30

	// 绘制表格边框
	tableX := padding
	tableY := y
	tableWidth := width - padding*2
	tableHeight := blockTableHeight

	// 表格背景
	drawRect(img, tableX, tableY, tableWidth, tableHeight, color.RGBA{255, 255, 255, 255})

	// 表头
	headerY := tableY + 25
	drawText(tableX+10, headerY, "方块名称", color.RGBA{0, 0, 0, 255})
	drawText(tableX+tableWidth-100, headerY, "数量", color.RGBA{0, 0, 0, 255})
	drawLine(img, tableX, headerY+10, tableX+tableWidth, headerY+10, color.RGBA{200, 200, 200, 255})

	// 排序方块名称
	blockNames := make([]string, 0, len(blockCounts))
	for name := range blockCounts {
		blockNames = append(blockNames, name)
	}
	sort.Strings(blockNames)

	// 绘制表格行
	rowY := headerY + 20
	for i, name := range blockNames {
		count := blockCounts[name]
		displayName := name
		if strings.HasPrefix(displayName, "minecraft:") {
			displayName = strings.TrimPrefix(displayName, "minecraft:")
		}
		if len(displayName) > 50 {
			displayName = displayName[:47] + "..."
		}

		// 交替行颜色
		if i%2 == 0 {
			drawRect(img, tableX, rowY-15, tableWidth, tableRowHeight, color.RGBA{250, 250, 250, 255})
		}

		drawText(tableX+10, rowY, displayName, color.RGBA{0, 0, 0, 255})
		drawText(tableX+tableWidth-100, rowY, strconv.Itoa(count), color.RGBA{0, 0, 0, 255})
		rowY += tableRowHeight
	}

	// 表格边框
	drawRectBorder(img, tableX, tableY, tableWidth, tableHeight, color.RGBA{180, 180, 180, 255})

	y += blockTableHeight + 30

	// 容器信息标题
	drawText(padding, y, "容器信息", color.RGBA{0, 0, 0, 255})
	y += 30

	if len(containers) == 0 {
		drawText(padding+20, y, "未发现容器", color.RGBA{120, 120, 120, 255})
	} else {
		for i, c := range containers {
			// 容器背景
			containerBoxY := y
			containerBoxHeight := 40 + len(c.Items)*containerItemHeight
			if len(c.Items) == 0 {
				containerBoxHeight = 40 + containerItemHeight
			}
			drawRect(img, padding, containerBoxY, width-padding*2, containerBoxHeight, color.RGBA{255, 255, 255, 255})
			drawRectBorder(img, padding, containerBoxY, width-padding*2, containerBoxHeight, color.RGBA{200, 200, 200, 255})

			itemY := containerBoxY + 25

			// 容器信息
			blockDisplayName := c.BlockName
			if strings.HasPrefix(blockDisplayName, "minecraft:") {
				blockDisplayName = strings.TrimPrefix(blockDisplayName, "minecraft:")
			}
			containerTitle := fmt.Sprintf("[容器 %d] %s", i+1, blockDisplayName)
			drawText(padding+10, itemY, containerTitle, color.RGBA{0, 0, 150, 255})
			itemY += 20

			coordText := fmt.Sprintf("坐标: (%d, %d, %d)", c.X, c.Y, c.Z)
			drawText(padding+20, itemY, coordText, color.RGBA{60, 60, 60, 255})
			itemY += 18

			if c.CustomName != "" {
				nameText := "名称: " + c.CustomName
				if len(nameText) > 80 {
					nameText = nameText[:77] + "..."
				}
				drawText(padding+20, itemY, nameText, color.RGBA{60, 60, 60, 255})
				itemY += 18
			}

			if len(c.Items) == 0 {
				drawText(padding+20, itemY, "物品: 空", color.RGBA{120, 120, 120, 255})
				itemY += containerItemHeight
			} else {
				drawText(padding+20, itemY, fmt.Sprintf("物品: %d 个", len(c.Items)), color.RGBA{60, 60, 60, 255})
				itemY += 20

				for _, item := range c.Items {
					itemName := item.Name
					if strings.HasPrefix(itemName, "minecraft:") {
						itemName = strings.TrimPrefix(itemName, "minecraft:")
					}
					if len(itemName) > 30 {
						itemName = itemName[:27] + "..."
					}

					itemText := fmt.Sprintf("  槽位 %d: %s × %d", item.Slot, itemName, item.Count)
					if item.CustomName != "" {
						customName := item.CustomName
						if len(customName) > 20 {
							customName = customName[:17] + "..."
						}
						itemText += " [" + customName + "]"
					}
					if len(item.Enchantments) > 0 {
						enchStrs := make([]string, len(item.Enchantments))
						for j, ench := range item.Enchantments {
							enchID := ench.ID
							if strings.HasPrefix(enchID, "minecraft:") {
								enchID = strings.TrimPrefix(enchID, "minecraft:")
							}
							enchStrs[j] = fmt.Sprintf("%s %d", enchID, ench.Level)
						}
						itemText += " [附魔: " + strings.Join(enchStrs, ", ") + "]"
					}

					if len(itemText) > 100 {
						itemText = itemText[:97] + "..."
					}

					drawText(padding+30, itemY, itemText, color.RGBA{0, 0, 0, 255})
					itemY += containerItemHeight
				}
			}

			y += containerBoxHeight + 20
		}
	}

	// 保存图片
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return png.Encode(file, img)
}

// 辅助函数：绘制矩形
func drawRect(img *image.RGBA, x, y, w, h int, clr color.Color) {
	rect := image.Rect(x, y, x+w, y+h)
	draw.Draw(img, rect, &image.Uniform{clr}, image.Point{}, draw.Src)
}

// 辅助函数：绘制矩形边框
func drawRectBorder(img *image.RGBA, x, y, w, h int, clr color.Color) {
	// 上
	drawLine(img, x, y, x+w, y, clr)
	// 下
	drawLine(img, x, y+h, x+w, y+h, clr)
	// 左
	drawLine(img, x, y, x, y+h, clr)
	// 右
	drawLine(img, x+w, y, x+w, y+h, clr)
}

// 辅助函数：绘制直线
func drawLine(img *image.RGBA, x1, y1, x2, y2 int, clr color.Color) {
	if x1 == x2 {
		// 垂直线
		if y1 > y2 {
			y1, y2 = y2, y1
		}
		for y := y1; y <= y2; y++ {
			if y >= 0 && y < img.Bounds().Dy() && x1 >= 0 && x1 < img.Bounds().Dx() {
				img.Set(x1, y, clr)
			}
		}
	} else if y1 == y2 {
		// 水平线
		if x1 > x2 {
			x1, x2 = x2, x1
		}
		for x := x1; x <= x2; x++ {
			if x >= 0 && x < img.Bounds().Dx() && y1 >= 0 && y1 < img.Bounds().Dy() {
				img.Set(x, y1, clr)
			}
		}
	}
}
