#!/usr/bin/env bash
set -euo pipefail

# Simple installer for ghpm on Linux
# - builds the binary with `go build`
# - installs to /usr/local/bin if writable, otherwise $HOME/.local/bin
# - does not use sudo to avoid prompting for passwords

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$script_dir"

echo "ghpm installer: building..."
if ! command -v go >/dev/null 2>&1; then
  echo "Error: Go is not installed or not in PATH. Install Go and retry."
  exit 1
fi

if ! go build -o ghpm; then
  echo "go build failed"
  exit 1
fi

TARGET="/usr/local/bin"
if [ ! -w "$TARGET" ]; then
  TARGET="$HOME/.local/bin"
fi

mkdir -p "$TARGET"
cp -f ghpm "$TARGET/ghpm"
chmod +x "$TARGET/ghpm"

echo "Installed ghpm -> $TARGET/ghpm"

if [[ ":$PATH:" != *":$TARGET:"* ]]; then
  echo "Warning: $TARGET is not in your PATH."
  echo "Add this to your shell profile, for example:"
  echo "  export PATH=\"$HOME/.local/bin:\$PATH\""
fi

echo "Done. Run 'ghpm' to verify." 
