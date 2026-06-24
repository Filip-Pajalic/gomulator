# Gomulator Build Makefile

NATIVE_EXT :=
ifeq ($(OS),Windows_NT)
NATIVE_EXT := .exe
endif
NATIVE_OUT := build/native/gomulator$(NATIVE_EXT)

# Default target
.PHONY: all
all: native wasm

# Build native executable for the current platform (release)
.PHONY: native
native:
	@echo "Building native executable (release)..."
	@mkdir -p build/native
	@go build -o $(NATIVE_OUT) ./cmd
	@echo "Native build complete: $(NATIVE_OUT)"

# Build native executable for the current platform (debug mode)
.PHONY: debug
debug:
	@echo "Building native executable (debug mode)..."
	@mkdir -p build/native
	@go build -tags debug -o build/native/gomulator-debug$(NATIVE_EXT) ./cmd
	@echo "Debug build complete: build/native/gomulator-debug$(NATIVE_EXT)"

# Build WASM version and static test page package
.PHONY: wasm
wasm:
	@bash build-wasm.sh

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -f gomulator gomulator-debug gomulator.exe gomulator-debug.exe
	@rm -f gomulator.wasm gomulator-wasm.zip wasm_exec.js
	@rm -rf dist/wasm
	@rm -rf build
	@echo "Clean complete"

# Run native version (requires ROM file path)
.PHONY: run
run: native
	@echo "Usage: $(NATIVE_OUT) path/to/rom.gb"
	@echo "Example: $(NATIVE_OUT) myrom.gb"

# Run GB test ROMs
.PHONY: test
test: native
	@echo "Running GB test ROM suite..."
	@bash run-tests.sh

# Help target
.PHONY: help
help:
	@echo "Gomulator Build System"
	@echo ""
	@echo "Available targets:"
	@echo "  all     - Build both native and WASM versions (default)"
	@echo "  native  - Build native executable (release mode)"
	@echo "  debug   - Build native executable (debug mode with DbgPrint)"
	@echo "  wasm    - Build WASM version and static test page in build/wasm"
	@echo "  test    - Run GB test ROM suite (requires bash)"
	@echo "  clean   - Remove build artifacts"
	@echo "  run     - Show usage for running emulator"
	@echo "  help    - Show this help message"
