# Build Guide

Quick commands to produce release artifacts for Linux and macOS. Run everything from the repo root.

## Linux

```bash
# Build tarball + .deb for linux/amd64 (requires Go, Rust, curl, unzip, dpkg-deb)
GOOS=linux GOARCH=amd64 scripts/package.sh linux-amd64

# Outputs:
#   dist/makeme-linux-amd64.tar.gz
#   dist/makeme_<version>_amd64.deb
```

```bash
# Build for linux/arm64
GOOS=linux GOARCH=arm64 scripts/package.sh linux-arm64

# Outputs:
#   dist/makeme-linux-arm64.tar.gz
#   dist/makeme_<version>_arm64.deb
```

## macOS

```bash
# Apple Silicon archive
GOOS=darwin GOARCH=arm64 scripts/package.sh darwin-arm64

# Intel archive
GOOS=darwin GOARCH=amd64 scripts/package.sh darwin-amd64

# Outputs:
#   dist/makeme-darwin-arm64.tar.gz
#   dist/makeme-darwin-amd64.tar.gz (if built)
```

After generating the macOS tarballs, refresh the Homebrew tap formula so SHA256 hashes stay in sync:

```bash
scripts/generate-homebrew-formula.sh
# -> packaging/homebrew/makeme.rb
```

From there, bump the release and publish it so Homebrew users can install:

```bash
# Pick a version and make sure the macOS archives in dist/ match it
export VERSION=1.2.3
VERSION="$VERSION" scripts/generate-homebrew-formula.sh

# Tag + push the repo release
git tag -a "v$VERSION" -m "MakeMe $VERSION"
git push origin "v$VERSION"

# Publish the GitHub release with the macOS tarball(s)
# (requires gh CLI; edit the command or use the web UI if you prefer)
gh release create "v$VERSION" dist/makeme-darwin-arm64.tar.gz \
  --title "MakeMe $VERSION" --notes "Release $VERSION"

# Update your Homebrew tap (set HOMEBREW_TAP_ORG + HOMEBREW_TAP_DIR)
export HOMEBREW_TAP_ORG=ThomasVuNguyen
export HOMEBREW_TAP_DIR=../homebrew-makeme
export HOMEBREW_TAP_REMOTE="https://github.com/${HOMEBREW_TAP_ORG}/homebrew-makeme.git"
if [ ! -d "$HOMEBREW_TAP_DIR/.git" ]; then
  rm -rf "$HOMEBREW_TAP_DIR"
  git clone "$HOMEBREW_TAP_REMOTE" "$HOMEBREW_TAP_DIR"
fi
install -d "$HOMEBREW_TAP_DIR/Formula"
cp packaging/homebrew/makeme.rb "$HOMEBREW_TAP_DIR/Formula/makeme.rb"
git -C "$HOMEBREW_TAP_DIR" add Formula/makeme.rb
git -C "$HOMEBREW_TAP_DIR" commit -m "makeme $VERSION"
git -C "$HOMEBREW_TAP_DIR" push

# Users can install once the tap update lands
brew tap ThomasVuNguyen/makeme
brew install makeme
```

> Tip: remove `dist/` before rebuilding if you want a clean artifact set.
