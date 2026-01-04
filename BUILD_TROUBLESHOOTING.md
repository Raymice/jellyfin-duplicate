# Build Troubleshooting Guide

## Common Build Issues and Solutions

### 1. Syntax Errors

If you encounter syntax errors like:
```
internal/jellyfin/client.go:60:3: syntax error: unexpected ., expected }
internal/jellyfin/client.go:77:3: syntax error: unexpected ., expected }
```

**Solution**: These errors are typically caused by incorrect method chaining syntax. The issues have been fixed in the current version by:

1. **Fixed method chaining**: Changed from:
   ```go
   c.client.R()
       .SetHeader(...)
       .SetResult(...)
   ```
   
   To:
   ```go
   c.client.R().
       SetHeader(...).
       SetResult(...)
   ```

2. **Removed unused imports**: Removed the unused `os` import from `main.go`

### 2. Missing Dependencies

If you get errors about missing packages:
```
cannot find package "github.com/gin-gonic/gin"
```

**Solution**: Run:
```bash
go mod tidy
```

This will download all required dependencies.

### 3. Go Version Issues

If you see errors about Go version requirements:
```
go: version 1.21 required by module, but go1.19 installed
```

**Solution**: Install Go 1.21 or later:
- **macOS**: `brew install go`
- **Linux**: Follow instructions at https://golang.org/doc/install
- **Windows**: Download installer from https://golang.org/dl/

### 4. Configuration File Issues

If the application fails to start with:
```
Failed to load config: open config.json: no such file or directory
```

**Solution**:
1. Make sure `config.json` exists in the same directory as the executable
2. Copy the sample configuration:
```bash
cp config.json.example config.json
```
3. Edit `config.json` with your Jellyfin server details

### 5. Template Loading Issues

If you see errors about templates:
```
html/template: pattern matches no files
```

**Solution**:
1. Make sure the `web/templates` directory exists
2. Verify the templates are in the correct location
3. Check that the template loading path is correct in `main.go`

## Verification Steps

### 1. Check Go Installation
```bash
go version
```
Should output: `go version go1.21.x` or later

### 2. Verify Dependencies
```bash
go mod verify
```

### 3. Test Build
```bash
go build -v ./...
```

### 4. Run Tests (if available)
```bash
go test ./...
```

## Manual Syntax Checking

If you don't have Go installed but want to verify the syntax, you can:

1. **Check for balanced braces**: Each `{` should have a matching `}`
2. **Verify method chaining**: Method calls should be properly chained with `.`
3. **Check imports**: All imports should be used in the file
4. **Validate JSON tags**: Struct field tags should be properly formatted

## Fixed Files Summary

The following syntax issues have been resolved:

### internal/jellyfin/client.go
- Fixed method chaining syntax in `getLibraries()` (line ~60)
- Fixed method chaining syntax in `getMoviesFromLibrary()` (line ~77)

### cmd/api/main.go
- Removed unused `os` import

### go.mod
- Updated to use correct Levenshtein package version

## Building from Scratch

If you're still having issues, try building from scratch:

```bash
# Remove existing build artifacts
rm -rf jellyfin-duplicate

# Clone fresh copy (replace with your actual repo)
git clone https://github.com/yourusername/jellyfin-duplicate.git
cd jellyfin-duplicate

# Initialize and update dependencies
go mod tidy

# Build the application
go build -o jellyfin-duplicate ./cmd/api

# Run the application
./jellyfin-duplicate
```

## Getting Help

If you're still experiencing build issues:

1. **Check the exact error message** and search for it online
2. **Verify your Go installation** with `go version`
3. **Try building a simple Go program** to verify your environment
4. **Check file permissions** to ensure you can read/write all files
5. **Consult the Go documentation** at https://golang.org/doc/