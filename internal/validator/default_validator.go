package validator

import (
	"fmt"
	"reflect"

	"github.com/game-data-builder/internal/model"
)

// DefaultValidator 默认验证器实现
type DefaultValidator struct {
	config map[string]interface{}
}

// NewDefaultValidator 创建默认验证器
func NewDefaultValidator() *DefaultValidator {
	return &DefaultValidator{}
}

// Init 初始化验证器
func (v *DefaultValidator) Init(config map[string]interface{}) error {
	v.config = config
	return nil
}

// Validate 验证单个数据表
func (v *DefaultValidator) Validate(sheet *model.DataSheet) []*model.ErrorInfo {
	errors := make([]*model.ErrorInfo, 0)

	// 验证每行数据
	for rowIndex, row := range sheet.Rows {
		// 验证必填字段
		for _, col := range sheet.Columns {
			if col.Required {
				if _, exists := row[col.Name]; !exists || row[col.Name] == nil || row[col.Name] == "" {
					errors = append(errors, &model.ErrorInfo{
						Sheet:  sheet.Name,
						Row:    rowIndex + 4, // 数据行从第4行开始
						Column: col.Name,
						Msg:    fmt.Sprintf("必填字段不能为空"),
					})
				}
			}

			// 验证数据类型
			if val, exists := row[col.Name]; exists && val != nil && val != "" {
				if !v.validateDataType(val, col.Type) {
					errors = append(errors, &model.ErrorInfo{
						Sheet:  sheet.Name,
						Row:    rowIndex + 4,
						Column: col.Name,
						Msg:    fmt.Sprintf("数据类型错误，期望 %s，实际 %T", col.Type, val),
					})
				}
			}

			// 验证枚举值
			if len(col.Options) > 0 {
				if val, exists := row[col.Name]; exists && val != nil {
					valStr, ok := val.(string)
					if !ok {
						continue // 非字符串类型跳过枚举验证
					}

					valid := false
					for _, opt := range col.Options {
						if opt == valStr {
							valid = true
							break
						}
					}

					if !valid {
						errors = append(errors, &model.ErrorInfo{
							Sheet:  sheet.Name,
							Row:    rowIndex + 4,
							Column: col.Name,
							Msg:    fmt.Sprintf("值不在可选范围内，可选值: %v", col.Options),
						})
					}
				}
			}
		}
	}

	return errors
}

// ValidateAll 验证所有数据表
func (v *DefaultValidator) ValidateAll(sheets []*model.DataSheet) []*model.ErrorInfo {
	errors := make([]*model.ErrorInfo, 0)

	// 验证每个表
	for _, sheet := range sheets {
		sheetErrors := v.Validate(sheet)
		errors = append(errors, sheetErrors...)
	}

	// 验证引用关系
	refErrors := v.ValidateRef(sheets)
	errors = append(errors, refErrors...)

	return errors
}

// ValidateRef 验证引用关系
func (v *DefaultValidator) ValidateRef(sheets []*model.DataSheet) []*model.ErrorInfo {
	errors := make([]*model.ErrorInfo, 0)

	// 构建引用索引
	refIndex := make(map[string]map[interface{}]bool)
	for _, sheet := range sheets {
		refIndex[sheet.Name] = make(map[interface{}]bool)
		for _, row := range sheet.Rows {
			// 默认使用第一列作为主键
			if len(sheet.Columns) > 0 {
				primaryKey := sheet.Columns[0].Name
				if val, exists := row[primaryKey]; exists && val != nil {
					refIndex[sheet.Name][val] = true
				}
			}
		}
	}

	// 验证每个表的引用关系
	for _, sheet := range sheets {
		for _, col := range sheet.Columns {
			if col.Ref != nil {
				// 检查引用的表是否存在
				if _, exists := refIndex[col.Ref.Sheet]; !exists {
					errors = append(errors, &model.ErrorInfo{
						Sheet:  sheet.Name,
						Column: col.Name,
						Msg:    fmt.Sprintf("引用的表 %s 不存在", col.Ref.Sheet),
					})
					continue
				}

				// 验证每行数据的引用值
				for rowIndex, row := range sheet.Rows {
					if val, exists := row[col.Name]; exists && val != nil {
						if !refIndex[col.Ref.Sheet][val] {
						errors = append(errors, &model.ErrorInfo{
							Sheet:  sheet.Name,
							Row:    rowIndex + 4,
							Column: col.Name,
							Msg:    fmt.Sprintf("引用值 %v 在表 %s 中不存在", val, col.Ref.Sheet),
						})
						}
					}
				}
			}
		}
	}

	return errors
}

// validateDataType 验证数据类型
func (v *DefaultValidator) validateDataType(value interface{}, expectedType string) bool {
	valType := reflect.TypeOf(value).String()
	switch expectedType {
	case "int", "integer":
		return valType == "int" || valType == "int32" || valType == "int64" || valType == "float64" // 允许数字类型
	case "float", "double", "number":
		return valType == "float32" || valType == "float64"
	case "bool", "boolean":
		return valType == "bool"
	case "string":
		return valType == "string"
	default:
		return true // 未知类型默认通过
	}
}
