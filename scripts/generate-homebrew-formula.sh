#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd)"
DIST_DIR="$ROOT_DIR/dist"
FORMULA_DIR="${FORMULA_DIR:-$ROOT_DIR/packaging/homebrew}"
FORMULA_PATH="$FORMULA_DIR/makeme.rb"

VERSION="${VERSION:-}"
if [[ -z "$VERSION" ]]; then
  if VERSION=$(git -C "$ROOT_DIR" describe --tags --always --dirty 2>/dev/null); then
    VERSION="${VERSION#v}"
  else
    VERSION="0.0.0"
  fi
fi

VERSION="${VERSION%-dirty}"

BASE_URL_DEFAULT='https://github.com/ThomasVuNguyen/MakeMe/releases/download/v#{version}'
BASE_URL="${HOMEBREW_TAR_BASE_URL:-$BASE_URL_DEFAULT}"

mkdir -p "$FORMULA_DIR"

calc_sha256() {
  local file="$1"
  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$file" | awk '{print $1}'
  elif command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$file" | awk '{print $1}'
  else
    echo "No SHA256 tool found (need shasum or sha256sum)" >&2
    exit 1
  fi
}

arm_archive="$DIST_DIR/makeme-darwin-arm64.tar.gz"
amd_archive="$DIST_DIR/makeme-darwin-amd64.tar.gz"

arm_sha=""
amd_sha=""

if [[ -f "$arm_archive" ]]; then
  arm_sha=$(calc_sha256 "$arm_archive")
fi

if [[ -f "$amd_archive" ]]; then
  amd_sha=$(calc_sha256 "$amd_archive")
fi

if [[ -z "$arm_sha" && -z "$amd_sha" ]]; then
  echo "No macOS archives found in $DIST_DIR" >&2
  exit 1
fi

{
  cat <<EOF
class Makeme < Formula
  desc "AI-powered 3D object generator CLI"
  homepage "https://github.com/ThomasVuNguyen/MakeMe"
  version "$VERSION"

  on_macos do
EOF

  if [[ -n "$arm_sha" ]]; then
    cat <<EOF
    on_arm do
      url "$BASE_URL/makeme-darwin-arm64.tar.gz"
      sha256 "$arm_sha"
    end
EOF
  fi

  if [[ -n "$amd_sha" ]]; then
    cat <<EOF
    on_intel do
      url "$BASE_URL/makeme-darwin-amd64.tar.gz"
      sha256 "$amd_sha"
    end
EOF
  fi

  cat <<'EOF'
  end

  def install
    libexec.install Dir["*"]

    %w[makeme stl2obj t3d].each do |exe|
      next unless (libexec/exe).exist?
      (bin/exe).write <<~EOS
        #!/bin/bash
        set -euo pipefail
        cd "#{libexec}"
        exec "#{libexec}/#{exe}" "$@"
      EOS
    end
  end

  def caveats
    <<~EOS
      Model assets are stored under "k/" relative to your working directory.
      Consider setting MAKEME_T3D if you install a different viewer.
    EOS
  end
end
EOF
} >"$FORMULA_PATH"

echo "Wrote Homebrew formula to $FORMULA_PATH"
