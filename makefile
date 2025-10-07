# Gomulator Build Makefile

# Default target
.PHONY: all
all: native wasm

# Build native Windows executable (release)
.PHONY: native
native:
	@echo Building native Windows executable (release)...
	@set GOOS=& set GOARCH=& go build -o gomulator.exe ./cmd
	@echo Native build complete: gomulator.exe

# Build native Windows executable (debug mode)
.PHONY: debug
debug:
	@echo Building native Windows executable (debug mode)...
	@set GOOS=& set GOARCH=& go build -tags debug -o gomulator-debug.exe ./cmd
	@echo Debug build complete: gomulator-debug.exe

# Build WASM version with package
.PHONY: wasm
wasm:
	@echo Building WASM package...
	@set GOOS=js& set GOARCH=wasm& go build -ldflags="-s -w" -o gomulator.wasm ./cmd
	@echo WASM build complete: gomulator.wasm
	@bash build-wasm.sh

# Clean build artifacts
.PHONY: clean
clean:
	@echo Cleaning build artifacts...
	@if exist gomulator.exe del gomulator.exe
	@if exist gomulator-debug.exe del gomulator-debug.exe
	@if exist gomulator.wasm del gomulator.wasm
	@if exist gomulator-wasm.zip del gomulator-wasm.zip
	@if exist wasm_exec.js del wasm_exec.js
	@echo Clean complete

# Run native version (requires ROM file path)
.PHONY: run
run: native
	@echo Usage: gomulator.exe path\to\rom.gb
	@echo Example: gomulator.exe myrom.gb

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
	@echo   native  - Build native Windows executable (release mode)
	@echo   debug   - Build native Windows executable (debug mode with DbgPrint)
	@echo   wasm    - Build WASM version and create deployment package (zip)
	@echo   test    - Run GB test ROM suite (requires bash)
	@echo   clean   - Remove build artifacts
	@echo   run     - Show usage for running emulator
	@echo   help    - Show this help message
