#!/bin/sh
# Install Go with Go binary and desktop entry for Omarchy / Linux main app menu.
set -e
cd "$(dirname "$0")/.."

mkdir -p ~/.local/bin
mkdir -p ~/.local/share/applications

if [ -f dist/go-with-go-linux ]; then
    cp dist/go-with-go-linux ~/.local/bin/go-with-go
else
    go build -o ~/.local/bin/go-with-go ./cmd/go-with-go
fi

chmod +x ~/.local/bin/go-with-go
cp assets/go-with-go.desktop ~/.local/share/applications/

if command -v update-desktop-database >/dev/null 2>&1; then
    update-desktop-database ~/.local/share/applications/
fi

echo "Installed Go with Go to ~/.local/bin/go-with-go"
echo "Added desktop shortcut to ~/.local/share/applications/go-with-go.desktop"
echo "Go with Go should now appear in your Omarchy / Linux main apps menu!"
