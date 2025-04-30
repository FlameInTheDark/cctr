# CCTR - Half-Life Closed Captions Translator

CCTR is a command-line tool for translating Half-Life game closed caption files using the DeepL translation API.

## Requirements

- Go 1.16 or higher
- DeepL API key (get one at [DeepL API](https://www.deepl.com/pro-api))

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/FlameInTheDark/cctr.git
cd cctr

# Build the project
go build -o cctr # cctr.exe for windows

# Move to a directory in your PATH (optional)
mv cctr /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/FlameInTheDark/cctr@latest
```

## Usage

CCTR provides several commands to help you translate Half-Life closed caption files:

### Translate a File

Translates a closed caption file to the specified language:

```bash
cctr translate --key YOUR_DEEPL_API_KEY --lang TARGET_LANGUAGE_CODE path/to/file.txt
```

Options:
- `--key, -k`: DeepL API key (required)
- `--lang, -l`: Target language code (default: "EN")
- `--out, -o`: Output file path (default: auto-generated filename)

### Count Symbols

Calculates the number of symbols to translate and compares with your DeepL plan limits:

```bash
cctr count path/to/file.txt
```

With API key to check against your plan limits:

```bash
cctr count --key YOUR_DEEPL_API_KEY path/to/file.txt
```

### List Supported Languages

Shows a list of languages supported by DeepL:

```bash
cctr languages --key YOUR_DEEPL_API_KEY
```

## Example

1. Check how many characters need to be translated:
   ```bash
   cctr count --key YOUR_DEEPL_API_KEY path/to/closedcaption_english.txt
   ```

2. Check available target languages:
   ```bash
   cctr languages --key YOUR_DEEPL_API_KEY
   ```

3. Translate the file to Spanish:
   ```bash
   cctr translate --key YOUR_DEEPL_API_KEY --lang ES --out closedcaption_spanish.txt path/to/closedcaption_english.txt
   ```

## Notes

- The DeepL API has character limits based on your subscription plan
- The tool will check if you have enough characters available before translation
- For large files, the translation is split into chunks to comply with API limitations

## License

[MIT License](LICENSE)
