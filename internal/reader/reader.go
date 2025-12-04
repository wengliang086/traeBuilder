package reader

import (
	"github.com/game-data-builder/internal/model"
)

// IReader 定义了读取数据文件的接口
type IReader interface {
	// Init 初始化读取器
	Init(config map[string]interface{}) error
	
	// ReadAll 读取所有数据表
	ReadAll(filePath string) ([]*model.DataSheet, error)
	
	// ReadSheet 读取指定工作表
	ReadSheet(filePath string, sheetName string) (*model.DataSheet, error)
	
	// GetSupportedFormats 获取支持的文件格式
	GetSupportedFormats() []string
}
