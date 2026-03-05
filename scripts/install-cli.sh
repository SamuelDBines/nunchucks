#!/usr/bin/env sh

set -eu

repo="SamuelDBines/nunjucks"
install_dir="${NUNCHUCKS_INSTALL_DIR:-$HOME/.local/bin}"
version="${1:-latest}"

os="$(uname -s)"
arch="$(uname -m)"

case "$os" in
  Darwin) goos="darwin" ;;
  Linux) goos="linux" ;;
  *)
    echo "unsupported operating system: $os" >&2
    exit 1
    ;;
esac

case "$arch" in
  x86_64|amd64) goarch="amd64" ;;
  arm64|aarch64) goarch="arm64" ;;
  *)
    echo "unsupported architecture: $arch" >&2
    exit 1
    ;;
esac

if [ "$version" = "latest" ]; then
  version="$(curl -fsSLI -o /dev/null -w '%{url_effective}' "https://github.com/$repo/releases/latest" | awk -F/ '{print $NF}')"
fi

asset="nunchucks_${version}_${goos}_${goarch}.tar.gz"
url="https://github.com/$repo/releases/download/${version}/${asset}"

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT INT TERM

mkdir -p "$install_dir"
curl -fsSL "$url" -o "$tmpdir/$asset"
tar -xzf "$tmpdir/$asset" -C "$tmpdir"

cp "$tmpdir/nunchucks_${version}_${goos}_${goarch}/nunchucks" "$install_dir/nunchucks"
chmod +x "$install_dir/nunchucks"

echo "installed nunchucks ${version} to $install_dir/nunchucks"
