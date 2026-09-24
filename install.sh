#!/bin/sh
# tars installer — downloads the latest release binary for this OS/arch.
#
#   curl -fsSL https://raw.githubusercontent.com/cloudwalksolutions/machine-setup/main/install.sh | sh
#
# Installs to ~/.local/bin by default; override with TARS_INSTALL_DIR.
set -eu

REPO="cloudwalksolutions/machine-setup"
INSTALL_DIR="${TARS_INSTALL_DIR:-$HOME/.local/bin}"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$os" in
  darwin | linux) ;;
  *) echo "unsupported OS: $os" >&2 && exit 1 ;;
esac

arch=$(uname -m)
case "$arch" in
  x86_64) arch=amd64 ;;
  aarch64 | arm64) arch=arm64 ;;
  *) echo "unsupported architecture: $arch" >&2 && exit 1 ;;
esac

tag=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" |
  sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -1)
[ -n "$tag" ] || { echo "could not resolve the latest release" >&2 && exit 1; }
version=${tag#v}
asset="tars_${version}_${os}_${arch}.tar.gz"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

echo "Downloading tars $tag ($os/$arch)..."
curl -fsSL -o "$tmp/$asset" "https://github.com/$REPO/releases/download/$tag/$asset"
curl -fsSL -o "$tmp/checksums.txt" "https://github.com/$REPO/releases/download/$tag/checksums.txt"

if command -v sha256sum >/dev/null 2>&1; then
  sha="sha256sum"
else
  sha="shasum -a 256"
fi
(cd "$tmp" && grep " $asset\$" checksums.txt | $sha -c - >/dev/null) ||
  { echo "checksum verification failed" >&2 && exit 1; }

tar -xzf "$tmp/$asset" -C "$tmp" tars
mkdir -p "$INSTALL_DIR"
install -m 0755 "$tmp/tars" "$INSTALL_DIR/tars"
echo "Installed tars $tag to $INSTALL_DIR/tars"

case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *) echo "NOTE: $INSTALL_DIR is not on your PATH; add:  export PATH=\"$INSTALL_DIR:\$PATH\"" ;;
esac
"$INSTALL_DIR/tars" --version