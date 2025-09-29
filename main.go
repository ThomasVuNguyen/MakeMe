package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type state int

const (
	stateInput state = iota
	stateDisplay
	stateSTLView
)

type model struct {
	state        state
	textInput    textinput.Model
	prompt       string
	object3D     string
	width        int
	height       int
	stlModel     *STLModel
	rotX, rotY   float64
	rotZ         float64
	renderStyle  string
	autoRotate   bool
	rotationSpeed float64
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
	ti.Placeholder = "a duck, a car, a flower..."
	ti.Focus()
	ti.CharLimit = 50
	ti.Width = 46

	stlModel, _ := ParseSTL("pikachu.stl")

	return model{
		state:         stateInput,
		textInput:     ti,
		stlModel:      stlModel,
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

	case tea.KeyMsg:
		switch m.state {
		case stateInput:
			switch msg.Type {
			case tea.KeyEnter:
				words := strings.Fields(m.textInput.Value())
				if strings.ToLower(m.textInput.Value()) == "pikachu" && m.stlModel != nil {
					m.prompt = m.textInput.Value()
					m.state = stateSTLView
					m.rotX, m.rotY, m.rotZ = 0, 0, 0
					m.autoRotate = true
					m.textInput.Reset()
					return m, tickCmd()
				} else if len(words) >= 2 && len(words) <= 5 {
					m.prompt = m.textInput.Value()
					m.object3D = generate3D(m.prompt)
					m.state = stateDisplay
					m.textInput.Reset()
					return m, nil
				}
			case tea.KeyCtrlC, tea.KeyEsc:
				return m, tea.Quit
			}

		case stateDisplay:
			switch msg.String() {
			case "m", "M":
				m.state = stateInput
				m.textInput.Focus()
				return m, textinput.Blink
			case "t", "T", "enter":
				words := strings.Fields(m.prompt)
				if len(words) >= 2 && len(words) <= 5 {
					m.object3D = generate3D(m.prompt)
					return m, nil
				}
			case "q", "ctrl+c", "esc":
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

	switch m.state {
	case stateInput:
		s.WriteString(promptStyle.Render("Describe an object (2-5 words) or type 'pikachu':") + "\n")
		s.WriteString(inputStyle.Render(m.textInput.View()) + "\n\n")
		
		wordCount := len(strings.Fields(m.textInput.Value()))
		status := fmt.Sprintf("Words: %d/5", wordCount)
		if wordCount < 2 && m.textInput.Value() != "" && strings.ToLower(m.textInput.Value()) != "pikachu" {
			status += " (minimum 2 words)"
		}
		s.WriteString(helpStyle.Render(status) + "\n")
		s.WriteString(helpStyle.Render("Press Enter to generate • Esc to quit"))

	case stateDisplay:
		s.WriteString(promptStyle.Render(fmt.Sprintf("Generated: %s", m.prompt)) + "\n")
		s.WriteString(objectStyle.Width(60).Height(15).Render(m.object3D) + "\n")
		
		makeNewBtn := buttonStyle.Render("[M] Make a new")
		tryBtn := buttonStyle.Render("[T] Try again")
		
		buttons := lipgloss.JoinHorizontal(lipgloss.Top, makeNewBtn, tryBtn)
		s.WriteString(lipgloss.PlaceHorizontal(60, lipgloss.Center, buttons) + "\n\n")
		s.WriteString(helpStyle.Render("Press M to make a new • T to try again • Esc to quit"))

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

func generate3D(prompt string) string {
	objects := map[string]string{
		"duck": `
       __
     <(o )___
      ( ._> /
       '---'`,
		"car": `
       ______
      /|_||_\'.__ 
     (   _    _ _\
     ='-(_)--(_)-'`,
		"flower": `
       .--.---.
      /  (   ) \
      \  '---' /
       '--\_/--'
          |
        __|__`,
		"house": `
         /\
        /  \
       /    \
      |  __  |
      | |  | |
      |_|__|_|`,
		"tree": `
        🌳
       /|\
      / | \
     /  |  \
        |
       _|_`,
		"cat": `
      /\_/\
     ( o.o )
      > ^ <
     /|   |\
    (_|   |_)`,
		"star": `
        ✨
       / \
      /   \
     |  *  |
      \   /
       \ /`,
	}

	prompt = strings.ToLower(prompt)
	
	for key, art := range objects {
		if strings.Contains(prompt, key) {
			return art
		}
	}

	ascii3D := fmt.Sprintf(`
    ╔════════════╗
    ║            ║
    ║   %s%-8s%s   ║
    ║            ║
    ║    🎲 3D    ║
    ║            ║
    ╚════════════╝
      Processing...`, "\033[1m", strings.ToUpper(prompt[:min(8, len(prompt))]), "\033[0m")

	return ascii3D
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}