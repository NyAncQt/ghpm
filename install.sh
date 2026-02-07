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
  # Determine a sensible profile to modify
  profile="$HOME/.profile"
  if [ -n "${SHELL:-}" ] && [[ "$SHELL" == */zsh ]]; then
    profile="$HOME/.zshrc"
  elif [ -n "${SHELL:-}" ] && [[ "$SHELL" == */bash ]]; then
    if [ -f "$HOME/.bash_profile" ]; then
      profile="$HOME/.bash_profile"
    else
      profile="$HOME/.bashrc"
    fi
  fi

  line="export PATH=\"$TARGET:\$PATH\""
  # Create profile if it doesn't exist and append the line if missing
  touch "$profile"
  if ! grep -Fqx "$line" "$profile" 2>/dev/null; then
    printf "\n# Added by ghpm install.sh\n%s\n" "$line" >> "$profile"
    echo "Added $TARGET to PATH in $profile"
    echo "You may need to run: source $profile"
  else
    echo "$TARGET already present in $profile"
  fi
fi

echo "Done. Run 'ghpm' to verify." 
