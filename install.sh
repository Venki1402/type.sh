#!/bin/sh
# type.sh installer — downloads a prebuilt binary from GitHub Releases.
#
#   curl -fsSL https://raw.githubusercontent.com/Venki1402/type.sh/main/install.sh | sh
#
# Env overrides:
#   VERSION=v0.1.0   install a specific tag (default: latest release)
#   BINDIR=~/.local/bin   install location (default: /usr/local/bin, or
#                         ~/.local/bin if /usr/local/bin isn't writable)
set -eu

REPO="Venki1402/type.sh"
BIN="typesh"

err() { echo "type.sh install: $*" >&2; exit 1; }

# --- detect OS ---------------------------------------------------------------
os=$(uname -s)
case "$os" in
  Darwin) OS=darwin ;;
  Linux)  OS=linux ;;
  *) err "unsupported OS '$os' (Homebrew or building from source may work)" ;;
esac

# --- detect arch -------------------------------------------------------------
arch=$(uname -m)
case "$arch" in
  x86_64|amd64) ARCH=amd64 ;;
  arm64|aarch64) ARCH=arm64 ;;
  *) err "unsupported architecture '$arch'" ;;
esac

# --- pick a downloader -------------------------------------------------------
if command -v curl >/dev/null 2>&1; then
  dl() { curl -fsSL "$1" -o "$2"; }
  fetch() { curl -fsSL "$1"; }
elif command -v wget >/dev/null 2>&1; then
  dl() { wget -qO "$2" "$1"; }
  fetch() { wget -qO - "$1"; }
else
  err "need curl or wget installed"
fi

# --- resolve version ---------------------------------------------------------
VERSION="${VERSION:-}"
if [ -z "$VERSION" ]; then
  api="https://api.github.com/repos/$REPO/releases/latest"
  VERSION=$(fetch "$api" | grep '"tag_name":' | head -n1 | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
  [ -n "$VERSION" ] || err "could not determine latest release; set VERSION=vX.Y.Z"
fi
# GoReleaser archive names use the version without the leading 'v'.
NUM=$(echo "$VERSION" | sed 's/^v//')

ARCHIVE="${BIN}_${NUM}_${OS}_${ARCH}.tar.gz"
BASE="https://github.com/$REPO/releases/download/$VERSION"

# --- download + verify -------------------------------------------------------
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

echo "Downloading $ARCHIVE ($VERSION)..."
dl "$BASE/$ARCHIVE" "$TMP/$ARCHIVE" || err "download failed: $BASE/$ARCHIVE"

if dl "$BASE/checksums.txt" "$TMP/checksums.txt" 2>/dev/null; then
  echo "Verifying checksum..."
  ( cd "$TMP" && grep " $ARCHIVE\$" checksums.txt | (sha256sum -c - 2>/dev/null || shasum -a 256 -c -) ) \
    || err "checksum verification failed"
fi

tar -xzf "$TMP/$ARCHIVE" -C "$TMP"
[ -f "$TMP/$BIN" ] || err "binary '$BIN' not found in archive"
chmod +x "$TMP/$BIN"

# --- install -----------------------------------------------------------------
BINDIR="${BINDIR:-/usr/local/bin}"
if [ ! -d "$BINDIR" ] || [ ! -w "$BINDIR" ]; then
  if [ -w "$(dirname "$BINDIR")" ] 2>/dev/null; then :; fi
  if command -v sudo >/dev/null 2>&1 && [ -d "$BINDIR" ]; then
    echo "Installing to $BINDIR (needs sudo)..."
    sudo install -m 0755 "$TMP/$BIN" "$BINDIR/$BIN"
  else
    BINDIR="$HOME/.local/bin"
    mkdir -p "$BINDIR"
    echo "Installing to $BINDIR..."
    install -m 0755 "$TMP/$BIN" "$BINDIR/$BIN"
    case ":$PATH:" in
      *":$BINDIR:"*) ;;
      *) echo "NOTE: add $BINDIR to your PATH, e.g. export PATH=\"$BINDIR:\$PATH\"" ;;
    esac
  fi
else
  install -m 0755 "$TMP/$BIN" "$BINDIR/$BIN"
fi

echo "Installed $BIN to $BINDIR. Run: $BIN"
