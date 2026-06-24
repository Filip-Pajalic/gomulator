# GoMulator
Game Boy emulator written in Go with Ebiten. Supports native Windows and WASM/browser targets.

## Quick Start

### Native Windows

```bash
make native
./build/native/gomulator.exe path/to/rom.gb
```

### WASM/Browser

```bash
make wasm
cd build/wasm
npx serve .
# or: python3 -m http.server 8080
```

## Build Commands

```bash
make all      # Build native and WASM
make native   # Build native executable
make debug    # Build with CPU instruction tracing
make wasm     # Build WASM package with a local test page
make test     # Run GB test ROM suite
make clean    # Remove build artifacts
```

Build commands write generated files under `build/`:

- `build/native/gomulator` or `build/native/gomulator.exe` - native executable
- `build/native/gomulator-debug` or `build/native/gomulator-debug.exe` - debug executable
- `build/wasm/` - self-contained browser package
- `build/artifacts/gomulator-wasm.zip` - zipped WASM browser package

`make wasm` writes this self-contained browser package to `build/wasm/`:

- `README.md` - package usage and serve commands
- `index.html` - minimal host page showing how to embed the emulator iframe
- `emulator-iframe.html` - the styled emulator page to embed
- `gomulator.wasm` - emulator build
- `wasm_exec.js` - Go WASM runtime copied from the active Go toolchain

Serve the generated browser package with either command:

```bash
cd build/wasm
npx serve .
```

```bash
cd build/wasm
python3 -m http.server 8080
```

For embedding, point an iframe at the generated `emulator-iframe.html`:

```html
<iframe src="/path/to/emulator-iframe.html"></iframe>
```

## Command Line Options

```bash
./gomulator [options] <rom_file>

Options:
  -debug        Enable debug logging
  -fps          Show FPS counter (toggle with F3)
```

## Controls

**Game:**
- Arrow keys: D-pad
- Z: B button  
- X: A button
- Enter: Start
- Tab: Select

**Debug:**
- F3: Toggle FPS display

## Build Tags

Platform-specific code uses Go build tags:

- `desktop.go` - `//go:build !js || !wasm`
- `wasm.go` - `//go:build js && wasm`
- `debug_enabled.go` - `//go:build debug` (CPU tracing)
- `debug_disabled.go` - `//go:build !debug` (production)

Build with debug:
```bash
go build -tags debug -o gomulator-debug.exe ./cmd
```

## Project Structure

```
cmd/
├── main.go          # Entry point
├── desktop.go       # Native platform
└── wasm.go          # WASM platform

internal/
├── cpu/             # CPU emulation
├── ui/              # Graphics and input
├── memory/          # Memory and cartridge
└── ...
```

## Testing

```bash
make test
```

Runs GB test ROMs from https://github.com/retrio/gb-test-roms

## CI/CD

GitHub Actions workflows:
- `build.yml` - Builds release artifacts on PRs and publishes a GitHub Release on merges to `main`
- `test.yml` - Runs test ROMs on PRs

Release assets published from `main`:

- `gomulator-windows-x64.exe` - Windows executable
- `gomulator-linux-arm64` - Linux ARM64 executable
- `gomulator-macos` - macOS executable
- `gomulator.wasm` - raw WASM binary
- `gomulator-wasm.zip` - zip containing the WASM test/embed page, iframe page, runtime, README, and WASM binary

See [CI_CD_TESTING.md](CI_CD_TESTING.md) for details.

## License

See [LICENSE](LICENSE)

