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
- **OpenSCAD** - Install via `brew install openscad`
- **Terminal3d** - Install via `brew install liam-ilan/terminal3d/terminal3d`

## 🚀 Quick Start

### 1. Build the Application

```bash
# Clone the repository
git clone <repository-url>
cd MakeMe

# Build the main application
go build -o makeme main.go stl.go

# Build the STL to OBJ converter
go build -o stl2obj stl2obj.go stl.go
```

### 2. Run MakeMe

```bash
./makeme
```

### 3. Create Something Amazing!

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
│   └── run             # Model runner binary
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
- Ensure the `k/` directory contains `k-1b-q8_0.gguf` and `run` binary
- Check that the `run` binary has execute permissions: `chmod +x k/run`

### OpenSCAD errors
- Verify OpenSCAD is installed: `openscad --version`
- Check that the AI generated valid SCAD code in `k/output.scad`

### 3D viewer issues
- Ensure terminal3d is installed: `t3d --version`
- Make sure your terminal supports mouse input and has sufficient size

### No output or stuck
- The AI model may take time to load on first run
- Check for error messages in red at the top of the interface

## 🤝 Contributing

Feel free to submit issues and enhancement requests!

## 📝 License

This project is open source and available under the MIT License.