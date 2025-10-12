package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type state int

const (
	stateInput state = iota
	stateGenerating
	stateConverting
)

type model struct {
	state          state
	textInput      textinput.Model
	selectedObject string
	width          int
	height         int
	generatingMsg  string
	err            error
}

const (
	terminal3dEnvOverride = "MAKEME_T3D"
	modelDirectory        = "k"
	modelFilename         = "k-1b-q8_0.gguf"
	modelDownloadURL      = "https://huggingface.co/ThomasTheMaker/k-1b-gguf/resolve/main/k-1b-q8_0.gguf"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF79C6")).
			MarginBottom(1)

	promptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8BE9FD")).
			MarginBottom(1)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#BD93F9")).
			Padding(1, 2).
			Width(50)

	objectStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("#50FA7B")).
			Padding(2).
			MarginTop(1).
			MarginBottom(1).
			Align(lipgloss.Center)

	buttonStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#44475A")).
			Foreground(lipgloss.Color("#F8F8F2")).
			Padding(0, 2).
			Margin(0, 1)

	activeButtonStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#FF79C6")).
				Foreground(lipgloss.Color("#282A36")).
				Padding(0, 2).
				Margin(0, 1).
				Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4")).
			MarginTop(1)
)

type runnerPackage struct {
	url           string
	archiveName   string
	extractSubdir string
	label         string
}

var (
	runnerPackages = map[string]runnerPackage{
		"darwin/arm64": {
			url:           "https://huggingface.co/ThomasTheMaker/llamacpp/resolve/main/llamacpp-macos-arm64.zip",
			archiveName:   "llamacpp-macos-arm64.zip",
			extractSubdir: filepath.Join("runtime", "darwin-arm64"),
			label:         "macOS arm64",
		},
		"linux/arm64": {
			url:           "https://huggingface.co/ThomasTheMaker/llamacpp/resolve/main/llamacpp-rpi5.zip",
			archiveName:   "llamacpp-rpi5.zip",
			extractSubdir: filepath.Join("runtime", "rpi5"),
			label:         "Raspberry Pi 5",
		},
	}
)

var runnerOverrideEnvVars = []string{"MAKEME_RUN", "MAKEME_LLAMAFILE"}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "a cat, a box, a house, a cake..."
	ti.Focus()
	ti.CharLimit = 50
	ti.Width = 46

	return model{
		state:     stateInput,
		textInput: ti,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case generateObjectMsg:
		if msg.err != nil {
			m.err = msg.err
			m.state = stateInput
			m.textInput.Focus()
			return m, textinput.Blink
		}

		t3dPath, err := findTerminal3DExecutable()
		if err != nil {
			m.err = err
			m.state = stateInput
			m.textInput.Focus()
			return m, textinput.Blink
		}

		// Display OBJ file using terminal3d
		// Exit the current app and launch terminal3d directly
		fmt.Printf("\n🎯 Opening 3D model: %s\n", msg.stlPath)
		fmt.Printf("Press Ctrl+C to return when done viewing.\n\n")

		cmd := exec.Command(t3dPath, msg.stlPath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Printf("Error launching terminal3d: %v\n", err)
		}

		fmt.Printf("\nReturning to MakeMe...\n")
		return m, tea.Quit

		// Return to input for next generation
		m.state = stateInput
		m.textInput.Focus()
		return m, textinput.Blink

	case tea.KeyMsg:
		switch m.state {
		case stateInput:
			switch msg.Type {
			case tea.KeyEnter:
				objectName := strings.TrimSpace(m.textInput.Value())
				if objectName != "" {
					m.selectedObject = objectName
					m.state = stateGenerating
					m.generatingMsg = "Generating " + objectName + "..."
					m.err = nil // Clear any previous errors
					m.textInput.Reset()
					return m, generateObjectCmd(objectName)
				}
			case tea.KeyCtrlC, tea.KeyEsc:
				return m, tea.Quit
			}

		case stateGenerating:
			switch msg.String() {
			case "ctrl+c", "esc":
				return m, tea.Quit
			}

		case stateConverting:
			switch msg.String() {
			case "ctrl+c", "esc":
				return m, tea.Quit
			}
		}
	}

	if m.state == stateInput {
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("🎨 MakeMe - 3D Object Creator") + "\n\n")

	// Display error if any
	if m.err != nil {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555")).
			Bold(true)
		s.WriteString(errorStyle.Render("Error: "+m.err.Error()) + "\n\n")
	}

	switch m.state {
	case stateInput:
		s.WriteString(promptStyle.Render("What would you like me to make?") + "\n")
		s.WriteString(helpStyle.Render("Try: a cat, a box, a house, a cake, or anything else!") + "\n\n")
		s.WriteString(inputStyle.Render(m.textInput.View()) + "\n\n")
		s.WriteString(helpStyle.Render("Press Enter to generate • Esc to quit"))

	case stateGenerating:
		s.WriteString(promptStyle.Render(m.generatingMsg) + "\n\n")
		s.WriteString(objectStyle.Width(60).Height(10).Render("\n\n⚙️  Running AI model...\n\nThis may take a moment...") + "\n")
		s.WriteString(helpStyle.Render("Please wait..."))

	case stateConverting:
		s.WriteString(promptStyle.Render("Converting to STL...") + "\n\n")
		s.WriteString(objectStyle.Width(60).Height(10).Render("\n\n🔧 Converting SCAD to STL...\n\nAlmost there...") + "\n")
		s.WriteString(helpStyle.Render("Please wait..."))
	}

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		s.String(),
	)
}

type generateObjectMsg struct {
	stlPath string
	err     error
}

func generateObjectCmd(objectName string) tea.Cmd {
	return func() tea.Msg {
		// Create k directory if it doesn't exist
		if err := os.MkdirAll(modelDirectory, 0755); err != nil {
			return generateObjectMsg{err: fmt.Errorf("failed to create k directory: %w", err)}
		}

		runnerPath, err := ensureModelRunner()
		if err != nil {
			return generateObjectMsg{err: err}
		}

		if err := ensureModelAssets(); err != nil {
			return generateObjectMsg{err: err}
		}

		// Run the AI model to generate SCAD code
		prompt := fmt.Sprintf("Hey cadmonkey, make me %s", objectName)
		cmd := exec.Command(runnerPath, "-m", "k-1b-q8_0.gguf", "-p", prompt)
		cmd.Dir = modelDirectory

		// Capture both stdout and stderr
		output, err := cmd.CombinedOutput()
		if err != nil {
			return generateObjectMsg{err: fmt.Errorf("failed to run model: %w\nOutput: %s", err, string(output))}
		}

		if len(output) == 0 {
			return generateObjectMsg{err: fmt.Errorf("model produced no output")}
		}

		// Clean model output - remove prompt echo and 'model' line
		cleanOutput := cleanModelOutput(string(output), objectName)

		if strings.TrimSpace(cleanOutput) == "" {
			return generateObjectMsg{err: fmt.Errorf("no valid SCAD code found after cleaning. Raw output:\n%s", string(output))}
		}

		// Save cleaned model output as SCAD file
		scadPath := filepath.Join(modelDirectory, "output.scad")
		if err := os.WriteFile(scadPath, []byte(cleanOutput), 0644); err != nil {
			return generateObjectMsg{err: fmt.Errorf("failed to write SCAD file: %w", err)}
		}

		// Convert SCAD to STL using OpenSCAD
		stlPath := filepath.Join(modelDirectory, "output.stl")
		cmd = exec.Command("openscad", "-o", stlPath, "-D", "$fn=100", "-D", "$fs=0.1", "-D", "$fa=1", scadPath)
		openscadOutput, err := cmd.CombinedOutput()
		if err != nil {
			return generateObjectMsg{err: fmt.Errorf("OpenSCAD error: %s\nCommand output: %s\nCleaned SCAD:\n%s", err.Error(), string(openscadOutput), cleanOutput)}
		}

		// Check if STL file was created
		if _, err := os.Stat(stlPath); os.IsNotExist(err) {
			return generateObjectMsg{err: fmt.Errorf("STL file was not created by OpenSCAD. Output: %s", string(openscadOutput))}
		}

		// Convert STL to OBJ using our custom converter
		objPath := filepath.Join(modelDirectory, "output.obj")
		cmd = exec.Command("./stl2obj", stlPath, objPath)
		stl2objOutput, err := cmd.CombinedOutput()
		if err != nil {
			return generateObjectMsg{err: fmt.Errorf("failed to convert STL to OBJ: %w\nOutput: %s", err, string(stl2objOutput))}
		}

		// Check if OBJ file was created
		if _, err := os.Stat(objPath); os.IsNotExist(err) {
			return generateObjectMsg{err: fmt.Errorf("OBJ file was not created. STL2OBJ output: %s", string(stl2objOutput))}
		}

		return generateObjectMsg{stlPath: objPath}
	}
}

func findTerminal3DExecutable() (string, error) {
	binaryNames := []string{"t3d"}
	if runtime.GOOS == "windows" {
		binaryNames = append([]string{"t3d.exe"}, binaryNames...)
	}

	seen := make(map[string]struct{})
	var candidates []string

	push := func(path string) {
		if path == "" {
			return
		}
		if _, ok := seen[path]; ok {
			return
		}
		seen[path] = struct{}{}
		candidates = append(candidates, path)
	}

	if override := os.Getenv(terminal3dEnvOverride); override != "" {
		push(override)
	}

	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		for _, name := range binaryNames {
			push(filepath.Join(exeDir, name))
			push(filepath.Join(exeDir, "bin", name))
			push(filepath.Join(exeDir, "deps", "terminal3d", "target", "release", name))
		}
	}

	for _, name := range binaryNames {
		push("./" + name)
		push(filepath.Join("deps", "terminal3d", "target", "release", name))
	}

	for _, candidate := range candidates {
		if hasPathSeparator(candidate) {
			resolved := candidate
			if !filepath.IsAbs(resolved) {
				if abs, err := filepath.Abs(resolved); err == nil {
					resolved = abs
				}
			}
			if isExecutable(resolved) {
				return resolved, nil
			}
			continue
		}

		if path, err := exec.LookPath(candidate); err == nil {
			return path, nil
		}
	}

	if path, err := exec.LookPath("t3d"); err == nil {
		return path, nil
	}

	return "", fmt.Errorf("terminal3d binary not found. Set %s to the binary path or ensure 't3d' is on PATH", terminal3dEnvOverride)
}

func hasPathSeparator(path string) bool {
	return strings.Contains(path, string(os.PathSeparator)) || strings.Contains(path, "/") || strings.Contains(path, "\\")
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if info.IsDir() {
		return false
	}
	if runtime.GOOS == "windows" {
		return true
	}
	return info.Mode().Perm()&0111 != 0
}

func ensureModelAssets() error {
	modelPath := filepath.Join(modelDirectory, modelFilename)

	if _, err := os.Stat(modelPath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("unable to access model file %s: %w", modelPath, err)
	}

	fmt.Printf("\n📥 Downloading model %s...\nThis may take a few minutes on first run.\n", modelFilename)
	written, err := downloadToFile(modelDownloadURL, modelPath, 0644)
	if err != nil {
		return fmt.Errorf("failed to download model: %w", err)
	}

	fmt.Printf("✅ Model downloaded (%.2f MB).\n", float64(written)/1024/1024)
	return nil
}

func ensureModelRunner() (string, error) {
	for _, envVar := range runnerOverrideEnvVars {
		if override := os.Getenv(envVar); override != "" {
			if isExecutable(override) {
				if abs, err := filepath.Abs(override); err == nil {
					return abs, nil
				}
				return override, nil
			}
			return "", fmt.Errorf("provided runner override %s=%q is not executable", envVar, override)
		}
	}

	defaultRun := filepath.Join(modelDirectory, "run")
	if isExecutable(defaultRun) {
		if abs, err := filepath.Abs(defaultRun); err == nil {
			return abs, nil
		}
		return defaultRun, nil
	}

	key := runtime.GOOS + "/" + runtime.GOARCH
	pkg, ok := runnerPackages[key]
	if !ok {
		return "", fmt.Errorf("no bundled llama.cpp runtime available for %s. Place an executable at %s or set %s", key, defaultRun, runnerOverrideEnvVars[0])
	}

	extractDir := filepath.Join(modelDirectory, pkg.extractSubdir)
	if runner, err := findRunnerBinary(extractDir); err == nil {
		if abs, absErr := filepath.Abs(runner); absErr == nil {
			return abs, nil
		}
		return runner, nil
	}

	fmt.Printf("\n📦 Preparing llama.cpp runtime (%s)...\n", pkg.label)
	archivePath := filepath.Join(modelDirectory, pkg.archiveName)
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		if _, err := downloadToFile(pkg.url, archivePath, 0644); err != nil {
			return "", fmt.Errorf("failed to download runtime archive: %w", err)
		}
	}

	if err := os.RemoveAll(extractDir); err != nil {
		return "", fmt.Errorf("failed to reset runtime directory: %w", err)
	}

	if err := unzipArchive(archivePath, extractDir); err != nil {
		return "", fmt.Errorf("failed to extract runtime archive: %w", err)
	}

	runner, err := findRunnerBinary(extractDir)
	if err != nil {
		return "", err
	}

	fmt.Printf("✅ llama.cpp runtime ready.\n")
	if abs, absErr := filepath.Abs(runner); absErr == nil {
		return abs, nil
	}
	return runner, nil
}

func unzipArchive(archivePath, destDir string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	for _, file := range reader.File {
		cleanPath, err := sanitizeExtractPath(destDir, file.Name)
		if err != nil {
			return err
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(cleanPath, 0755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(cleanPath), 0755); err != nil {
			return err
		}

		src, err := file.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(cleanPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			_ = src.Close()
			return err
		}
		if _, err := io.Copy(out, src); err != nil {
			_ = src.Close()
			_ = out.Close()
			return err
		}
		_ = src.Close()
		_ = out.Close()
	}

	return nil
}

var errRunnerLocated = errors.New("runner located")

func findRunnerBinary(root string) (string, error) {
	info, err := os.Stat(root)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("runtime path %s is not a directory", root)
	}

	var runner string
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() != "run" {
			return nil
		}
		if err := os.Chmod(path, 0755); err != nil && !errors.Is(err, os.ErrPermission) {
			return err
		}
		if isExecutable(path) {
			runner = path
			return errRunnerLocated
		}
		return nil
	})
	if err != nil && !errors.Is(err, errRunnerLocated) {
		return "", err
	}
	if runner == "" {
		return "", fmt.Errorf("no executable 'run' found under %s", root)
	}
	return runner, nil
}

func sanitizeExtractPath(destination, filePath string) (string, error) {
	destination = filepath.Clean(destination)
	targetPath := filepath.Join(destination, filePath)
	if !strings.HasPrefix(filepath.Clean(targetPath), destination+string(os.PathSeparator)) && filepath.Clean(targetPath) != destination {
		return "", fmt.Errorf("illegal file path %s", filePath)
	}
	return filepath.Clean(targetPath), nil
}

func downloadToFile(url, destination string, perm os.FileMode) (int64, error) {
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status %s", resp.Status)
	}

	if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
		return 0, err
	}

	tmpPath := destination + ".download"
	out, err := os.Create(tmpPath)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = out.Close()
	}()

	written, err := io.Copy(out, resp.Body)
	if err != nil {
		_ = os.Remove(tmpPath)
		return 0, err
	}

	if err := out.Sync(); err != nil {
		_ = os.Remove(tmpPath)
		return 0, err
	}

	if err := out.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return 0, err
	}

	if err := os.Rename(tmpPath, destination); err != nil {
		_ = os.Remove(tmpPath)
		return 0, err
	}

	if perm != 0 {
		if err := os.Chmod(destination, perm); err != nil {
			return 0, err
		}
	}

	return written, nil
}

func cleanModelOutput(output, objectName string) string {
	lines := strings.Split(output, "\n")
	var scadLines []string
	foundModel := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for the "model" marker
		if line == "model" {
			foundModel = true
			continue
		}

		// After finding "model", collect SCAD code until we hit EOF or debug info
		if foundModel {
			// Stop at EOF marker or debug output
			if strings.Contains(line, "> EOF by user") ||
				strings.Contains(line, "llama_perf_") ||
				strings.Contains(line, "main:") ||
				strings.Contains(line, "sampler") ||
				strings.Contains(line, "system_info:") ||
				strings.Contains(line, "ggml_") {
				break
			}

			// Skip empty lines at the beginning
			if len(scadLines) == 0 && line == "" {
				continue
			}

			// Only add lines that look like SCAD code or are empty continuation lines
			if line != "" {
				scadLines = append(scadLines, line)
			}
		}
	}

	return strings.Join(scadLines, "\n")
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
