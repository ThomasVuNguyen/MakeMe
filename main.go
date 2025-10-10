package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	state        state
	textInput    textinput.Model
	selectedObject string
	width        int
	height       int
	generatingMsg string
	err          error
}

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
		
		// Display OBJ file using terminal3d
		// Exit the current app and launch terminal3d directly
		fmt.Printf("\n🎯 Opening 3D model: %s\n", msg.stlPath)
		fmt.Printf("Press Ctrl+C to return when done viewing.\n\n")
		
		cmd := exec.Command("t3d", msg.stlPath)
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
		s.WriteString(errorStyle.Render("Error: " + m.err.Error()) + "\n\n")
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
		if err := os.MkdirAll("k", 0755); err != nil {
			return generateObjectMsg{err: fmt.Errorf("failed to create k directory: %w", err)}
		}

		// Run the AI model to generate SCAD code
		prompt := fmt.Sprintf("Hey cadmonkey, make me %s", objectName)
		cmd := exec.Command("./run", "-m", "k-1b-q8_0.gguf", "-p", prompt)
		cmd.Dir = "k"
		
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
		scadPath := filepath.Join("k", "output.scad")
		if err := os.WriteFile(scadPath, []byte(cleanOutput), 0644); err != nil {
			return generateObjectMsg{err: fmt.Errorf("failed to write SCAD file: %w", err)}
		}

		// Convert SCAD to STL using OpenSCAD
		stlPath := filepath.Join("k", "output.stl")
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
		objPath := filepath.Join("k", "output.obj")
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