package converter

import (
	"encoding/json"
	"fmt"

	"github.com/game-data-builder/internal/model"
)

// JSONConverter JSON转换器实现
type JSONConverter struct {
	config map[string]interface{}
}

// NewJSONConverter 创建JSON转换器
func NewJSONConverter() *JSONConverter {
	return &JSONConverter{}
}

// Init 初始化转换器
func (c *JSONConverter) Init(config map[string]interface{}) error {
	c.config = config
	return nil
}

// Convert 将数据转换为JSON格式
func (c *JSONConverter) Convert(sheet *model.DataSheet) (*model.ConvertResult, error) {
	// 转换数据
	data := make(map[string]interface{})
	data["name"] = sheet.Name
	data["columns"] = sheet.Columns
	data["rows"] = sheet.Rows
	data["meta"] = sheet.Meta

	// 格式化JSON
	var content []byte
	var err error

	// 检查是否需要格式化输出
	if indent, ok := c.config["indent"].(bool); ok && indent {
		content, err = json.MarshalIndent(data, "", "  ")
	} else {
		content, err = json.Marshal(data)
	}

	if err != nil {
		return nil, err
	}

	// 创建转换结果
	result := &model.ConvertResult{
		FileName: fmt.Sprintf("%s.json", sheet.Name),
		Content:  content,
		Format:   "json",
	}

	return result, nil
}

// GetFormat 获取支持的格式类型
func (c *JSONConverter) GetFormat() string {
	return "json"
}

// BatchConvert 批量转换多个数据表
func (c *JSONConverter) BatchConvert(sheets []*model.DataSheet) ([]*model.ConvertResult, error) {
	results := make([]*model.ConvertResult, 0)

	for _, sheet := range sheets {
		result, err := c.Convert(sheet)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}
