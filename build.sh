#!/usr/bin/env bash

set -e

PROJECT_NAME="lanscan"
BUILD_DIR="./bin"
OUTPUT_FILE=""
GOOS=""
GOARCH=""
INSTALL_DIR=""

detect_arch() {
  local arch
  arch=$(uname -m)
  case "$arch" in
    x86_64) GOARCH="amd64" ;;
    i386 | i686) GOARCH="386" ;;
    aarch64 | arm64) GOARCH="arm64" ;;
    armv7l) GOARCH="arm" ;;
    ppc64le) GOARCH="ppc64le" ;;
    s390x) GOARCH="s390x" ;;
    *)
      echo "Unsupported architecture: $arch"
      exit 1
      ;;
  esac
}

detect_os() {
  local os
  os=$(uname -s)
  case "$os" in
    Linux*) GOOS="linux" ;;
    Darwin*) GOOS="darwin" ;;
    CYGWIN* | MINGW* | MSYS*) GOOS="windows" ;;
    *)
      echo "Unsupported OS: $os"
      exit 1
      ;;
  esac
}

set_install_dir() {
  if [ "$GOOS" = "windows" ]; then
    INSTALL_DIR="$HOME/bin"
  else
    INSTALL_DIR="/usr/local/bin"
  fi
}

build_binary() {
  mkdir -p "$BUILD_DIR"

  if [ "$GOOS" == "windows" ]; then
    OUTPUT_FILE="$PROJECT_NAME.exe"
  else
    OUTPUT_FILE="$PROJECT_NAME"
  fi

  echo "Downloading dependencies..."
  go mod tidy

  echo "Building project for $GOOS/$GOARCH..."
  GOOS="$GOOS" GOARCH="$GOARCH" go build -o "$BUILD_DIR/$OUTPUT_FILE"
}

copy_binary() {
  mkdir -p "$INSTALL_DIR"

  echo "Copying binary to $INSTALL_DIR..."

  local target="$INSTALL_DIR/$OUTPUT_FILE"

  if [ "$GOOS" != "windows" ] && [ ! -w "$INSTALL_DIR" ]; then
    echo "Using sudo to copy binary..."
    sudo cp "$BUILD_DIR/$OUTPUT_FILE" "$target"
  else
    cp "$BUILD_DIR/$OUTPUT_FILE" "$target"
  fi

  echo "Binary installed at: $target"
}

ensure_path_contains_install_dir() {
  if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo ""
    echo "⚠️  $INSTALL_DIR is not in your PATH."
    echo "To fix this, add the following line to your shell config:"
    if [ "$GOOS" = "windows" ]; then
      echo "  export PATH=\"\$HOME/bin:\$PATH\""  # For Git Bash
    else
      echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
    fi
    echo ""
  fi
}

verify_binary() {
  if ! "$INSTALL_DIR/$OUTPUT_FILE" -h >/dev/null 2>&1; then
    echo "❌ Failed to run the built binary."
    exit 1
  fi
}

main() {
  detect_arch
  detect_os
  set_install_dir
  build_binary
  copy_binary
  ensure_path_contains_install_dir
  verify_binary
}

main "$@"