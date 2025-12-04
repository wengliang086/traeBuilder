package converter

import (
	"github.com/game-data-builder/internal/model"
)

// IConverter 定义了数据转换的接口
type IConverter interface {
	// Init 初始化转换器
	Init(config map[string]interface{}) error

	// Convert 将数据转换为目标格式
	Convert(sheet *model.DataSheet) (*model.ConvertResult, error)

	// GetFormat 获取支持的格式类型
	GetFormat() string

	// BatchConvert 批量转换多个数据表
	BatchConvert(sheets []*model.DataSheet) ([]*model.ConvertResult, error)
}
