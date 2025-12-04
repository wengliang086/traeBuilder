# 游戏数据构建工具

一个基于 Go 语言开发的游戏数据构建工具，用于将游戏策划配置的 Excel 或 CSV 文件转换为游戏客户端和服务器可用的多种数据格式。

## 功能特性

- **数据转换**：支持从 Excel 或 CSV 文件读取数据，并转换为游戏所需的数据格式。
- **多格式输出**：能够生成 PHP、JSON 和 FlatBuffers 等不同格式的数据文件。
- **性能优化**：
  - 异步处理机制，提高转换速度。
  - 快速模式功能，仅处理修改过的文件，提高开发效率。
- **数据完整性**：提供内置的数据验证和引用关系检查，确保数据的准确性。
- **灵活配置**：通过 JSON 配置文件进行配置管理，允许自定义数据转换规则和列映射。
- **模块化设计**：包括配置处理、数据模型、读写器、转换器和验证器接口的实现。

## 安装

### 环境要求

- Go 1.24.0 或更高版本

### 安装步骤

1. 克隆项目到本地：
   ```bash
   git clone https://github.com/game-data-builder.git
   cd game-data-builder
   ```

2. 安装依赖：
   ```bash
   go mod tidy
   ```

3. 编译项目：
   ```bash
   go build -o builder cmd/main.go
   ```

## 使用方法

### 配置文件

配置文件位于 `conf/` 目录：

- `config.json`：主配置文件，定义源目录、输出目录、转换格式等。
- `combine.json`：表合并配置，定义如何合并多个表。
- `replaceColumn.json`：列替换配置，定义如何替换列值。

### 运行工具

```bash
./builder [options]
```

#### 可选参数

- `-conf string`：配置文件目录 (默认 "./conf")
- `-fast`：快速模式，只处理修改过的文件
- `-async`：异步处理，并发转换数据
- `-help`：显示帮助信息

### 示例

使用默认配置运行：
```bash
./builder
```

使用快速模式运行：
```bash
./builder -fast
```

使用异步处理运行：
```bash
./builder -async
```

## 工作流程

1. **初始化**：加载配置文件，设置转换参数。
2. **数据读取**：通过 IReader 接口读取 Excel 或 CSV 源文件。
3. **数据验证**：验证数据完整性和引用关系。
4. **格式转换**：通过不同的 IConverter 实现转换为目标格式。
5. **输出处理**：将转换后的数据输出到指定目录。
6. **同步更新**：可选地将数据同步到游戏目录。

## 核心接口

### IReader
用于读取源文件（Excel、CSV）。

### IConverter
用于将数据转换为目标格式（PHP、JSON、FBS）。

### IValidator
用于数据验证。

## 项目结构

```
game-data-builder/
├── cmd/                    # 主程序入口
│   └── main.go             # 主程序
├── internal/               # 内部包
│   ├── config/             # 配置处理
│   ├── converter/          # 转换器实现
│   ├── model/              # 数据模型
│   ├── reader/             # 读取器实现
│   └── validator/          # 验证器实现
├── conf/                   # 配置文件
│   ├── config.json         # 主配置
│   ├── combine.json        # 表合并配置
│   └── replaceColumn.json  # 列替换配置
├── examples/               # 示例数据
│   ├── items.csv           # 示例物品表
│   └── weapons.csv         # 示例武器表
├── output/                 # 输出目录
├── test/                   # 测试用例
└── README.md               # 项目说明
```

## 配置说明

### 主配置文件 (config.json)

```json
{
  "sourceDir": "./examples",       // 源文件目录
  "outputDir": "./output",         // 输出目录
  "formats": ["json", "php", "fbs"],  // 转换格式
  "async": false,                   // 是否异步处理
  "fastMode": false,                // 快速模式
  "syncToGame": false,              // 是否同步到游戏目录
  "gameDir": "",                   // 游戏目录
  "readers": {                      // 读取器配置
    "default": {
      "type": "default",
      "enabled": true,
      "options": {
        "skipEmptyRows": true
      }
    }
  },
  "converters": {                   // 转换器配置
    "json": {
      "type": "json",
      "enabled": true,
      "outputPath": "json",
      "options": {
        "indent": true
      }
    },
    "php": {
      "type": "php",
      "enabled": true,
      "outputPath": "php",
      "options": {}
    },
    "fbs": {
      "type": "fbs",
      "enabled": true,
      "outputPath": "fbs",
      "options": {}
    }
  },
  "validators": {                   // 验证器配置
    "default": {
      "type": "default",
      "enabled": true,
      "options": {
        "strict": true
      }
    }
  }
}
```

## 扩展开发

### 添加新的读取器

1. 实现 `IReader` 接口。
2. 在 `reader_factory.go` 中注册读取器。

### 添加新的转换器

1. 实现 `IConverter` 接口。
2. 在 `converter_factory.go` 中注册转换器。

### 添加新的验证器

1. 实现 `IValidator` 接口。
2. 在主程序中使用新的验证器。

## 测试

运行测试用例：

```bash
go test ./test/...
```

## 许可证

MIT
