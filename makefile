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

# Build WASM version
.PHONY: wasm
wasm:
	@echo Building WASM version...
	@set GOOS=js& set GOARCH=wasm& go build -ldflags="-s -w" -o gomulator.wasm ./cmd
	@echo WASM build complete: gomulator.wasm

# Clean build artifacts
.PHONY: clean
clean:
	@echo Cleaning build artifacts...
	@if exist gomulator.exe del gomulator.exe
	@if exist gomulator.wasm del gomulator.wasm
	@if exist test_runner.exe del test_runner.exe
	@echo Clean complete

# Run native version with Tetris
.PHONY: run
run: native
	@echo Running Tetris...
	@gomulator.exe "Tetris (World) (Rev A).gb"

# Start HTTP server for WASM version
.PHONY: serve
serve: wasm
	@echo Starting HTTP server on http://localhost:8080
	@python -m http.server 8080

# Build test runner
.PHONY: test-runner
test-runner:
	@echo Building test runner...
	@go build -tags testrunner -o test_runner.exe ./cmd/test_runner.go
	@echo Test runner build complete: test_runner.exe

# Run GB test ROMs
.PHONY: test
test: test-runner
	@echo Running GB test ROM suite...
	@powershell -ExecutionPolicy Bypass -File run-tests.ps1

# Help target
.PHONY: help
help:
	@echo Gomulator Build System
	@echo.
	@echo Available targets:
	@echo   all          - Build both native and WASM versions (default)
	@echo   native       - Build native Windows executable
	@echo   wasm         - Build WASM version for browsers
	@echo   test-runner  - Build headless test runner
	@echo   test         - Run GB test ROM suite
	@echo   clean        - Remove build artifacts
	@echo   run          - Build and run native version with Tetris
	@echo   serve        - Build WASM and start HTTP server
	@echo   help         - Show this help message
