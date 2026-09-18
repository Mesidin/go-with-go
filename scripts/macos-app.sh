#!/bin/sh
# Build a macOS .app you can put in /Applications and drag onto the Dock.
# Run this on a Mac, from the repo root:
#   sh scripts/macos-app.sh
set -e
cd "$(dirname "$0")/.."

NAME="Go with Go"
BIN="go-with-go"
PKG="./cmd/go-with-go"
IDENT="dev.erickson.go-with-go"
APP="${NAME}.app"

rm -rf "$APP"
mkdir -p "$APP/Contents/MacOS" "$APP/Contents/Resources"

go build -o "$APP/Contents/MacOS/$BIN" "$PKG"

cat > "$APP/Contents/Info.plist" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleExecutable</key>
  <string>$BIN</string>
  <key>CFBundleIdentifier</key>
  <string>$IDENT</string>
  <key>CFBundleName</key>
  <string>$NAME</string>
  <key>CFBundleDisplayName</key>
  <string>$NAME</string>
  <key>CFBundlePackageType</key>
  <string>APPL</string>
  <key>CFBundleVersion</key>
  <string>1.0</string>
  <key>LSMinimumSystemVersion</key>
  <string>11.0</string>
  <key>NSHighResolutionCapable</key>
  <true/>
</dict>
</plist>
EOF

echo "Built $APP"
echo "Copy it to Applications, then drag it onto the Dock:"
echo "  cp -R \"$APP\" /Applications/"
echo "  open /Applications"
