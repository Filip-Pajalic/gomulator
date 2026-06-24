# GoMulator
Game Boy emulator written in Go with Ebiten. Supports native Windows and WASM/browser targets.

## Quick Start

### Native Windows

```bash
make native
./gomulator.exe path/to/rom.gb
```

### WASM/Browser

```bash
make wasm
# Open dist/wasm/index.html directly, or serve it over HTTP:
# python3 -m http.server 8080 --directory dist/wasm
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

`make wasm` writes a self-contained browser package to `dist/wasm/`:

- `index.html` - minimal host page showing how to embed the emulator iframe
- `emulator-iframe.html` - the styled emulator page to embed
- `gomulator.wasm` - emulator build
- `wasm_exec.js` - Go WASM runtime copied from the active Go toolchain

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
- `build.yml` - Builds native + WASM on every push
- `test.yml` - Runs test ROMs on PRs

See [CI_CD_TESTING.md](CI_CD_TESTING.md) for details.

## License

See [LICENSE](LICENSE)

