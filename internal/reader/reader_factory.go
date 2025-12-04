package reader

import (
	"path/filepath"
)

// ReaderFactory 读取器工厂
type ReaderFactory struct {
	readers map[string]IReader
}

// NewReaderFactory 创建读取器工厂
func NewReaderFactory() *ReaderFactory {
	factory := &ReaderFactory{
		readers: make(map[string]IReader),
	}

	// 注册默认读取器
	factory.RegisterReader(&CSVReader{})
	factory.RegisterReader(&ExcelReader{})

	return factory
}

// RegisterReader 注册读取器
func (f *ReaderFactory) RegisterReader(reader IReader) {
	for _, format := range reader.GetSupportedFormats() {
		f.readers[format] = reader
	}
}

// GetReader 根据文件扩展名获取读取器
func (f *ReaderFactory) GetReader(filePath string) IReader {
	ext := filepath.Ext(filePath)
	return f.readers[ext]
}

// CreateReader 创建并初始化读取器
func (f *ReaderFactory) CreateReader(filePath string, config map[string]interface{}) (IReader, error) {
	reader := f.GetReader(filePath)
	if reader == nil {
		return nil, nil
	}

	// 根据读取器类型创建新实例
	var newReader IReader
	switch reader.(type) {
	case *CSVReader:
		newReader = NewCSVReader()
	case *ExcelReader:
		newReader = NewExcelReader()
	default:
		return nil, nil
	}

	// 初始化读取器
	if err := newReader.Init(config); err != nil {
		return nil, err
	}

	return newReader, nil
}
