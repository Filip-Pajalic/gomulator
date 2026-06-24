#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BUILD_DIR="${BUILD_DIR:-"${ROOT_DIR}/build"}"
WASM_DIR="${WASM_DIR:-"${BUILD_DIR}/wasm"}"
ARTIFACT_DIR="${ARTIFACT_DIR:-"${BUILD_DIR}/artifacts"}"
WASM_ZIP="${WASM_ZIP:-"${ARTIFACT_DIR}/gomulator-wasm.zip"}"

GOROOT="$(go env GOROOT)"
WASM_EXEC="${GOROOT}/lib/wasm/wasm_exec.js"
if [[ ! -f "${WASM_EXEC}" ]]; then
  echo "wasm_exec.js not found at ${WASM_EXEC}" >&2
  exit 1
fi
if ! command -v zip >/dev/null 2>&1; then
  echo "zip is required to build the WASM package" >&2
  exit 1
fi

rm -rf "${WASM_DIR}"
mkdir -p "${WASM_DIR}" "${ARTIFACT_DIR}"
rm -f "${ROOT_DIR}/gomulator-wasm.zip"
rm -rf "${ROOT_DIR}/dist/wasm"
: "${GOCACHE:="${TMPDIR:-/tmp}/gomulator-go-build-cache"}"
export GOCACHE
mkdir -p "${GOCACHE}"

echo "Building gomulator.wasm..."
(
  cd "${ROOT_DIR}"
  GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o "${WASM_DIR}/gomulator.wasm" ./cmd
)

cp "${WASM_EXEC}" "${WASM_DIR}/wasm_exec.js"
cp "${ROOT_DIR}/web/wasm/index.html" "${WASM_DIR}/index.html"
cp "${ROOT_DIR}/web/wasm/emulator-iframe.html" "${WASM_DIR}/emulator-iframe.html"
cp "${ROOT_DIR}/web/wasm/README.md" "${WASM_DIR}/README.md"

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
      sub(/\r$/, "")
      sub(marker, "const GOMULATOR_WASM_BASE64 = \"")
      printf "%s", $0
      exit
    }
    {
      sub(/\r$/, "")
      print
    }
  ' "${html_file}" > "${tmp_file}"

  base64_nowrap "${wasm_file}" >> "${tmp_file}"
  printf '";\n' >> "${tmp_file}"

  awk -v marker="${marker}" '
    found {
      sub(/\r$/, "")
      print
    }
    index($0, marker) { found = 1 }
  ' "${html_file}" >> "${tmp_file}"

  mv "${tmp_file}" "${html_file}"
  chmod 0644 "${html_file}"
}

inject_inline_wasm "${WASM_DIR}/emulator-iframe.html" "${WASM_DIR}/gomulator.wasm"

(
  cd "${WASM_DIR}"
  zip -qr "${WASM_ZIP}" README.md index.html emulator-iframe.html wasm_exec.js gomulator.wasm
)

cat <<EOF
WASM package ready:
  ${WASM_DIR}/README.md
  ${WASM_DIR}/index.html
  ${WASM_DIR}/emulator-iframe.html
  ${WASM_DIR}/gomulator.wasm
  ${WASM_DIR}/wasm_exec.js
  ${WASM_ZIP}

Open the standalone page directly:
  ${WASM_DIR}/index.html

Or serve it locally with:
  npx serve ${WASM_DIR}

Or:
  python3 -m http.server 8080 --directory ${WASM_DIR}

Then open:
  http://localhost:8080/

Embed later with:
  <iframe src="/path/to/emulator-iframe.html"></iframe>
EOF
