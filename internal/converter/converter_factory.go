package converter

// ConverterFactory 转换器工厂
type ConverterFactory struct {
	converters map[string]IConverter
}

// NewConverterFactory 创建转换器工厂
func NewConverterFactory() *ConverterFactory {
	factory := &ConverterFactory{
		converters: make(map[string]IConverter),
	}

	// 注册默认转换器
	factory.RegisterConverter(&JSONConverter{})
	factory.RegisterConverter(&PHPConverter{})
	factory.RegisterConverter(&FBSConverter{})

	return factory
}

// RegisterConverter 注册转换器
func (f *ConverterFactory) RegisterConverter(converter IConverter) {
	f.converters[converter.GetFormat()] = converter
}

// GetConverter 根据格式获取转换器
func (f *ConverterFactory) GetConverter(format string) IConverter {
	return f.converters[format]
}

// CreateConverter 创建并初始化转换器
func (f *ConverterFactory) CreateConverter(format string, config map[string]interface{}) (IConverter, error) {
	converter := f.GetConverter(format)
	if converter == nil {
		return nil, nil
	}

	// 根据转换器类型创建新实例
	var newConverter IConverter
	switch converter.(type) {
	case *JSONConverter:
		newConverter = NewJSONConverter()
	case *PHPConverter:
		newConverter = NewPHPConverter()
	case *FBSConverter:
		newConverter = NewFBSConverter()
	default:
		return nil, nil
	}

	// 初始化转换器
	if err := newConverter.Init(config); err != nil {
		return nil, err
	}

	return newConverter, nil
}
