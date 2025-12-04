package reader

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/game-data-builder/internal/model"
	"github.com/xuri/excelize/v2"
)

// ExcelReader Excel读取器实现
type ExcelReader struct {
	config map[string]interface{}
}

// NewExcelReader 创建Excel读取器
func NewExcelReader() *ExcelReader {
	return &ExcelReader{}
}

// Init 初始化读取器
func (r *ExcelReader) Init(config map[string]interface{}) error {
	r.config = config
	return nil
}

// ReadAll 读取所有数据表
func (r *ExcelReader) ReadAll(filePath string) ([]*model.DataSheet, error) {
	// 打开Excel文件
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// 获取所有工作表名
	sheetNames := f.GetSheetList()
	sheets := make([]*model.DataSheet, 0)

	// 读取每个工作表
	for _, sheetName := range sheetNames {
		// 跳过以_开头的工作表（隐藏表）
		if strings.HasPrefix(sheetName, "_") {
			continue
		}

		sheet, err := r.readSheet(f, sheetName)
		if err != nil {
			return nil, err
		}
		if sheet != nil {
			sheets = append(sheets, sheet)
		}
	}

	return sheets, nil
}

// ReadSheet 读取指定工作表
func (r *ExcelReader) ReadSheet(filePath string, sheetName string) (*model.DataSheet, error) {
	// 打开Excel文件
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// 如果未指定工作表名，使用第一个工作表
	if sheetName == "" {
		sheetNames := f.GetSheetList()
		if len(sheetNames) == 0 {
			return nil, nil
		}
		sheetName = sheetNames[0]
	}

	return r.readSheet(f, sheetName)
}

// readSheet 读取单个工作表
func (r *ExcelReader) readSheet(f *excelize.File, sheetName string) (*model.DataSheet, error) {
	// 获取工作表的所有行
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, err
	}

	if len(rows) < 3 { // 至少需要表头、类型、注释行
		return nil, nil
	}

	// 解析列信息
	columns := make([]model.ColumnInfo, 0)
	headerRow := rows[0]
	typeRow := rows[1]
	commentRow := rows[2]

	for i, name := range headerRow {
		if name == "" {
			continue // 跳过空列
		}

		colInfo := model.ColumnInfo{
			Name:     name,
			Comment:  commentRow[i],
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
	dataRows := make([]map[string]interface{}, 0)
	for rowIndex := 3; rowIndex < len(rows); rowIndex++ {
		row := rows[rowIndex]
		if len(row) == 0 || row[0] == "" {
			continue // 跳过空行
		}

		rowData := make(map[string]interface{})
		for i, col := range columns {
			var cellValue string
			if i < len(row) {
				cellValue = row[i]
			}

			if cellValue == "" {
				rowData[col.Name] = col.Default
				continue
			}

			// 转换数据类型
			convertedValue, err := r.convertValue(cellValue, col.Type)
			if err != nil {
				return nil, fmt.Errorf("sheet %s, row %d, column %s: %v", sheetName, rowIndex+1, col.Name, err)
			}
			rowData[col.Name] = convertedValue
		}
		dataRows = append(dataRows, rowData)
	}

	// 创建数据表
	sheet := &model.DataSheet{
		Name:    sheetName,
		Columns: columns,
		Rows:    dataRows,
		Meta:    make(map[string]interface{}),
	}

	return sheet, nil
}

// GetSupportedFormats 获取支持的文件格式
func (r *ExcelReader) GetSupportedFormats() []string {
	return []string{".xlsx", ".xlsm", ".xltx", ".xltm"}
}

// parseCommentMetadata 解析注释中的元数据
func (r *ExcelReader) parseCommentMetadata(col model.ColumnInfo, comment string) model.ColumnInfo {
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
func (r *ExcelReader) convertValue(value string, dataType string) (interface{}, error) {
	// 这是一个简化的实现，实际项目中可能需要更复杂的类型转换
	switch strings.ToLower(dataType) {
	case "int", "integer":
		return strconv.Atoi(value)
	case "float", "double", "number":
		return strconv.ParseFloat(value, 64)
	case "bool", "boolean":
		value = strings.ToLower(value)
		if value == "true" || value == "1" || value == "yes" {
			return true, nil
		}
		return false, nil
	case "string":
		return value, nil
	default:
		return value, nil
	}
}
