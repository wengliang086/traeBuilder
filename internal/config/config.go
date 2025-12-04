package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config 主配置结构
type Config struct {
	SourceDir     string            `json:"sourceDir"`      // 源文件目录
	OutputDir     string            `json:"outputDir"`      // 输出目录
	Formats       []string          `json:"formats"`        // 转换格式
	Async         bool              `json:"async"`          // 是否异步处理
	FastMode      bool              `json:"fastMode"`       // 快速模式
	SyncToGame    bool              `json:"syncToGame"`     // 是否同步到游戏目录
	GameDir       string            `json:"gameDir"`        // 游戏目录
	Readers       map[string]ReaderConfig `json:"readers"`    // 读取器配置
	Converters    map[string]ConverterConfig `json:"converters"` // 转换器配置
	Validators    map[string]ValidatorConfig `json:"validators"` // 验证器配置
}

// ReaderConfig 读取器配置
type ReaderConfig struct {
	Type       string                 `json:"type"`         // 读取器类型
	Enabled    bool                   `json:"enabled"`      // 是否启用
	Options    map[string]interface{} `json:"options"`      // 选项
}

// ConverterConfig 转换器配置
type ConverterConfig struct {
	Type       string                 `json:"type"`         // 转换器类型
	Enabled    bool                   `json:"enabled"`      // 是否启用
	OutputPath string                 `json:"outputPath"`   // 输出路径
	Options    map[string]interface{} `json:"options"`      // 选项
}

// ValidatorConfig 验证器配置
type ValidatorConfig struct {
	Type       string                 `json:"type"`         // 验证器类型
	Enabled    bool                   `json:"enabled"`      // 是否启用
	Options    map[string]interface{} `json:"options"`      // 选项
}

// CombineConfig 合并配置
type CombineConfig struct {
	Sheets map[string]CombineSheet `json:"sheets"` // 合并表配置
}

// CombineSheet 合并表配置
type CombineSheet struct {
	SourceSheets []string `json:"sourceSheets"` // 源表列表
	KeyColumn    string   `json:"keyColumn"`    // 主键列
	OutputName   string   `json:"outputName"`   // 输出表名
}

// ReplaceColumnConfig 列替换配置
type ReplaceColumnConfig struct {
	Sheets map[string]ReplaceRules `json:"sheets"` // 表替换规则
}

// ReplaceRules 替换规则
type ReplaceRules struct {
	Columns map[string]ReplaceRule `json:"columns"` // 列替换规则
}

// ReplaceRule 替换规则
type ReplaceRule struct {
	From string `json:"from"` // 原内容
	To   string `json:"to"`   // 替换内容
}

// ConfigManager 配置管理器
type ConfigManager struct {
	Config          *Config
	CombineConfig   *CombineConfig
	ReplaceConfig   *ReplaceColumnConfig
}

// NewConfigManager 创建配置管理器
func NewConfigManager() *ConfigManager {
	return &ConfigManager{}
}

// Load 加载所有配置文件
func (cm *ConfigManager) Load(confDir string) error {
	// 加载主配置
	if err := cm.loadMainConfig(confDir); err != nil {
		return err
	}

	// 加载合并配置
	if err := cm.loadCombineConfig(confDir); err != nil {
		return err
	}

	// 加载列替换配置
	if err := cm.loadReplaceConfig(confDir); err != nil {
		return err
	}

	return nil
}

// loadMainConfig 加载主配置
func (cm *ConfigManager) loadMainConfig(confDir string) error {
	path := filepath.Join(confDir, "config.json")
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var config Config
	if err := json.Unmarshal(content, &config); err != nil {
		return err
	}

	cm.Config = &config
	return nil
}

// loadCombineConfig 加载合并配置
func (cm *ConfigManager) loadCombineConfig(confDir string) error {
	path := filepath.Join(confDir, "combine.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// 配置文件不存在，使用默认值
		cm.CombineConfig = &CombineConfig{Sheets: make(map[string]CombineSheet)}
		return nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var combineConfig CombineConfig
	if err := json.Unmarshal(content, &combineConfig); err != nil {
		return err
	}

	cm.CombineConfig = &combineConfig
	return nil
}

// loadReplaceConfig 加载列替换配置
func (cm *ConfigManager) loadReplaceConfig(confDir string) error {
	path := filepath.Join(confDir, "replaceColumn.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// 配置文件不存在，使用默认值
		cm.ReplaceConfig = &ReplaceColumnConfig{Sheets: make(map[string]ReplaceRules)}
		return nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var replaceConfig ReplaceColumnConfig
	if err := json.Unmarshal(content, &replaceConfig); err != nil {
		return err
	}

	cm.ReplaceConfig = &replaceConfig
	return nil
}

// GetReaderConfig 获取读取器配置
func (cm *ConfigManager) GetReaderConfig(readerType string) *ReaderConfig {
	if cm.Config == nil || cm.Config.Readers == nil {
		return nil
	}
	cfg := cm.Config.Readers[readerType]
	return &cfg
}

// GetConverterConfig 获取转换器配置
func (cm *ConfigManager) GetConverterConfig(format string) *ConverterConfig {
	if cm.Config == nil || cm.Config.Converters == nil {
		return nil
	}
	cfg := cm.Config.Converters[format]
	return &cfg
}

// GetValidatorConfig 获取验证器配置
func (cm *ConfigManager) GetValidatorConfig(validatorType string) *ValidatorConfig {
	if cm.Config == nil || cm.Config.Validators == nil {
		return nil
	}
	cfg := cm.Config.Validators[validatorType]
	return &cfg
}
