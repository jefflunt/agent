#!/bin/bash
set -e

# Detect OS and Arch
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Map to match release filenames
if [ "$ARCH" == "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" == "arm64" ] || [ "$ARCH" == "aarch64" ]; then
    ARCH="arm64"
fi

if [ "$OS" == "darwin" ]; then
    OS="darwin"
elif [ "$OS" == "linux" ]; then
    OS="linux"
elif [ "$OS" == "mingw64_nt" ] || [ "$OS" == "msys_nt" ]; then
    OS="windows"
fi

# We need a tag
TAG=$1

if [ -z "$TAG" ] || [ "$TAG" == "latest" ]; then
    echo "Fetching latest release tag..."
    TAG=$(curl -sL https://api.github.com/repos/jefflunt/agent/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$TAG" ]; then
        echo "Error: Could not determine the latest release tag."
        exit 1
    fi
    echo "Latest release is $TAG"
fi

BINARY_NAME="agent-${OS}-${ARCH}"
if [ "$OS" == "windows" ]; then
    BINARY_NAME="${BINARY_NAME}.exe"
fi

URL="https://github.com/jefflunt/agent/releases/download/${TAG}/${BINARY_NAME}"


echo "Downloading ${BINARY_NAME} from ${URL}..."

curl -sL "$URL" -o agent

chmod +x agent

echo "Installed agent to ./agent"
echo "Move it to your PATH to use it."
