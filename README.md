# MakeMe 🎨

**AI-Powered 3D Object Generator**

MakeMe is a CLI application that transforms natural language descriptions into beautiful 3D models. Simply describe what you want, and watch as AI generates OpenSCAD code, converts it to 3D formats, and displays stunning terminal-based 3D visualizations.

## ✨ Features

- **Natural Language Input**: Describe any object in plain English
- **AI-Powered Generation**: Uses a local k-1b quantized model for SCAD code generation
- **High-Quality 3D Output**: Converts SCAD → STL → OBJ with configurable resolution
- **Terminal 3D Visualization**: Beautiful interactive 3D rendering right in your terminal
- **Offline Operation**: Everything runs locally, no internet required

## 🛠️ Prerequisites

- **Go** (1.19 or later)
- **Rust** toolchain (cargo) to build the bundled Terminal3d viewer
- **OpenSCAD** - Install via `brew install openscad`
- **One-time internet** - The first run downloads the default `k-1b-q8_0.gguf` model and a llama.cpp runtime bundle if missing

## 🚀 Quick Start

### macOS (Homebrew)

```bash
brew install --cask openscad
brew tap ThomasVuNguyen/makeme
brew install makeme
makeme
```

### Install OpenSCAD (required)

```bash
# macOS (Homebrew)
brew install --cask openscad

# Debian/Ubuntu
sudo apt-get install openscad

# Arch Linux
sudo pacman -S openscad

# Windows / other platforms
# Download the installer from https://openscad.org/downloads.html
```

### Run the CLI

```bash
# Homebrew or PATH installs
makeme

# Extracted archive or local build
./makeme
```

Set `MAKEME_T3D` if you want to point the viewer to a custom `t3d` location; otherwise the app uses the bundled binary.

The first launch fetches `k/k-1b-q8_0.gguf` from Hugging Face and a llama.cpp runtime bundle (extracted under `k/runtime/<platform>/`) when either file is absent, so expect a short download.

Override the runner location with `MAKEME_RUN=/path/to/run` (the alias `MAKEME_LLAMAFILE` is still honoured for compatibility).

### Use a Packaged Release

```bash
# Download the appropriate archive (e.g., makeme-darwin-arm64.tar.gz)
tar -xzf makeme-<platform>.tar.gz
cd makeme-<platform>

# Run the app
./makeme
```

The archive already contains `t3d`, so no extra installation steps are needed.

### Create Something Amazing!

1. **Enter your description**: Type what you want to create (e.g., "a dragon", "a coffee mug", "a spaceship")
2. **Press Enter**: Watch the AI model generate OpenSCAD code
3. **View in 3D**: The terminal 3D viewer launches automatically
4. **Interact**: Use mouse to rotate, scroll to zoom, keyboard shortcuts for different modes

## 🎮 3D Viewer Controls

- **Mouse drag**: Rotate the model
- **Scroll**: Zoom in/out
- **Shift + drag**: Pan around
- **B**: Toggle between braille and block rendering
- **P**: Toggle between edges and vertices mode
- **Ctrl+C**: Exit viewer and return to prompt

## 📁 Project Structure

```
MakeMe/
├── main.go              # Main CLI application
├── stl.go               # STL parsing and 3D rendering utilities
├── stl2obj.go           # STL to OBJ converter
├── k/                   # AI model directory
│   ├── k-1b-q8_0.gguf  # Quantized AI model
│   └── runtime/        # Downloaded llama.cpp bundles (e.g., runtime/darwin-arm64/run)
└── deps/terminal3d/     # 3D terminal viewer (source)
```

## 🔧 How It Works

1. **Natural Language Processing**: Your description is sent to the local k-1b AI model
2. **SCAD Generation**: The AI generates OpenSCAD code for your object
3. **3D Conversion**: OpenSCAD renders the code to STL format
4. **Format Conversion**: Custom Go converter transforms STL to OBJ
5. **Visualization**: Terminal3d displays the interactive 3D model

## 🎯 Example Objects

Try describing these objects:
- "a cat" - Generates a detailed cat model
- "a box" - Creates a simple cube
- "a house" - Builds a basic house structure
- "a cake" - Makes a layered cake
- "a dragon" - Creates a fantasy dragon
- "a coffee mug" - Designs a functional mug

## ⚙️ Configuration

The application uses high-resolution settings by default:
- **$fn=100**: 100 fragments for smooth curves
- **$fs=0.1**: Minimum fragment size of 0.1mm
- **$fa=1**: 1-degree minimum fragment angle

## 🐛 Troubleshooting

### Model doesn't load
- Ensure the `k/` directory contains `k-1b-q8_0.gguf` and a llama.cpp runner (e.g., `k/runtime/darwin-arm64/run`)
- Check that the `run` binary has execute permissions: `chmod +x k/runtime/<platform>/run`

### OpenSCAD errors
- Verify OpenSCAD is installed: `openscad --version`
- Check that the AI generated valid SCAD code in `k/output.scad`

### 3D viewer issues
- Ensure the bundled binary exists: `deps/terminal3d/target/release/t3d`
- (Optional) point `MAKEME_T3D` to a custom viewer path
- Make sure your terminal supports mouse input and has sufficient size

### No output or stuck
- The AI model may take time to load on first run
- Check for error messages in red at the top of the interface

## 📦 Release Bundles

Create self-contained archives that include MakeMe, the STL converter, and the Terminal3d viewer.

```
# Build for the current platform
scripts/package.sh

# Or specify a target (darwin-arm64, darwin-amd64, linux-amd64, linux-arm64, windows-amd64)
scripts/package.sh darwin-arm64
```

### Debian / Ubuntu (`.deb`)

When you target Linux and have `dpkg-deb` available the packaging script now produces a `.deb` alongside the tarball:

```
scripts/package.sh linux-amd64
# => dist/makeme_<version>_amd64.deb
```

Publish the `.deb` through an apt repository (e.g. `reprepro`, `aptly`, Cloudsmith). Users add your repo, run `apt update`, and install with `sudo apt install makeme`. The package installs the full bundle under `/opt/makeme` and adds wrappers in `/usr/bin` so binaries still find their relative assets.

### Homebrew (macOS)

Turn the generated tarballs into a tap formula with:

```
scripts/package.sh darwin-arm64
scripts/package.sh darwin-amd64   # optional if you ship Intel builds
scripts/generate-homebrew-formula.sh
# => packaging/homebrew/makeme.rb
```

Commit the formula to your tap (`homebrew-makeme`, for example) and push. Homebrew installs into `libexec` and adds shims that change into that directory before launching, keeping the bundled viewer and llama.cpp runtime reachable.

Artifacts are written to `dist/`. At runtime MakeMe automatically looks for a `t3d` binary next to the executable, inside `deps/terminal3d/target/release/`, or anywhere on `PATH`. Set `MAKEME_T3D=/custom/path/to/t3d` to point to an alternative viewer if needed. The packaging script also fetches and unpacks the llama.cpp runtime bundle for supported targets when it is not already cached under `k/`.

The GGUF model is not bundled inside the archives; if `k/k-1b-q8_0.gguf` is missing the application downloads it from Hugging Face on demand.

Each archive ships with the matching llama.cpp runtime under `k/runtime/<platform>/` and a convenience copy at `k/run`. Point `MAKEME_RUN` (or `MAKEME_LLAMAFILE`) to a custom build if you need special hardware acceleration.

## 🔨 Optional: Build from Source

```bash
# Clone the repository
git clone <repository-url>
cd MakeMe

# Build the bundled terminal3d viewer (once per target)
cargo build --release --manifest-path deps/terminal3d/Cargo.toml

# Build the primary binaries
go build -o makeme main.go stl.go
go build -o stl2obj stl2obj.go stl.go
```

Use this path if you want to hack on the codebase or build for a platform without published release artifacts.

## 🤝 Contributing

Feel free to submit issues and enhancement requests!

## 📝 License

This project is open source and available under the MIT License.
