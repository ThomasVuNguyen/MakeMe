#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd)"
DIST_DIR="$ROOT_DIR/dist"
VERSION="${VERSION:-}"
GO_CACHE_DIR="${GO_CACHE_DIR:-$ROOT_DIR/.cache/go-build}"
mkdir -p "$GO_CACHE_DIR"
export GOCACHE="$GO_CACHE_DIR"

for tool in go cargo rustc curl unzip; do
  if ! command -v "$tool" >/dev/null 2>&1; then
    echo "Missing required tool: $tool" >&2
    exit 1
  fi
done

RUNNER_ARCHIVE=""
RUNNER_URL=""
RUNNER_SUBDIR=""
RUNNER_LABEL=""

TARGET_ARG="${1-}"
if [[ -n "$TARGET_ARG" ]]; then
  case "$TARGET_ARG" in
    darwin-arm64|darwin-amd64|linux-amd64|linux-arm64|linux-arm|windows-amd64)
      IFS=- read -r GOOS GOARCH <<<"$TARGET_ARG"
      ;;
    *)
      echo "Unsupported target '$TARGET_ARG'." >&2
      echo "Use darwin-arm64, darwin-amd64, linux-amd64, linux-arm64, linux-arm, or windows-amd64." >&2
      exit 1
      ;;
  esac
else
  GOOS="${GOOS:-$(go env GOOS)}"
  GOARCH="${GOARCH:-$(go env GOARCH)}"
fi

if [[ -z "${GOOS:-}" || -z "${GOARCH:-}" ]]; then
  echo "Unable to determine GOOS/GOARCH." >&2
  exit 1
fi

case "${GOOS}-${GOARCH}" in
  darwin-arm64)
    RUST_TARGET="${RUST_TARGET:-aarch64-apple-darwin}"
    ;;
  darwin-amd64)
    RUST_TARGET="${RUST_TARGET:-x86_64-apple-darwin}"
    ;;
  linux-amd64)
    RUST_TARGET="${RUST_TARGET:-x86_64-unknown-linux-gnu}"
    ;;
  linux-arm64)
    RUST_TARGET="${RUST_TARGET:-aarch64-unknown-linux-gnu}"
    ;;
  linux-arm)
    RUST_TARGET="${RUST_TARGET:-armv7-unknown-linux-gnueabihf}"
    ;;
  windows-amd64)
    RUST_TARGET="${RUST_TARGET:-x86_64-pc-windows-msvc}"
    ;;
  *)
    RUST_TARGET="${RUST_TARGET:-}"
    ;;
esac

case "${GOOS}-${GOARCH}" in
  darwin-arm64)
    RUNNER_ARCHIVE="llamacpp-macos-arm64.zip"
    RUNNER_URL="https://huggingface.co/ThomasTheMaker/llamacpp/resolve/main/llamacpp-macos-arm64.zip"
    RUNNER_SUBDIR="runtime/darwin-arm64"
    RUNNER_LABEL="macOS arm64"
    ;;
  linux-arm64)
    RUNNER_ARCHIVE="llamacpp-rpi5.zip"
    RUNNER_URL="https://huggingface.co/ThomasTheMaker/llamacpp/resolve/main/llamacpp-rpi5.zip"
    RUNNER_SUBDIR="runtime/rpi5"
    RUNNER_LABEL="Raspberry Pi 5"
    ;;
esac

if [[ -z "$RUST_TARGET" ]]; then
  RUST_TARGET="$(rustc -Vv | awk '/^host: /{print $2; exit}')"
fi

if [[ -z "$RUST_TARGET" ]]; then
  echo "Unable to determine Rust target triple. Set RUST_TARGET explicitly." >&2
  exit 1
fi

if command -v rustup >/dev/null 2>&1; then
  if ! rustup target list --installed | grep -Fxq "$RUST_TARGET"; then
    echo "Installing Rust target $RUST_TARGET..."
    rustup target add "$RUST_TARGET"
  fi
fi

PACKAGE_BASENAME="makeme-${GOOS}-${GOARCH}"
PACKAGE_DIR="$DIST_DIR/$PACKAGE_BASENAME"
rm -rf "$PACKAGE_DIR"
mkdir -p "$PACKAGE_DIR"

if [[ -z "$VERSION" ]]; then
  if VERSION=$(git -C "$ROOT_DIR" describe --tags --always --dirty 2>/dev/null); then
    VERSION="${VERSION#v}"
  else
    VERSION="0.0.0"
  fi
fi

VERSION="${VERSION%-dirty}"

BIN_EXT=""
if [[ "$GOOS" == "windows" ]]; then
  BIN_EXT=".exe"
fi

echo "Building Go binaries for ${GOOS}/${GOARCH}..."
CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" go build -o "$PACKAGE_DIR/makeme$BIN_EXT" main.go stl.go
CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" go build -o "$PACKAGE_DIR/stl2obj$BIN_EXT" stl2obj.go stl.go

CARGO_TARGET_DIR="$ROOT_DIR/deps/terminal3d/target"

T3D_NAME="t3d"
if [[ "$GOOS" == "windows" ]]; then
  T3D_NAME="t3d.exe"
fi

echo "Building terminal3d for ${RUST_TARGET}..."
(
  cd "$ROOT_DIR/deps/terminal3d"
  cargo build --release --target "$RUST_TARGET"
)

T3D_SOURCE="$CARGO_TARGET_DIR/$RUST_TARGET/release/$T3D_NAME"
if [[ ! -f "$T3D_SOURCE" ]]; then
  echo "terminal3d binary not found at $T3D_SOURCE" >&2
  exit 1
fi
cp "$T3D_SOURCE" "$PACKAGE_DIR/$T3D_NAME"
chmod +x "$PACKAGE_DIR/$T3D_NAME"

mkdir -p "$PACKAGE_DIR/k"

if [[ -n "$RUNNER_URL" ]]; then
  echo "Packaging llama.cpp runtime (${RUNNER_LABEL})..."
  ARCHIVE_CACHE="$ROOT_DIR/k/$RUNNER_ARCHIVE"
  if [[ ! -f "$ARCHIVE_CACHE" ]]; then
    mkdir -p "$ROOT_DIR/k"
    curl -L "$RUNNER_URL" -o "$ARCHIVE_CACHE"
  fi

  DEST_DIR="$PACKAGE_DIR/k/$RUNNER_SUBDIR"
  rm -rf "$DEST_DIR"
  mkdir -p "$DEST_DIR"
  unzip -oq "$ARCHIVE_CACHE" -d "$DEST_DIR"

  PRIMARY_RUN=$(find "$DEST_DIR" -type f -name run | head -n 1)
  if [[ -n "$PRIMARY_RUN" ]]; then
    chmod +x "$PRIMARY_RUN"
    cp "$PRIMARY_RUN" "$PACKAGE_DIR/k/run"
    chmod +x "$PACKAGE_DIR/k/run"
  fi
else
  if [[ -f "$ROOT_DIR/k/run" ]]; then
    cp "$ROOT_DIR/k/run" "$PACKAGE_DIR/k/run"
    chmod +x "$PACKAGE_DIR/k/run"
  else
    echo "Warning: No llama.cpp runtime configured for ${GOOS}/${GOARCH}." >&2
  fi
fi

cp "$ROOT_DIR/README.md" "$PACKAGE_DIR/README.md"

ARCHIVE_NAME="$PACKAGE_BASENAME"
mkdir -p "$DIST_DIR"

if [[ "$GOOS" == "windows" ]]; then
  ARCHIVE_PATH="$DIST_DIR/${ARCHIVE_NAME}.zip"
  rm -f "$ARCHIVE_PATH"
  (
    cd "$DIST_DIR"
    zip -qr "${ARCHIVE_NAME}.zip" "$ARCHIVE_NAME"
  )
  echo "Created $ARCHIVE_PATH"
else
  ARCHIVE_PATH="$DIST_DIR/${ARCHIVE_NAME}.tar.gz"
  rm -f "$ARCHIVE_PATH"
  (
    cd "$DIST_DIR"
    tar -czf "${ARCHIVE_NAME}.tar.gz" "$ARCHIVE_NAME"
  )
  echo "Created $ARCHIVE_PATH"
fi

maybe_build_deb() {
  local goarch="$1"
  local pkg_dir="$2"

  if [[ "$GOOS" != "linux" ]]; then
    return
  fi

  if ! command -v dpkg-deb >/dev/null 2>&1; then
    echo "dpkg-deb not found; skipping .deb creation." >&2
    return
  fi

  local deb_arch
  case "$goarch" in
    amd64) deb_arch="amd64" ;;
    arm64) deb_arch="arm64" ;;
    arm)   deb_arch="armhf" ;;
    *)
      echo "Unsupported GOARCH '$goarch' for Debian package; skipping." >&2
      return
      ;;
  esac

  local stage_dir
  stage_dir="$(mktemp -d "$DIST_DIR/makeme-deb.XXXXXX")"
  trap 'rm -rf "$stage_dir"' EXIT

  local install_prefix="$stage_dir/opt/makeme"
  mkdir -p "$install_prefix"
  cp -R "$pkg_dir"/. "$install_prefix"

  mkdir -p "$stage_dir/usr/bin"
  cat >"$stage_dir/usr/bin/makeme" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
APP_DIR="/opt/makeme"
cd "$APP_DIR"
exec "$APP_DIR/makeme" "$@"
EOF
  chmod 0755 "$stage_dir/usr/bin/makeme"

  if [[ -f "$install_prefix/stl2obj" ]]; then
    cat >"$stage_dir/usr/bin/stl2obj" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
APP_DIR="/opt/makeme"
exec "$APP_DIR/stl2obj" "$@"
EOF
    chmod 0755 "$stage_dir/usr/bin/stl2obj"
  fi

  if [[ -f "$install_prefix/t3d" ]]; then
    cat >"$stage_dir/usr/bin/t3d" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
APP_DIR="/opt/makeme"
exec "$APP_DIR/t3d" "$@"
EOF
    chmod 0755 "$stage_dir/usr/bin/t3d"
  fi

  local debian_dir="$stage_dir/DEBIAN"
  mkdir -p "$debian_dir"

  local installed_size
  installed_size=$(du -sk "$stage_dir/opt/makeme" | awk '{print $1}')

  cat >"$debian_dir/control" <<EOF
Package: makeme
Version: $VERSION
Section: utils
Priority: optional
Architecture: $deb_arch
Maintainer: MakeMe Developers <support@makeme.local>
Depends: libc6, libstdc++6, bash
Installed-Size: $installed_size
Description: AI-powered 3D object generator CLI
 MakeMe turns natural language prompts into rendered 3D models using OpenSCAD.
EOF

  local deb_path="$DIST_DIR/makeme_${VERSION}_${deb_arch}.deb"
  dpkg-deb --build "$stage_dir" "$deb_path"
  echo "Created $deb_path"

  rm -rf "$stage_dir"
  trap - EXIT
}

maybe_build_deb "$GOARCH" "$PACKAGE_DIR"
