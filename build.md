# Build Guide

Quick commands to produce release artifacts for Linux and macOS. Run everything from the repo root.

## Linux Build & Release (run on the Linux build machine)

### Build linux/amd64

```bash
GOOS=linux GOARCH=amd64 scripts/package.sh linux-amd64 && \
  latest_deb=$(ls -t dist/makeme_*_amd64.deb | head -n1) && \
  mv "$latest_deb" dist/makeme-linux-x86.deb

# Outputs:
#   dist/makeme-linux-amd64.tar.gz
#   dist/makeme-linux-x86.deb
```

### Build linux/arm64 (Raspberry Pi)

```bash
# One-time toolchain prep
rustup target add aarch64-unknown-linux-gnu
sudo apt install gcc-aarch64-linux-gnu
export CC_aarch64_unknown_linux_gnu=aarch64-linux-gnu-gcc
export CARGO_TARGET_AARCH64_UNKNOWN_LINUX_GNU_LINKER=aarch64-linux-gnu-gcc
export AR_aarch64_unknown_linux_gnu=aarch64-linux-gnu-ar

# Build
GOOS=linux GOARCH=arm64 scripts/package.sh linux-arm64 && \
  latest_deb=$(ls -t dist/makeme_*_arm64.deb | head -n1) && \
  mv "$latest_deb" dist/makeme-linux-arm64.deb

# Outputs:
#   dist/makeme-linux-arm64.tar.gz
#   dist/makeme-linux-arm64.deb
```

If you want to sanity-check the packages on the build box, install with:

```bash
sudo dpkg -i dist/makeme-linux-x86.deb   # swap to arm64 deb when testing Pi builds
sudo apt-get install -f                # only if dpkg reports missing dependencies
```

Long term, publishing via an apt repository gives users `apt install makeme`; otherwise point them at the manual `.deb` or tarball install instructions in `README.md`.

### Publish Linux artifacts to GitHub Releases

```bash
export VERSION=1.0.0               # adjust per release
export RELEASE_TAG="v$VERSION"

# Attach Linux artifacts to the release created from macOS
gh release upload "$RELEASE_TAG" \
  dist/makeme-linux-amd64.tar.gz \
  dist/makeme-linux-arm64.tar.gz \
  dist/makeme-linux-x86.deb \
  dist/makeme-linux-arm64.deb \
  --clobber

# If the release doesn’t exist yet, include the mac tarballs here too:
# gh release create "$RELEASE_TAG" dist/makeme-* --title "MakeMe $VERSION" --notes "Release $VERSION"
```

## macOS Build & Release (run on the macOS build machine)

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
gh release create "v$VERSION" \
  dist/makeme-darwin-arm64.tar.gz \
  dist/makeme-darwin-amd64.tar.gz \
  --title "MakeMe $VERSION" \
  --notes "Release $VERSION"

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
