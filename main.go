package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type state int

const (
	stateInput state = iota
	stateGenerating
	stateConverting
	stateSTLView
)

type model struct {
	state        state
	textInput    textinput.Model
	prompt       string
	object3D     string
	selectedObject string
	width        int
	height       int
	stlModel     *STLModel
	rotX, rotY   float64
	rotZ         float64
	renderStyle  string
	autoRotate   bool
	rotationSpeed float64
	generatingMsg string
	err          error
}

type tickMsg time.Time

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
		state:         stateInput,
		textInput:     ti,
		renderStyle:   "solid",
		autoRotate:    true,
		rotationSpeed: 0.03,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*50, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		if m.state == stateSTLView && m.autoRotate {
			m.rotY += m.rotationSpeed
			return m, tickCmd()
		}
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.state == stateSTLView && m.autoRotate {
			return m, tickCmd()
		}
		return m, nil

	case generateObjectMsg:
		if msg.err != nil {
			m.err = msg.err
			m.state = stateInput
			m.textInput.Focus()
			return m, textinput.Blink
		}
		
		// Load the generated STL
		stlModel, err := ParseSTL(msg.stlPath)
		if err != nil {
			m.err = err
			m.state = stateInput
			m.textInput.Focus()
			return m, textinput.Blink
		}
		
		m.stlModel = stlModel
		m.stlModel.Name = m.selectedObject
		m.state = stateSTLView
		m.rotX, m.rotY, m.rotZ = 0, 0, 0
		m.autoRotate = true
		return m, tickCmd()
	
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

		case stateSTLView:
			switch msg.String() {
			case " ":
				m.autoRotate = !m.autoRotate
				if m.autoRotate {
					return m, tickCmd()
				}
				return m, nil
			case "left", "a":
				m.autoRotate = false
				m.rotY += 0.1
			case "right", "d":
				m.autoRotate = false
				m.rotY -= 0.1
			case "up", "w":
				m.autoRotate = false
				m.rotX += 0.1
			case "down", "s":
				m.autoRotate = false
				m.rotX -= 0.1
			case "q", "Q":
				m.autoRotate = false
				m.rotZ += 0.1
			case "e", "E":
				m.autoRotate = false
				m.rotZ -= 0.1
			case "r", "R":
				m.rotX, m.rotY, m.rotZ = 0, 0, 0
				m.autoRotate = true
				return m, tickCmd()
			case "v", "V":
				if m.renderStyle == "solid" {
					m.renderStyle = "wireframe"
				} else {
					m.renderStyle = "solid"
				}
			case "m", "M":
				m.autoRotate = false
				m.state = stateInput
				m.textInput.Focus()
				return m, textinput.Blink
			case "t", "T", "enter":
				m.rotX, m.rotY, m.rotZ = 0, 0, 0
				m.autoRotate = true
				return m, tickCmd()
			case "ctrl+c", "esc":
				return m, tea.Quit
			}
			return m, nil
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

	case stateSTLView:
		// Calculate render dimensions based on terminal size
		// Account for header, status, controls and margins
		headerLines := 6  // title, status, spacing
		controlLines := 4 // control instructions
		marginLines := 2  // top/bottom margins
		
		renderHeight := m.height - headerLines - controlLines - marginLines
		if renderHeight < 20 {
			renderHeight = 20
		}
		if renderHeight > 150 {
			renderHeight = 150
		}
		
		// Width calculation - use most of terminal width
		renderWidth := m.width - 4
		if renderWidth < 40 {
			renderWidth = 40
		}
		if renderWidth > 300 {
			renderWidth = 300
		}
		
		rotationStatus := "Auto-rotating"
		if !m.autoRotate {
			rotationStatus = "Manual control"
		}
		
		s.WriteString(promptStyle.Render(fmt.Sprintf("3D Model: %s", m.stlModel.Name)) + "\n")
		s.WriteString(helpStyle.Render(fmt.Sprintf("Mode: %s | %s | X:%.1f Y:%.1f Z:%.1f", 
			m.renderStyle, rotationStatus, m.rotX, m.rotY, m.rotZ)) + "\n")
		
		renderer := NewRenderer(renderWidth-4, renderHeight-2)
		stlRender := renderer.RenderModel(m.stlModel, m.rotX, m.rotY, m.rotZ, m.renderStyle)
		
		// Remove extra padding from objectStyle and center properly
		modelDisplay := lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("#50FA7B")).
			Align(lipgloss.Center).
			Width(renderWidth).
			Height(renderHeight).
			Render(stlRender)
		
		s.WriteString("\n" + modelDisplay + "\n")
		
		s.WriteString(helpStyle.Render("Controls:") + "\n")
		s.WriteString(helpStyle.Render("SPACE: Pause/Resume rotation | Arrow keys: Manual rotate") + "\n")
		s.WriteString(helpStyle.Render("R: Reset & auto-rotate | V: Toggle solid/wireframe") + "\n")
		s.WriteString(helpStyle.Render("M: Make a new | T: Reset rotation | Esc: Quit"))
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
			return generateObjectMsg{err: err}
		}

		// Run the AI model to generate SCAD code
		prompt := fmt.Sprintf("Hey cadmonkey, make me %s", objectName)
		cmd := exec.Command("./run", "-m", "k-1b-q8_0.gguf", "-p", prompt)
		cmd.Dir = "k"
		
		output, err := cmd.Output()
		if err != nil {
			return generateObjectMsg{err: fmt.Errorf("failed to run model: %w", err)}
		}

		// Clean model output - remove prompt echo and 'model' line
		cleanOutput := cleanModelOutput(string(output), objectName)
		
		// Save cleaned model output as SCAD file
		scadPath := filepath.Join("k", "output.scad")
		if err := os.WriteFile(scadPath, []byte(cleanOutput), 0644); err != nil {
			return generateObjectMsg{err: fmt.Errorf("failed to write SCAD file: %w", err)}
		}

		// Convert SCAD to STL using OpenSCAD
		stlPath := filepath.Join("k", "output.stl")
		cmd = exec.Command("openscad", "-o", stlPath, scadPath)
		openscadOutput, err := cmd.CombinedOutput()
		if err != nil {
			return generateObjectMsg{err: fmt.Errorf("OpenSCAD error: %s\nCommand output: %s\nCleaned SCAD:\n%s\nRaw model output:\n%s", err.Error(), string(openscadOutput), cleanOutput, string(output))}
		}

		return generateObjectMsg{stlPath: stlPath}
	}
}

func cleanModelOutput(output, objectName string) string {
	lines := strings.Split(output, "\n")
	var cleanLines []string
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip the "user" line
		if line == "user" {
			continue
		}
		// Skip the prompt echo line
		if strings.Contains(line, "Hey cadmonkey, make me") {
			continue
		}
		// Skip the "model" line
		if line == "model" {
			continue
		}
		// Skip EOF marker
		if strings.Contains(line, "> EOF by user") {
			continue
		}
		// Skip empty lines at the beginning
		if len(cleanLines) == 0 && line == "" {
			continue
		}
		cleanLines = append(cleanLines, line)
	}
	
	return strings.Join(cleanLines, "\n")
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}