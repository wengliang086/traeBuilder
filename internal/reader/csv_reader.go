package reader

import (
	"encoding/csv"
	"os"
	"strconv"
	"strings"

	"github.com/game-data-builder/internal/model"
)

// CSVReader CSV读取器实现
type CSVReader struct {
	config map[string]interface{}
}

// NewCSVReader 创建CSV读取器
func NewCSVReader() *CSVReader {
	return &CSVReader{}
}

// Init 初始化读取器
func (r *CSVReader) Init(config map[string]interface{}) error {
	r.config = config
	return nil
}

// ReadAll 读取所有数据表
func (r *CSVReader) ReadAll(filePath string) ([]*model.DataSheet, error) {
	// CSV文件只有一个工作表
	sheet, err := r.ReadSheet(filePath, "")
	if err != nil {
		return nil, err
	}
	return []*model.DataSheet{sheet}, nil
}

// ReadSheet 读取指定工作表
func (r *CSVReader) ReadSheet(filePath string, sheetName string) (*model.DataSheet, error) {
	// 打开CSV文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 创建CSV阅读器
	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true

	// 读取所有行
	allLines, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(allLines) < 3 { // 至少需要表头、类型、注释行
		return nil, nil
	}

	// 解析列信息
	columns := make([]model.ColumnInfo, 0)
	headerRow := allLines[0]
	typeRow := allLines[1]
	commentRow := allLines[2]

	for i, name := range headerRow {
		if name == "" {
			continue // 跳过空列
		}

		colInfo := model.ColumnInfo{
			Name:    name,
			Comment: commentRow[i],
			Required: true,
		}

		// 解析类型
		colType := typeRow[i]
		colInfo.Type = colType

		// 解析注释中的元数据
		colInfo = r.parseCommentMetadata(colInfo, commentRow[i])

		columns = append(columns, colInfo)
	}

	// 解析数据行
	rows := make([]map[string]interface{}, 0)
	for rowIndex := 3; rowIndex < len(allLines); rowIndex++ {
		line := allLines[rowIndex]
		if len(line) == 0 || line[0] == "" {
			continue // 跳过空行
		}

		rowData := make(map[string]interface{})
		for i, col := range columns {
			if i >= len(line) {
				rowData[col.Name] = col.Default
				continue
			}

			value := line[i]
			if value == "" {
				rowData[col.Name] = col.Default
				continue
			}

			// 转换数据类型
			convertedValue, err := r.convertValue(value, col.Type)
			if err != nil {
				return nil, err
			}
			rowData[col.Name] = convertedValue
		}
		rows = append(rows, rowData)
	}

	// 获取文件名作为表名
	tableName := ""
	parts := strings.Split(filePath, "/")
	if len(parts) > 0 {
		tableName = parts[len(parts)-1]
	}
	// 移除后缀
	tableName = strings.TrimSuffix(tableName, ".csv")
	tableName = strings.TrimSuffix(tableName, ".CSV")

	// 创建数据表
	sheet := &model.DataSheet{
		Name:    tableName,
		Columns: columns,
		Rows:    rows,
		Meta:    make(map[string]interface{}),
	}

	return sheet, nil
}

// GetSupportedFormats 获取支持的文件格式
func (r *CSVReader) GetSupportedFormats() []string {
	return []string{".csv", ".CSV"}
}

// parseCommentMetadata 解析注释中的元数据
func (r *CSVReader) parseCommentMetadata(col model.ColumnInfo, comment string) model.ColumnInfo {
	// 示例注释格式："必填|默认:0|选项:a,b,c|引用:table.column"
	parts := strings.Split(comment, "|")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "必填") {
			col.Required = true
		} else if strings.HasPrefix(part, "选填") {
			col.Required = false
		} else if strings.HasPrefix(part, "默认:") {
			defaultVal := strings.TrimPrefix(part, "默认:")
			val, _ := r.convertValue(defaultVal, col.Type)
			col.Default = val
		} else if strings.HasPrefix(part, "选项:") {
			optionsStr := strings.TrimPrefix(part, "选项:")
			col.Options = strings.Split(optionsStr, ",")
		} else if strings.HasPrefix(part, "引用:") {
			refStr := strings.TrimPrefix(part, "引用:")
			refParts := strings.Split(refStr, ".")
			if len(refParts) == 2 {
				col.Ref = &model.RefInfo{
					Sheet:  refParts[0],
					Column: refParts[1],
				}
			}
		}
	}
	return col
}

// convertValue 转换数据类型
func (r *CSVReader) convertValue(value string, dataType string) (interface{}, error) {
	switch dataType {
	case "int", "integer":
		return strconv.Atoi(value)
	case "float", "double", "number":
		return strconv.ParseFloat(value, 64)
	case "bool", "boolean":
		return strconv.ParseBool(value)
	case "string":
		return value, nil
	default:
		return value, nil
	}
}
