#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DIST_DIR="${WASM_DIST_DIR:-"${ROOT_DIR}/dist/wasm"}"

GOROOT="$(go env GOROOT)"
WASM_EXEC="${GOROOT}/lib/wasm/wasm_exec.js"
if [[ ! -f "${WASM_EXEC}" ]]; then
  echo "wasm_exec.js not found at ${WASM_EXEC}" >&2
  exit 1
fi

rm -rf "${DIST_DIR}"
mkdir -p "${DIST_DIR}"
rm -f "${ROOT_DIR}/gomulator-wasm.zip"
: "${GOCACHE:="${TMPDIR:-/tmp}/gomulator-go-build-cache"}"
export GOCACHE
mkdir -p "${GOCACHE}"

echo "Building gomulator.wasm..."
(
  cd "${ROOT_DIR}"
  GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o "${DIST_DIR}/gomulator.wasm" ./cmd
)

cp "${WASM_EXEC}" "${DIST_DIR}/wasm_exec.js"
cp "${ROOT_DIR}/index.html" "${DIST_DIR}/index.html"
cp "${ROOT_DIR}/web/wasm/emulator-iframe.html" "${DIST_DIR}/emulator-iframe.html"

base64_nowrap() {
  if base64 --help 2>/dev/null | grep -q -- "-w"; then
    base64 -w 0 "$1"
  else
    base64 "$1" | tr -d "\n"
  fi
}

inject_inline_wasm() {
  local html_file="$1"
  local wasm_file="$2"
  local marker='const GOMULATOR_WASM_BASE64 = "";'
  local tmp_file
  tmp_file="$(mktemp)"

  awk -v marker="${marker}" '
    index($0, marker) {
      sub(marker, "const GOMULATOR_WASM_BASE64 = \"")
      printf "%s", $0
      exit
    }
    { print }
  ' "${html_file}" > "${tmp_file}"

  base64_nowrap "${wasm_file}" >> "${tmp_file}"
  printf '";\n' >> "${tmp_file}"

  awk -v marker="${marker}" '
    found { print }
    index($0, marker) { found = 1 }
  ' "${html_file}" >> "${tmp_file}"

  mv "${tmp_file}" "${html_file}"
  chmod 0644 "${html_file}"
}

inject_inline_wasm "${DIST_DIR}/emulator-iframe.html" "${DIST_DIR}/gomulator.wasm"

if command -v zip >/dev/null 2>&1; then
  (
    cd "${DIST_DIR}"
    zip -qr "${ROOT_DIR}/gomulator-wasm.zip" index.html emulator-iframe.html wasm_exec.js gomulator.wasm
  )
fi

cat <<EOF
WASM package ready:
  ${DIST_DIR}/index.html
  ${DIST_DIR}/emulator-iframe.html
  ${DIST_DIR}/gomulator.wasm
  ${DIST_DIR}/wasm_exec.js

Open the standalone page directly:
  ${DIST_DIR}/index.html

Or serve it locally with:
  python3 -m http.server 8080 --directory ${DIST_DIR}

Then open:
  http://localhost:8080/

Embed later with:
  <iframe src="/path/to/emulator-iframe.html"></iframe>
EOF
