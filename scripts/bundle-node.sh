#!/bin/bash
set -e

VERSION="20.18.2"
PLATFORM=$1
ARCH=$2

if [ -z "$PLATFORM" ] || [ -z "$ARCH" ]; then
    echo "Usage: $0 <platform> <arch>"
    exit 1
fi

OUTPUT_DIR="internal/runtime/node-${PLATFORM}-${ARCH}"
mkdir -p "$OUTPUT_DIR"

echo "📦 Downloading Node.js ${VERSION} for ${PLATFORM}/${ARCH}..."

case "$PLATFORM" in
    darwin)
        if [ "$ARCH" = "arm64" ]; then
            URL="https://nodejs.org/dist/v${VERSION}/node-v${VERSION}-darwin-arm64.tar.gz"
            ARCHIVE_TYPE="tar.gz"
        else
            URL="https://nodejs.org/dist/v${VERSION}/node-v${VERSION}-darwin-x64.tar.gz"
            ARCHIVE_TYPE="tar.gz"
        fi
        ;;
    linux)
        URL="https://nodejs.org/dist/v${VERSION}/node-v${VERSION}-linux-x64.tar.gz"
        ARCHIVE_TYPE="tar.gz"
        ;;
    windows)
        URL="https://nodejs.org/dist/v${VERSION}/node-v${VERSION}-win-x64.zip"
        ARCHIVE_TYPE="zip"
        ;;
    *)
        echo "Unsupported platform: $PLATFORM"
        exit 1
        ;;
esac

if [ "$ARCHIVE_TYPE" = "zip" ]; then
    TEMP_FILE="/tmp/node-${PLATFORM}-${ARCH}.zip"
else
    TEMP_FILE="/tmp/node-${PLATFORM}-${ARCH}.tar.gz"
fi

curl -L "$URL" -o "$TEMP_FILE"

echo "📂 Extracting Node.js..."
if [ "$ARCHIVE_TYPE" = "zip" ]; then
    7z x "$TEMP_FILE" -o/tmp -y > /dev/null
else
    tar -xzf "$TEMP_FILE" -C /tmp
fi

EXTRACTED=$(ls -d /tmp/node-v${VERSION}* | head -n 1)

echo "📋 Copying Node.js runtime..."
mkdir -p "$OUTPUT_DIR/bin"

if [ "$PLATFORM" = "windows" ]; then
    cp "$EXTRACTED/node.exe" "$OUTPUT_DIR/bin/"
    cp "$EXTRACTED/npm.cmd" "$OUTPUT_DIR/bin/" 2>/dev/null || true
    cp "$EXTRACTED/npx.cmd" "$OUTPUT_DIR/bin/" 2>/dev/null || true
    mkdir -p "$OUTPUT_DIR/node_modules"
    cp -r "$EXTRACTED/node_modules" "$OUTPUT_DIR/"
else
    cp "$EXTRACTED/bin/node" "$OUTPUT_DIR/bin/"
    cp "$EXTRACTED/bin/npm" "$OUTPUT_DIR/bin/"
    [ -f "$EXTRACTED/bin/npx" ] && cp "$EXTRACTED/bin/npx" "$OUTPUT_DIR/bin/"
    cp -r "$EXTRACTED/lib" "$OUTPUT_DIR/"
fi

echo "✅ Node.js bundled to $OUTPUT_DIR"
rm -rf "$TEMP_FILE" "$EXTRACTED"
