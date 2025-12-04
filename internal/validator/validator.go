package validator

import (
	"github.com/game-data-builder/internal/model"
)

// IValidator 定义了数据验证的接口
type IValidator interface {
	// Init 初始化验证器
	Init(config map[string]interface{}) error

	// Validate 验证单个数据表
	Validate(sheet *model.DataSheet) []*model.ErrorInfo

	// ValidateAll 验证所有数据表
	ValidateAll(sheets []*model.DataSheet) []*model.ErrorInfo

	// ValidateRef 验证引用关系
	ValidateRef(sheets []*model.DataSheet) []*model.ErrorInfo
}
