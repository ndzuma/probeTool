#!/bin/bash
set -e

# probeTool installer script

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Installing probeTool...${NC}"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    arm64) ARCH="arm64" ;;
    *)
        echo -e "${RED}Error: Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

case $OS in
    darwin) OS="darwin" ;;
    linux) OS="linux" ;;
    *)
        echo -e "${RED}Error: Unsupported OS: $OS${NC}"
        echo "Please download manually from: https://github.com/ndzuma/probeTool/releases"
        exit 1
        ;;
esac

# Get latest version from GitHub API
echo -e "${YELLOW}Fetching latest version...${NC}"
LATEST_VERSION=$(curl -s https://api.github.com/repos/ndzuma/probeTool/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_VERSION" ]; then
    echo -e "${RED}Error: Could not fetch latest version${NC}"
    exit 1
fi

echo -e "${GREEN}Latest version: $LATEST_VERSION${NC}"

# Construct download URL
BINARY="probeTool_${LATEST_VERSION}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/ndzuma/probeTool/releases/download/${LATEST_VERSION}/${BINARY}"

echo -e "${YELLOW}Downloading: $URL${NC}"

# Download to temp directory
TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

if ! curl -L -o probe.tar.gz "$URL"; then
    echo -e "${RED}Error: Failed to download${NC}"
    echo "URL: $URL"
    rm -rf "$TMP_DIR"
    exit 1
fi

# Extract
echo -e "${YELLOW}Extracting...${NC}"
tar -xzf probe.tar.gz

# Check if binary exists
if [ ! -f "probe" ]; then
    echo -e "${RED}Error: probe binary not found in archive${NC}"
    rm -rf "$TMP_DIR"
    exit 1
fi

# Install
INSTALL_DIR="/usr/local/bin"
echo -e "${YELLOW}Installing to $INSTALL_DIR...${NC}"

if [ -w "$INSTALL_DIR" ]; then
    mv probe "$INSTALL_DIR/probe"
else
    echo -e "${YELLOW}Requesting sudo access to install to $INSTALL_DIR...${NC}"
    sudo mv probe "$INSTALL_DIR/probe"
fi

# Cleanup
cd - > /dev/null
rm -rf "$TMP_DIR"

# Verify installation
if command -v probe &> /dev/null; then
    echo -e "${GREEN}✅ probeTool installed successfully!${NC}"
    echo ""
    echo "Version: $(probe -v)"
    echo ""
    echo "Quick start:"
    echo "  probe config add-provider openrouter"
    echo "  probe --full"
    echo "  probe tray"
    echo ""
    echo "Documentation: https://github.com/ndzuma/probeTool#readme"
else
    echo -e "${RED}Error: Installation verification failed${NC}"
    exit 1
fi
