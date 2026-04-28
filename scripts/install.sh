#!/usr/bin/env sh
set -eu

REPO="pantheon-org/skill-quality-auditor"
BINARY="skill-auditor"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# ── helpers ────────────────────────────────────────────────────────────────────

info()  { printf '  \033[34m•\033[0m %s\n' "$*"; }
ok()    { printf '  \033[32m✓\033[0m %s\n' "$*"; }
err()   { printf '  \033[31m✗\033[0m %s\n' "$*" >&2; exit 1; }

need() {
  for cmd in "$@"; do
    command -v "$cmd" >/dev/null 2>&1 || err "Required tool not found: $cmd"
  done
}

# ── detect platform ────────────────────────────────────────────────────────────

detect_os() {
  case "$(uname -s)" in
    Linux*)  echo "linux"  ;;
    Darwin*) echo "darwin" ;;
    *)       err "Unsupported OS: $(uname -s)" ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64)          echo "amd64" ;;
    amd64)           echo "amd64" ;;
    aarch64|arm64)   echo "arm64" ;;
    *)               err "Unsupported architecture: $(uname -m)" ;;
  esac
}

# ── resolve latest release ─────────────────────────────────────────────────────

latest_version() {
  url="https://api.github.com/repos/${REPO}/releases/latest"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url" \
      | grep '"tag_name"' \
      | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/'
  elif command -v wget >/dev/null 2>&1; then
    wget -qO- "$url" \
      | grep '"tag_name"' \
      | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/'
  else
    err "curl or wget is required"
  fi
}

download() {
  url="$1"
  dest="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL -o "$dest" "$url"
  else
    wget -qO "$dest" "$url"
  fi
}

# ── verify checksum ────────────────────────────────────────────────────────────

verify_checksum() {
  file="$1"
  checksums_file="$2"
  filename="$(basename "$file")"

  expected=$(grep " ${filename}$" "$checksums_file" | awk '{print $1}')
  if [ -z "$expected" ]; then
    err "No checksum found for ${filename} in checksums.txt"
  fi

  if command -v sha256sum >/dev/null 2>&1; then
    actual=$(sha256sum "$file" | awk '{print $1}')
  elif command -v shasum >/dev/null 2>&1; then
    actual=$(shasum -a 256 "$file" | awk '{print $1}')
  else
    err "sha256sum or shasum is required for checksum verification"
  fi

  if [ "$actual" != "$expected" ]; then
    err "Checksum mismatch for ${filename}: expected ${expected}, got ${actual}"
  fi
}

# ── install ────────────────────────────────────────────────────────────────────

main() {
  VERSION="${VERSION:-}"
  need grep sed awk

  OS=$(detect_os)
  ARCH=$(detect_arch)

  if [ -z "$VERSION" ]; then
    info "Resolving latest release..."
    VERSION=$(latest_version)
    [ -z "$VERSION" ] && err "Could not determine latest release version"
  fi

  VERSION_TAG="$VERSION"
  ARCHIVE="${BINARY}_${OS}_${ARCH}.tar.gz"
  BASE_URL="https://github.com/${REPO}/releases/download/${VERSION_TAG}"

  TMP=$(mktemp -d)
  trap 'rm -rf "$TMP"' EXIT

  info "Downloading ${BINARY} ${VERSION_TAG} (${OS}/${ARCH})..."
  download "${BASE_URL}/${ARCHIVE}" "${TMP}/${ARCHIVE}"

  info "Downloading checksums..."
  download "${BASE_URL}/checksums.txt" "${TMP}/checksums.txt"

  info "Verifying checksum..."
  verify_checksum "${TMP}/${ARCHIVE}" "${TMP}/checksums.txt"
  ok "Checksum verified"

  info "Extracting..."
  tar -xzf "${TMP}/${ARCHIVE}" -C "$TMP"

  if [ ! -f "${TMP}/${BINARY}" ]; then
    err "Binary '${BINARY}' not found in archive"
  fi

  info "Installing to ${INSTALL_DIR}/${BINARY}..."
  if [ -w "$INSTALL_DIR" ]; then
    mv "${TMP}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
    chmod +x "${INSTALL_DIR}/${BINARY}"
  else
    sudo mv "${TMP}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
    sudo chmod +x "${INSTALL_DIR}/${BINARY}"
  fi

  ok "Installed ${BINARY} ${VERSION_TAG}"
  printf '\nRun: %s --help\n' "$BINARY"
}

main "$@"
