package converter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/game-data-builder/internal/model"
)

// FBSConverter FlatBuffers转换器实现
type FBSConverter struct {
	config map[string]interface{}
}

// NewFBSConverter 创建FlatBuffers转换器
func NewFBSConverter() *FBSConverter {
	return &FBSConverter{}
}

// Init 初始化转换器
func (c *FBSConverter) Init(config map[string]interface{}) error {
	c.config = config
	return nil
}

// Convert 将数据转换为FlatBuffers格式
func (c *FBSConverter) Convert(sheet *model.DataSheet) (*model.ConvertResult, error) {
	// 构建FlatBuffers schema
	schema := c.buildSchema(sheet)

	// 构建JSON数据
	jsonData := c.buildJSONData(sheet)

	// 保存schema和JSON数据到临时文件
	tempDir := os.TempDir()
	schemaPath := filepath.Join(tempDir, fmt.Sprintf("%s.fbs", sheet.Name))
	jsonPath := filepath.Join(tempDir, fmt.Sprintf("%s.json", sheet.Name))
	outputPath := filepath.Join(tempDir, fmt.Sprintf("%s.bin", sheet.Name))

	// 写入schema文件
	if err := os.WriteFile(schemaPath, []byte(schema), 0644); err != nil {
		return nil, err
	}
	defer os.Remove(schemaPath)

	// 写入JSON文件
	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		return nil, err
	}
	defer os.Remove(jsonPath)

	// 检查flatc命令是否存在
	if _, err := exec.LookPath("flatc"); err != nil {
		// flatc命令不存在，返回schema和JSON数据
		result := &model.ConvertResult{
			FileName: fmt.Sprintf("%s.fbs", sheet.Name),
			Content:  []byte(schema),
			Format:   "fbs",
		}
		return result, nil
	}

	// 运行flatc命令生成二进制文件
	cmd := exec.Command("flatc", "-b", schemaPath, jsonPath)
	cmd.Dir = tempDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// 命令执行失败，返回schema和JSON数据
		result := &model.ConvertResult{
			FileName: fmt.Sprintf("%s.fbs", sheet.Name),
			Content:  []byte(schema),
			Format:   "fbs",
		}
		return result, nil
	}

	// 读取生成的二进制文件
	binContent, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, err
	}
	defer os.Remove(outputPath)

	// 创建转换结果
	result := &model.ConvertResult{
		FileName: fmt.Sprintf("%s.bin", sheet.Name),
		Content:  binContent,
		Format:   "fbs",
	}

	return result, nil
}

// GetFormat 获取支持的格式类型
func (c *FBSConverter) GetFormat() string {
	return "fbs"
}

// BatchConvert 批量转换多个数据表
func (c *FBSConverter) BatchConvert(sheets []*model.DataSheet) ([]*model.ConvertResult, error) {
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

// buildSchema 构建FlatBuffers schema
func (c *FBSConverter) buildSchema(sheet *model.DataSheet) string {
	var builder strings.Builder

	// 添加文件头
	builder.WriteString(fmt.Sprintf("// 自动生成的 %s 数据schema\n\n", sheet.Name))

	// 定义数据结构
	builder.WriteString(fmt.Sprintf("namespace %s;\n\n", sheet.Name))

	// 定义列类型枚举
	builder.WriteString("enum ColumnType : byte {\n")
	builder.WriteString("    INT,\n")
	builder.WriteString("    FLOAT,\n")
	builder.WriteString("    BOOL,\n")
	builder.WriteString("    STRING,\n")
	builder.WriteString("}\n\n")

	// 定义列信息结构
	builder.WriteString("table ColumnInfo {\n")
	builder.WriteString("    name:string;\n")
	builder.WriteString("    type:ColumnType;\n")
	builder.WriteString("    comment:string;\n")
	builder.WriteString("    required:bool = true;\n")
	builder.WriteString("    default:string;\n")
	builder.WriteString("    options:[string];\n")
	builder.WriteString("}\n\n")

	// 定义行数据结构
	builder.WriteString(fmt.Sprintf("table RowData_%s {\n", sheet.Name))
	for _, col := range sheet.Columns {
		fbsType := c.getFBSType(col.Type)
		builder.WriteString(fmt.Sprintf("    %s:%s;\n", col.Name, fbsType))
	}
	builder.WriteString("}\n\n")

	// 定义数据表结构
	builder.WriteString(fmt.Sprintf("table Data_%s {\n", sheet.Name))
	builder.WriteString("    name:string;\n")
	builder.WriteString("    columns:[ColumnInfo];\n")
	builder.WriteString(fmt.Sprintf("    rows:[RowData_%s];\n", sheet.Name))
	builder.WriteString("    meta:[string];\n")
	builder.WriteString("}\n\n")

	// 定义根类型
	builder.WriteString(fmt.Sprintf("root_type Data_%s;\n", sheet.Name))

	return builder.String()
}

// buildJSONData 构建JSON数据
func (c *FBSConverter) buildJSONData(sheet *model.DataSheet) []byte {
	// 转换数据
	data := make(map[string]interface{})
	data["name"] = sheet.Name

	// 转换列信息
	columns := make([]map[string]interface{}, 0)
	for _, col := range sheet.Columns {
		colData := make(map[string]interface{})
		colData["name"] = col.Name
		colData["type"] = c.getColumnTypeValue(col.Type)
		colData["comment"] = col.Comment
		colData["required"] = col.Required
		if col.Default != nil {
			colData["default"] = fmt.Sprintf("%v", col.Default)
		}
		colData["options"] = col.Options
		columns = append(columns, colData)
	}
	data["columns"] = columns

	// 转换行数据
	rows := make([]map[string]interface{}, 0)
	for _, row := range sheet.Rows {
		rowData := make(map[string]interface{})
		for _, col := range sheet.Columns {
			if val, exists := row[col.Name]; exists {
				rowData[col.Name] = val
			}
		}
		rows = append(rows, rowData)
	}
	data["rows"] = rows

	// 转换元数据
	meta := make([]string, 0)
	for key, val := range sheet.Meta {
		meta = append(meta, fmt.Sprintf("%s:%v", key, val))
	}
	data["meta"] = meta

	// 格式化JSON
	content, _ := json.MarshalIndent(data, "", "  ")
	return content
}

// getFBSType 获取FlatBuffers类型
func (c *FBSConverter) getFBSType(colType string) string {
	switch colType {
	case "int", "integer":
		return "int32"
	case "float", "double", "number":
		return "float64"
	case "bool", "boolean":
		return "bool"
	case "string":
		return "string"
	default:
		return "string"
	}
}

// getColumnTypeValue 获取列类型枚举值
func (c *FBSConverter) getColumnTypeValue(colType string) int {
	switch colType {
	case "int", "integer":
		return 0
	case "float", "double", "number":
		return 1
	case "bool", "boolean":
		return 2
	case "string":
		return 3
	default:
		return 3
	}
}
