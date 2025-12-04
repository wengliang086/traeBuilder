package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/game-data-builder/internal/config"
	"github.com/game-data-builder/internal/converter"
	"github.com/game-data-builder/internal/model"
	"github.com/game-data-builder/internal/reader"
	"github.com/game-data-builder/internal/validator"
)

// Builder 数据构建器
type Builder struct {
	configManager    *config.ConfigManager
	readerFactory    *reader.ReaderFactory
	converterFactory *converter.ConverterFactory
	validator        *validator.DefaultValidator
}

// NewBuilder 创建数据构建器
func NewBuilder() *Builder {
	return &Builder{
		configManager:    config.NewConfigManager(),
		readerFactory:    reader.NewReaderFactory(),
		converterFactory: converter.NewConverterFactory(),
		validator:        validator.NewDefaultValidator(),
	}
}

// LoadConfig 加载配置
func (b *Builder) LoadConfig(confDir string) error {
	return b.configManager.Load(confDir)
}

// Build 执行构建过程
func (b *Builder) Build() error {
	startTime := time.Now()

	// 1. 读取源文件
	sheets, err := b.readSourceFiles()
	if err != nil {
		return fmt.Errorf("读取源文件失败: %v", err)
	}

	// 2. 验证数据
	errors := b.validateData(sheets)
	if len(errors) > 0 {
		// 打印验证错误
		for _, err := range errors {
			fmt.Printf("[ERROR] %s:%s[%d]: %s\n", err.Sheet, err.Column, err.Row, err.Msg)
		}
		return fmt.Errorf("数据验证失败，共 %d 个错误", len(errors))
	}

	// 3. 转换数据
	results, err := b.convertData(sheets)
	if err != nil {
		return fmt.Errorf("转换数据失败: %v", err)
	}

	// 4. 输出处理
	if err := b.outputResults(results); err != nil {
		return fmt.Errorf("输出处理失败: %v", err)
	}

	// 5. 同步更新
	if b.configManager.Config.SyncToGame {
		if err := b.syncToGame(results); err != nil {
			return fmt.Errorf("同步到游戏目录失败: %v", err)
		}
	}

	// 6. 打印构建信息
	fmt.Printf("构建完成，耗时 %v，共处理 %d 个表，生成 %d 个文件\n",
		time.Since(startTime), len(sheets), len(results))

	return nil
}

// readSourceFiles 读取源文件
func (b *Builder) readSourceFiles() ([]*model.DataSheet, error) {
	allSheets := make([]*model.DataSheet, 0)

	// 遍历源文件目录
	err := filepath.WalkDir(b.configManager.Config.SourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// 检查文件扩展名
		reader := b.readerFactory.GetReader(path)
		if reader == nil {
			return nil // 跳过不支持的文件
		}

		// 快速模式：检查文件是否修改
		if b.configManager.Config.FastMode {
			if !b.needProcess(path) {
				fmt.Printf("跳过未修改文件: %s\n", path)
				return nil
			}
		}

		// 创建并初始化读取器
		r, err := b.readerFactory.CreateReader(path, b.configManager.Config.Readers["default"].Options)
		if err != nil {
			return err
		}

		// 读取文件
		fmt.Printf("读取文件: %s\n", path)
		sheets, err := r.ReadAll(path)
		if err != nil {
			return fmt.Errorf("读取 %s 失败: %v", path, err)
		}

		allSheets = append(allSheets, sheets...)
		return nil
	})

	if err != nil {
		return nil, err
	}

	// 应用合并配置
	allSheets = b.applyCombineConfig(allSheets)

	// 应用列替换配置
	allSheets = b.applyReplaceConfig(allSheets)

	return allSheets, nil
}

// needProcess 检查文件是否需要处理
func (b *Builder) needProcess(filePath string) bool {
	// 获取文件修改时间
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return true
	}

	fileModTime := fileInfo.ModTime()

	// 检查输出文件是否存在且修改时间晚于源文件
	for _, format := range b.configManager.Config.Formats {
		convConfig := b.configManager.GetConverterConfig(format)
		if convConfig == nil || !convConfig.Enabled {
			continue
		}

		// 构建输出路径
		outputDir := b.configManager.Config.OutputDir
		if convConfig.OutputPath != "" {
			outputDir = filepath.Join(outputDir, convConfig.OutputPath)
		}

		// 构建输出文件名
		fileName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
		var outputFileName string
		switch format {
		case "json":
			outputFileName = fmt.Sprintf("%s.json", fileName)
		case "php":
			outputFileName = fmt.Sprintf("%s.php", fileName)
		case "fbs":
			outputFileName = fmt.Sprintf("%s.bin", fileName)
		default:
			continue
		}

		outputPath := filepath.Join(outputDir, outputFileName)

		// 检查输出文件是否存在
		outputInfo, err := os.Stat(outputPath)
		if err != nil {
			return true // 输出文件不存在，需要处理
		}

		if outputInfo.ModTime().Before(fileModTime) {
			return true // 输出文件早于源文件，需要处理
		}
	}

	return false // 所有输出文件都存在且最新，不需要处理
}

// applyCombineConfig 应用合并配置
func (b *Builder) applyCombineConfig(sheets []*model.DataSheet) []*model.DataSheet {
	if b.configManager.CombineConfig == nil {
		return sheets
	}

	// 构建表名到表的映射
	sheetMap := make(map[string]*model.DataSheet)
	for _, sheet := range sheets {
		sheetMap[sheet.Name] = sheet
	}

	// 处理合并表
	combinedSheets := make([]*model.DataSheet, 0)
	processedSheets := make(map[string]bool)

	// 先添加非合并表
	for _, sheet := range sheets {
		processedSheets[sheet.Name] = false
	}

	// 处理合并配置
	for _, combineSheet := range b.configManager.CombineConfig.Sheets {
		// 检查所有源表是否存在
		allExists := true
		for _, sourceSheetName := range combineSheet.SourceSheets {
			if _, exists := sheetMap[sourceSheetName]; !exists {
				allExists = false
				break
			}
		}

		if !allExists {
			continue
		}

		// 创建合并表
		combinedSheet := &model.DataSheet{
			Name:    combineSheet.OutputName,
			Columns: []model.ColumnInfo{},
			Rows:    []map[string]interface{}{},
			Meta:    make(map[string]interface{}),
		}

		// 合并列信息（使用第一个表的列）
		if len(combineSheet.SourceSheets) > 0 {
			firstSheet := sheetMap[combineSheet.SourceSheets[0]]
			combinedSheet.Columns = firstSheet.Columns
		}

		// 合并行数据
		for _, sourceSheetName := range combineSheet.SourceSheets {
			sourceSheet := sheetMap[sourceSheetName]
			combinedSheet.Rows = append(combinedSheet.Rows, sourceSheet.Rows...)
			processedSheets[sourceSheetName] = true
		}

		combinedSheets = append(combinedSheets, combinedSheet)
	}

	// 添加未合并的表
	for _, sheet := range sheets {
		if !processedSheets[sheet.Name] {
			combinedSheets = append(combinedSheets, sheet)
		}
	}

	return combinedSheets
}

// applyReplaceConfig 应用列替换配置
func (b *Builder) applyReplaceConfig(sheets []*model.DataSheet) []*model.DataSheet {
	if b.configManager.ReplaceConfig == nil {
		return sheets
	}

	// 遍历每个表
	for _, sheet := range sheets {
		// 检查是否有替换规则
		replaceRules, exists := b.configManager.ReplaceConfig.Sheets[sheet.Name]
		if !exists {
			continue
		}

		// 遍历每行数据
		for _, row := range sheet.Rows {
			// 遍历每个替换规则
			for columnName, rule := range replaceRules.Columns {
				// 检查列是否存在
				if val, exists := row[columnName]; exists {
					// 替换值
					if strVal, ok := val.(string); ok {
						strVal = strings.ReplaceAll(strVal, rule.From, rule.To)
						row[columnName] = strVal
					}
				}
			}
		}
	}

	return sheets
}

// validateData 验证数据
func (b *Builder) validateData(sheets []*model.DataSheet) []*model.ErrorInfo {
	return b.validator.ValidateAll(sheets)
}

// convertData 转换数据
func (b *Builder) convertData(sheets []*model.DataSheet) ([]*model.ConvertResult, error) {
	results := make([]*model.ConvertResult, 0)

	// 异步处理
	if b.configManager.Config.Async {
		return b.asyncConvertData(sheets)
	}

	// 同步处理
	// 遍历每个格式
	for _, format := range b.configManager.Config.Formats {
		convConfig := b.configManager.GetConverterConfig(format)
		if convConfig == nil || !convConfig.Enabled {
			continue
		}

		// 创建并初始化转换器
		conv, err := b.converterFactory.CreateConverter(format, convConfig.Options)
		if err != nil {
			return nil, err
		}

		// 转换数据
		fmt.Printf("转换为 %s 格式\n", format)
		convResults, err := conv.BatchConvert(sheets)
		if err != nil {
			return nil, err
		}

		results = append(results, convResults...)
	}

	return results, nil
}

// asyncConvertData 异步转换数据
func (b *Builder) asyncConvertData(sheets []*model.DataSheet) ([]*model.ConvertResult, error) {
	results := make([]*model.ConvertResult, 0)
	resultChan := make(chan []*model.ConvertResult, len(b.configManager.Config.Formats))
	errChan := make(chan error, len(b.configManager.Config.Formats))

	// 遍历每个格式
	for _, format := range b.configManager.Config.Formats {
		go func(f string) {
			convConfig := b.configManager.GetConverterConfig(f)
			if convConfig == nil || !convConfig.Enabled {
				resultChan <- nil
				errChan <- nil
				return
			}

			// 创建并初始化转换器
			conv, err := b.converterFactory.CreateConverter(f, convConfig.Options)
			if err != nil {
				resultChan <- nil
				errChan <- err
				return
			}

			// 转换数据
			fmt.Printf("异步转换为 %s 格式\n", f)
			convResults, err := conv.BatchConvert(sheets)
			resultChan <- convResults
			errChan <- err
		}(format)
	}

	// 收集结果
	for i := 0; i < len(b.configManager.Config.Formats); i++ {
		convResults := <-resultChan
		err := <-errChan
		if err != nil {
			return nil, err
		}
		if convResults != nil {
			results = append(results, convResults...)
		}
	}

	return results, nil
}

// outputResults 输出结果
func (b *Builder) outputResults(results []*model.ConvertResult) error {
	// 遍历每个转换结果
	for _, result := range results {
		// 获取转换器配置
		convConfig := b.configManager.GetConverterConfig(result.Format)
		if convConfig == nil {
			continue
		}

		// 构建输出路径
		outputDir := b.configManager.Config.OutputDir
		if convConfig.OutputPath != "" {
			outputDir = filepath.Join(outputDir, convConfig.OutputPath)
		}

		// 创建输出目录
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("创建输出目录失败: %v", err)
		}

		// 构建输出文件路径
		outputPath := filepath.Join(outputDir, result.FileName)

		// 写入文件
		if err := os.WriteFile(outputPath, result.Content, 0644); err != nil {
			return fmt.Errorf("写入文件失败: %v", err)
		}

		fmt.Printf("生成文件: %s\n", outputPath)
	}

	return nil
}

// syncToGame 同步到游戏目录
func (b *Builder) syncToGame(results []*model.ConvertResult) error {
	if b.configManager.Config.GameDir == "" {
		return nil
	}

	// 遍历每个转换结果
	for _, result := range results {
		// 获取转换器配置
		convConfig := b.configManager.GetConverterConfig(result.Format)
		if convConfig == nil {
			continue
		}

		// 构建游戏目录下的输出路径
		gameOutputDir := b.configManager.Config.GameDir
		if convConfig.OutputPath != "" {
			gameOutputDir = filepath.Join(gameOutputDir, convConfig.OutputPath)
		}

		// 创建目录
		if err := os.MkdirAll(gameOutputDir, 0755); err != nil {
			return fmt.Errorf("创建游戏输出目录失败: %v", err)
		}

		// 构建输出文件路径
		outputPath := filepath.Join(gameOutputDir, result.FileName)

		// 写入文件
		if err := os.WriteFile(outputPath, result.Content, 0644); err != nil {
			return fmt.Errorf("写入游戏文件失败: %v", err)
		}

		fmt.Printf("同步到游戏目录: %s\n", outputPath)
	}

	return nil
}

func main() {
	// 解析命令行参数
	confDir := flag.String("conf", "./conf", "配置文件目录")
	fastMode := flag.Bool("fast", false, "快速模式，只处理修改过的文件")
	async := flag.Bool("async", false, "异步处理")
	help := flag.Bool("help", false, "显示帮助信息")
	flag.Parse()

	// 显示帮助信息
	if *help {
		fmt.Println("游戏数据构建工具")
		fmt.Println("Usage:")
		fmt.Println("  builder [options]")
		fmt.Println("Options:")
		fmt.Println("  -conf string   配置文件目录 (default \"./conf\")")
		fmt.Println("  -fast          快速模式，只处理修改过的文件")
		fmt.Println("  -async         异步处理")
		fmt.Println("  -help          显示帮助信息")
		return
	}

	// 创建构建器
	builder := NewBuilder()

	// 加载配置
	if err := builder.LoadConfig(*confDir); err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 覆盖配置
	if *fastMode {
		builder.configManager.Config.FastMode = true
	}
	if *async {
		builder.configManager.Config.Async = true
	}

	// 执行构建
	if err := builder.Build(); err != nil {
		fmt.Printf("构建失败: %v\n", err)
		os.Exit(1)
	}
}
