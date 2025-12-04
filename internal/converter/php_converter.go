package converter

import (
	"fmt"
	"strings"

	"github.com/game-data-builder/internal/model"
)

// PHPConverter PHP转换器实现
type PHPConverter struct {
	config map[string]interface{}
}

// NewPHPConverter 创建PHP转换器
func NewPHPConverter() *PHPConverter {
	return &PHPConverter{}
}

// Init 初始化转换器
func (c *PHPConverter) Init(config map[string]interface{}) error {
	c.config = config
	return nil
}

// Convert 将数据转换为PHP格式
func (c *PHPConverter) Convert(sheet *model.DataSheet) (*model.ConvertResult, error) {
	// 构建PHP数组字符串
	var builder strings.Builder

	// 添加文件头注释
	builder.WriteString("<?php\n")
	builder.WriteString(fmt.Sprintf("// 自动生成的 %s 数据文件\n", sheet.Name))
	builder.WriteString(fmt.Sprintf("// 表名: %s\n\n", sheet.Name))

	// 开始数组
	builder.WriteString(fmt.Sprintf("return [\n"))

	// 添加元数据
	builder.WriteString(fmt.Sprintf("    'name' => '%s',\n", sheet.Name))

	// 添加列信息
	builder.WriteString("    'columns' => [\n")
	for i, col := range sheet.Columns {
		builder.WriteString(fmt.Sprintf("        %d => [\n", i))
		builder.WriteString(fmt.Sprintf("            'name' => '%s',\n", col.Name))
		builder.WriteString(fmt.Sprintf("            'type' => '%s',\n", col.Type))
		builder.WriteString(fmt.Sprintf("            'comment' => '%s',\n", col.Comment))
		builder.WriteString(fmt.Sprintf("            'required' => %s,\n", c.boolToString(col.Required)))

		// 默认值
		if col.Default != nil {
			builder.WriteString(fmt.Sprintf("            'default' => %s,\n", c.valueToString(col.Default)))
		} else {
			builder.WriteString("            'default' => null,\n")
		}

		// 选项
		if len(col.Options) > 0 {
			builder.WriteString("            'options' => [\n")
			for j, opt := range col.Options {
				builder.WriteString(fmt.Sprintf("                %d => '%s',\n", j, opt))
			}
			builder.WriteString("            ],\n")
		} else {
			builder.WriteString("            'options' => [],\n")
		}

		// 引用信息
		if col.Ref != nil {
			builder.WriteString(fmt.Sprintf("            'ref' => ['sheet' => '%s', 'column' => '%s'],\n", col.Ref.Sheet, col.Ref.Column))
		} else {
			builder.WriteString("            'ref' => null,\n")
		}

		builder.WriteString("        ],\n")
	}
	builder.WriteString("    ],\n")

	// 添加行数据
	builder.WriteString("    'rows' => [\n")
	for i, row := range sheet.Rows {
		builder.WriteString(fmt.Sprintf("        %d => [\n", i))
		for _, col := range sheet.Columns {
			if val, exists := row[col.Name]; exists {
				builder.WriteString(fmt.Sprintf("            '%s' => %s,\n", col.Name, c.valueToString(val)))
			} else {
				builder.WriteString(fmt.Sprintf("            '%s' => null,\n", col.Name))
			}
		}
		builder.WriteString("        ],\n")
	}
	builder.WriteString("    ],\n")

	// 添加元数据
	builder.WriteString("    'meta' => [\n")
	for key, val := range sheet.Meta {
		builder.WriteString(fmt.Sprintf("        '%s' => %s,\n", key, c.valueToString(val)))
	}
	builder.WriteString("    ],\n")

	// 结束数组
	builder.WriteString("];\n")

	// 创建转换结果
	result := &model.ConvertResult{
		FileName: fmt.Sprintf("%s.php", sheet.Name),
		Content:  []byte(builder.String()),
		Format:   "php",
	}

	return result, nil
}

// GetFormat 获取支持的格式类型
func (c *PHPConverter) GetFormat() string {
	return "php"
}

// BatchConvert 批量转换多个数据表
func (c *PHPConverter) BatchConvert(sheets []*model.DataSheet) ([]*model.ConvertResult, error) {
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

// valueToString 将值转换为PHP字符串
func (c *PHPConverter) valueToString(val interface{}) string {
	switch v := val.(type) {
	case string:
		return fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "\\'"))
	case bool:
		return c.boolToString(v)
	case int, int32, int64, float32, float64:
		return fmt.Sprintf("%v", v)
	default:
		return "null"
	}
}

// boolToString 将布尔值转换为PHP字符串
func (c *PHPConverter) boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
