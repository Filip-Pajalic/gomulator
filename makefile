# Gomulator Build Makefile

# Default target
.PHONY: all
all: native wasm

# Build native executable for the current platform (release)
.PHONY: native
native:
	@echo Building native executable (release)...
	@go build -o gomulator ./cmd
	@echo Native build complete: gomulator

# Build native executable for the current platform (debug mode)
.PHONY: debug
debug:
	@echo Building native executable (debug mode)...
	@go build -tags debug -o gomulator-debug ./cmd
	@echo Debug build complete: gomulator-debug

# Build WASM version and static test page package
.PHONY: wasm
wasm:
	@bash build-wasm.sh

# Clean build artifacts
.PHONY: clean
clean:
	@echo Cleaning build artifacts...
	@rm -f gomulator gomulator-debug gomulator.exe gomulator-debug.exe
	@rm -f gomulator.wasm gomulator-wasm.zip wasm_exec.js
	@rm -rf dist/wasm
	@echo Clean complete

# Run native version (requires ROM file path)
.PHONY: run
run: native
	@echo Usage: ./gomulator path/to/rom.gb
	@echo Example: ./gomulator myrom.gb

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
	@echo   native  - Build native executable (release mode)
	@echo   debug   - Build native executable (debug mode with DbgPrint)
	@echo   wasm    - Build WASM version and static test page in dist/wasm
	@echo   test    - Run GB test ROM suite (requires bash)
	@echo   clean   - Remove build artifacts
	@echo   run     - Show usage for running emulator
	@echo   help    - Show this help message
