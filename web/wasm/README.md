# Gomulator WASM Package

This package contains a self-contained browser build:

- `index.html` - example host page that embeds the emulator iframe
- `emulator-iframe.html` - styled emulator page with the ROM picker, canvas, controls, and inline WASM fallback
- `gomulator.wasm` - emulator WebAssembly binary
- `wasm_exec.js` - Go WebAssembly runtime

Open `index.html` directly from disk, or serve the package directory locally:

```bash
npx serve .
```

Or:

```bash
python3 -m http.server 8080
```

Then open `http://localhost:8080/`.

To embed the emulator in another page, point an iframe at `emulator-iframe.html`:

```html
<iframe src="/path/to/emulator-iframe.html"></iframe>
```
