# Gomulator Build Makefile

# Default target
.PHONY: all
all: native wasm

# Build native Windows executable
.PHONY: native
native:
	@echo Building native Windows executable...
	@set GOOS=& set GOARCH=& go build -o gomulator.exe ./cmd
	@echo Native build complete: gomulator.exe

# Build WASM version with package
.PHONY: wasm
wasm:
	@echo Building WASM package...
	@set GOOS=js& set GOARCH=wasm& go build -ldflags="-s -w" -o gomulator.wasm ./cmd
	@echo WASM build complete: gomulator.wasm
	@echo Downloading wasm_exec.js...
	@powershell -Command "$$ver = (go version) -replace '.*go(\d+\.\d+\.\d+).*', 'go$$1'; $$url = \"https://raw.githubusercontent.com/golang/go/refs/tags/$$ver/lib/wasm/wasm_exec.js\"; Write-Host \"Downloading from $$url\"; Invoke-WebRequest -Uri $$url -OutFile wasm_exec.js; Write-Host \"Downloaded wasm_exec.js for $$ver\""
	@echo Creating WASM package...
	@powershell -Command "Compress-Archive -Force -Path gomulator.wasm,wasm_exec.js,index.html -DestinationPath gomulator-wasm.zip"
	@echo WASM package created: gomulator-wasm.zip

# Clean build artifacts
.PHONY: clean
clean:
	@echo Cleaning build artifacts...
	@if exist gomulator.exe del gomulator.exe
	@if exist gomulator.wasm del gomulator.wasm
	@if exist gomulator-wasm.zip del gomulator-wasm.zip
	@if exist wasm_exec.js del wasm_exec.js
	@echo Clean complete

# Run native version (requires ROM file path)
.PHONY: run
run: native
	@echo Usage: gomulator.exe path\to\rom.gb
	@echo Example: gomulator.exe myrom.gb

# Start HTTP server for WASM version
.PHONY: serve
serve: wasm
	@echo Starting HTTP server on http://localhost:8080
	@echo Serving WASM files...
	@python -m http.server 8080

# Run GB test ROMs
.PHONY: test
test: native
	@echo Running GB test ROM suite...
	@bash run-tests.sh

# Help target
.PHONY: help
help:
	@echo Gomulator Build System
	@echo.
	@echo Available targets:
	@echo   all     - Build both native and WASM versions (default)
	@echo   native  - Build native Windows executable
	@echo   wasm    - Build WASM version and create deployment package (zip)
	@echo   test    - Run GB test ROM suite (requires bash)
	@echo   clean   - Remove build artifacts
	@echo   run     - Show usage for running emulator
	@echo   serve   - Build WASM and start HTTP server
	@echo   help    - Show this help message
