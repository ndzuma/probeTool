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
        else
            URL="https://nodejs.org/dist/v${VERSION}/node-v${VERSION}-darwin-x64.tar.gz"
        fi
        ;;
    linux)
        URL="https://nodejs.org/dist/v${VERSION}/node-v${VERSION}-linux-x64.tar.gz"
        ;;
    windows)
        URL="https://nodejs.org/dist/v${VERSION}/node-v${VERSION}-win-x64.zip"
        ;;
    *)
        echo "Unsupported platform: $PLATFORM"
        exit 1
        ;;
esac

TEMP_FILE="/tmp/node-${PLATFORM}-${ARCH}.tar.gz"
curl -L "$URL" -o "$TEMP_FILE"

echo "📂 Extracting Node.js..."
tar -xzf "$TEMP_FILE" -C /tmp

EXTRACTED=$(ls -d /tmp/node-v${VERSION}* | head -n 1)

echo "📋 Copying Node.js runtime..."
mkdir -p "$OUTPUT_DIR/bin"
cp "$EXTRACTED/bin/node" "$OUTPUT_DIR/bin/"
cp "$EXTRACTED/bin/npm" "$OUTPUT_DIR/bin/"
[ -f "$EXTRACTED/bin/npx" ] && cp "$EXTRACTED/bin/npx" "$OUTPUT_DIR/bin/"

cp -r "$EXTRACTED/lib" "$OUTPUT_DIR/"

echo "✅ Node.js bundled to $OUTPUT_DIR"
rm -rf "$TEMP_FILE" "$EXTRACTED"
