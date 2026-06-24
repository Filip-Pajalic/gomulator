# Gomulator WASM Package

This package contains a self-contained browser build:

- `index.html` - example host page that embeds the emulator iframe
- `emulator-iframe.html` - styled emulator page with the ROM picker, canvas, controls, and inline WASM fallback
- `gomulator.wasm` - emulator WebAssembly binary
- `wasm_exec.js` - Go WebAssembly runtime

Open `index.html` directly from disk, or serve the package directory locally.

From the repository root after `make wasm`:

```bash
npx serve build/wasm
```

Or:

```bash
python3 -m http.server 8080 --directory build/wasm
```

If you are already inside the package directory:

```bash
npx serve .
```

Or:

```bash
python3 -m http.server 8080
```

Then open `http://localhost:8080/`.

If `npx` fails with `ENOENT: no such file or directory, uv_cwd`, your shell is probably still inside an old `build/wasm` directory that was deleted and recreated by `make wasm`. Run `cd /home/pajalic/git/gomulator/build/wasm` again, or run one of the repository-root commands above.

To embed the emulator in another page, point an iframe at `emulator-iframe.html`:

```html
<iframe src="/path/to/emulator-iframe.html"></iframe>
```
