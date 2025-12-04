# Game Data Builder

A powerful Go-based game data building tool that converts Excel/CSV files to multiple formats (JSON, PHP, FlatBuffers) with validation, async processing, and fast mode support.

## Features

### Core Features
- **Multi-format Conversion**: Convert Excel/CSV files to JSON, PHP, and FlatBuffers formats
- **Data Validation**: Built-in data integrity checks and reference validation
- **Async Processing**: Concurrent format conversion for improved performance
- **Fast Mode**: Only process modified files to save time
- **Flexible Configuration**: JSON-based configuration for easy setup
- **Modular Design**: Interfaces for readers, converters, and validators for extensibility

### Advanced Features
- **Data Merging**: Merge multiple sheets into one output file
- **Column Replacement**: Replace values in specific columns based on rules
- **Game Directory Sync**: Automatically sync generated files to game directory
- **Multi-file Configuration**: Separate config files for main settings, merging, and replacements

## Installation

### Prerequisites
- Go 1.20 or later

### Setup
1. Clone or download the project
2. Install dependencies:
   ```bash
go mod tidy
   ```
3. Build the binary:
   ```bash
go build -o builder ./cmd/main.go
   ```

## Configuration

### Directory Structure
```
conf/
├── config.json          # Main configuration
├── combine.json         # Sheet merging configuration
└── replaceColumn.json   # Column replacement rules
```

### Main Configuration (config.json)
```json
{
  "sourceDir": "./data",
  "outputDir": "./output",
  "formats": ["json", "php", "fbs"],
  "async": false,
  "fastMode": false,
  "syncToGame": false,
  "gameDir": "./game",
  "readers": {
    "default": {
      "type": "excel",
      "enabled": true,
      "options": {
        "sheetName": "",
        "headerRow": 1
      }
    }
  },
  "converters": {
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
      "options": {
        "arrayName": "data"
      }
    },
    "fbs": {
      "type": "fbs",
      "enabled": true,
      "outputPath": "fbs",
      "options": {
        "schemaPath": "./schema"
      }
    }
  },
  "validators": {
    "default": {
      "type": "default",
      "enabled": true,
      "options": {
        "checkReferences": true
      }
    }
  }
}
```

### Sheet Merging Configuration (combine.json)
```json
{
  "sheets": {
    "all_items": {
      "sourceSheets": ["items", "weapons"],
      "keyColumn": "id",
      "outputName": "all_items"
    }
  }
}
```

### Column Replacement Configuration (replaceColumn.json)
```json
{
  "sheets": {
    "items": {
      "columns": {
        "description": {
          "from": "\n",
          "to": "<br>"
        }
      }
    }
  }
}
```

## Usage

### Basic Usage
```bash
# Run with default configuration
./builder

# Run with custom config directory
./builder -conf ./custom_conf

# Run in fast mode (only modified files)
./builder -fast

# Run with async processing
./builder -async

# Show help
./builder -help
```

### Command-line Options
| Option | Description | Default |
|--------|-------------|---------|
| `-conf` | Configuration file directory | `./conf` |
| `-fast` | Fast mode, only process modified files | `false` |
| `-async` | Async processing | `false` |
| `-help` | Show help information | - |

## Project Structure

```
.
├── cmd/
│   └── main.go              # Main program entry
├── internal/
│   ├── config/              # Configuration management
│   ├── converter/           # Format converters (JSON, PHP, FBS)
│   ├── model/               # Core data structures
│   ├── reader/              # File readers (Excel, CSV)
│   └── validator/           # Data validators
├── conf/                    # Configuration files
├── data/                    # Source Excel/CSV files
├── output/                  # Generated output files
└── README.md                # This file
```

## Architecture

### Core Interfaces

#### IReader
```go
type IReader interface {
    ReadAll(path string) ([]*model.DataSheet, error)
    ReadSheet(path string, sheetName string) (*model.DataSheet, error)
}
```

#### IConverter
```go
type IConverter interface {
    Convert(sheet *model.DataSheet) (*model.ConvertResult, error)
    BatchConvert(sheets []*model.DataSheet) ([]*model.ConvertResult, error)
}
```

#### IValidator
```go
type IValidator interface {
    Validate(sheet *model.DataSheet) []*model.ErrorInfo
    ValidateAll(sheets []*model.DataSheet) []*model.ErrorInfo
}
```

### Workflow
1. **Initialization**: Load configuration files
2. **Data Reading**: Read Excel/CSV files using appropriate readers
3. **Data Validation**: Validate data integrity and references
4. **Data Conversion**: Convert to configured formats (sync or async)
5. **Output Processing**: Write generated files to output directory
6. **Game Sync**: Optionally sync files to game directory

## Example

### Source File (items.xlsx)
| id | name       | type | price |
|----|------------|------|-------|
| 1  | Sword      | 1    | 100   |
| 2  | Shield     | 2    | 150   |
| 3  | Health Pot | 3    | 50    |

### Generated Files

#### items.json
```json
{
  "1": {
    "id": 1,
    "name": "Sword",
    "type": 1,
    "price": 100
  },
  "2": {
    "id": 2,
    "name": "Shield",
    "type": 2,
    "price": 150
  },
  "3": {
    "id": 3,
    "name": "Health Pot",
    "type": 3,
    "price": 50
  }
}
```

#### items.php
```php
<?php
return array(
    '1' => array(
        'id' => 1,
        'name' => 'Sword',
        'type' => 1,
        'price' => 100,
    ),
    '2' => array(
        'id' => 2,
        'name' => 'Shield',
        'type' => 2,
        'price' => 150,
    ),
    '3' => array(
        'id' => 3,
        'name' => 'Health Pot',
        'type' => 3,
        'price' => 50,
    ),
);
```

## Extending the Tool

### Adding a New Reader
1. Implement the `IReader` interface
2. Register the reader in `reader_factory.go`
3. Configure the reader in `config.json`

### Adding a New Converter
1. Implement the `IConverter` interface
2. Register the converter in `converter_factory.go`
3. Configure the converter in `config.json`

### Adding a New Validator
1. Implement the `IValidator` interface
2. Register the validator in `validator_factory.go`
3. Configure the validator in `config.json`

## Performance Tips

1. **Use Fast Mode**: Enable `fastMode` in config to only process modified files
2. **Enable Async Processing**: Set `async` to true for concurrent format conversion
3. **Limit Formats**: Only enable the formats you actually need
4. **Optimize Source Files**: Keep source files organized and remove unnecessary sheets

## Troubleshooting

### Common Issues

1. **"Failed to load config"**: Check the config file path and syntax
2. **"File not found"**: Verify the source files exist in the configured `sourceDir`
3. **"Validation errors"**: Check the error messages for data integrity issues
4. **"Permission denied"**: Ensure write permissions for output and game directories

### Logging
The tool outputs detailed information about the build process, including:
- Files being processed
- Generated output files
- Validation errors
- Build time and statistics

## Contributing

Contributions are welcome! Please feel free to submit issues, feature requests, or pull requests.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
