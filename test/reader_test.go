package test

import (
	"testing"

	"github.com/game-data-builder/internal/reader"
)

// TestReaderFactory 测试读取器工厂
func TestReaderFactory(t *testing.T) {
	factory := reader.NewReaderFactory()

	// 测试获取CSV读取器
	csvReader := factory.GetReader("test.csv")
	if csvReader == nil {
		t.Error("Expected CSV reader, got nil")
	}

	// 测试获取Excel读取器
	excelReader := factory.GetReader("test.xlsx")
	if excelReader == nil {
		t.Error("Expected Excel reader, got nil")
	}

	// 测试获取不支持的读取器
	invalidReader := factory.GetReader("test.txt")
	if invalidReader != nil {
		t.Error("Expected nil for invalid file type, got reader")
	}
}

// TestCSVReaderFormats 测试CSV读取器支持的格式
func TestCSVReaderFormats(t *testing.T) {
	csvReader := reader.NewCSVReader()
	formats := csvReader.GetSupportedFormats()

	if len(formats) != 2 {
		t.Errorf("Expected 2 formats, got %d", len(formats))
	}

	expectedFormats := map[string]bool{
		".csv": true,
		".CSV": true,
	}

	for _, format := range formats {
		if !expectedFormats[format] {
			t.Errorf("Unexpected format: %s", format)
		}
	}
}

// TestExcelReaderFormats 测试Excel读取器支持的格式
func TestExcelReaderFormats(t *testing.T) {
	excelReader := reader.NewExcelReader()
	formats := excelReader.GetSupportedFormats()

	if len(formats) != 4 {
		t.Errorf("Expected 4 formats, got %d", len(formats))
	}

	expectedFormats := map[string]bool{
		".xlsx": true,
		".xlsm": true,
		".xltx": true,
		".xltm": true,
	}

	for _, format := range formats {
		if !expectedFormats[format] {
			t.Errorf("Unexpected format: %s", format)
		}
	}
}
