#!/usr/bin/env sh

set -eu

repo="SamuelDBines/nunjucks"
version="${1:-latest}"

fetch() {
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$@"
    return
  fi
  if command -v wget >/dev/null 2>&1; then
    wget -qO- "$1"
    return
  fi
  echo "curl or wget is required" >&2
  exit 1
}

fetch_to_file() {
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$1" -o "$2"
    return
  fi
  if command -v wget >/dev/null 2>&1; then
    wget -qO "$2" "$1"
    return
  fi
  echo "curl or wget is required" >&2
  exit 1
}

pick_install_dir() {
  for candidate in "${NUNCHUCKS_INSTALL_DIR:-}" "$HOME/.local/bin" "/usr/local/bin" "/opt/homebrew/bin"; do
    if [ -z "${candidate}" ]; then
      continue
    fi
    if [ -d "$candidate" ] && [ -w "$candidate" ]; then
      printf '%s\n' "$candidate"
      return
    fi
  done

  printf '%s\n' "$HOME/.local/bin"
}

is_in_path() {
  case ":$PATH:" in
    *:"$1":*) return 0 ;;
    *) return 1 ;;
  esac
}

print_path_hint() {
  shell_name="$(basename "${SHELL:-sh}")"
  case "$shell_name" in
    zsh)
      printf 'Add this to ~/.zshrc and restart your shell:\n  export PATH="%s:$PATH"\n' "$1"
      ;;
    bash)
      printf 'Add this to ~/.bashrc and restart your shell:\n  export PATH="%s:$PATH"\n' "$1"
      ;;
    fish)
      printf 'Add this to your Fish config:\n  fish_add_path %s\n' "$1"
      ;;
    *)
      printf 'Add %s to your PATH, then restart your shell.\n' "$1"
      ;;
  esac
}

install_dir="$(pick_install_dir)"

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
  if command -v curl >/dev/null 2>&1; then
    version="$(curl -fsSLI -o /dev/null -w '%{url_effective}' "https://github.com/$repo/releases/latest" | awk -F/ '{print $NF}')"
  else
    version="$(fetch "https://api.github.com/repos/$repo/releases/latest" | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -n 1)"
  fi
fi

asset="nunchucks_${version}_${goos}_${goarch}.tar.gz"
url="https://github.com/$repo/releases/download/${version}/${asset}"

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT INT TERM

mkdir -p "$install_dir"
fetch_to_file "$url" "$tmpdir/$asset"
tar -xzf "$tmpdir/$asset" -C "$tmpdir"

cp "$tmpdir/nunchucks_${version}_${goos}_${goarch}/nunchucks" "$install_dir/nunchucks"
chmod +x "$install_dir/nunchucks"

echo "installed nunchucks ${version} to $install_dir/nunchucks"

if "$install_dir/nunchucks" version >/dev/null 2>&1; then
  echo "verified: $install_dir/nunchucks version"
fi

if is_in_path "$install_dir"; then
  echo "run: nunchucks help"
else
  print_path_hint "$install_dir"
  echo "or run directly:"
  echo "  $install_dir/nunchucks help"
fi
