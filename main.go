package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type state int

const (
	stateInput state = iota
	stateDisplay
)

type model struct {
	state      state
	textInput  textinput.Model
	prompt     string
	object3D   string
	width      int
	height     int
	err        error
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
	ti.Placeholder = "a duck, a car, a flower..."
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

	case tea.KeyMsg:
		switch m.state {
		case stateInput:
			switch msg.Type {
			case tea.KeyEnter:
				words := strings.Fields(m.textInput.Value())
				if len(words) >= 2 && len(words) <= 5 {
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
			case "i", "I":
				fmt.Printf("\n✅ Created 3D object: %s\n", m.prompt)
				return m, tea.Quit
			case "t", "T", "enter":
				m.state = stateInput
				m.textInput.Focus()
				return m, textinput.Blink
			case "q", "ctrl+c", "esc":
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

	switch m.state {
	case stateInput:
		s.WriteString(promptStyle.Render("Describe an object (2-5 words):") + "\n")
		s.WriteString(inputStyle.Render(m.textInput.View()) + "\n\n")
		
		wordCount := len(strings.Fields(m.textInput.Value()))
		status := fmt.Sprintf("Words: %d/5", wordCount)
		if wordCount < 2 && m.textInput.Value() != "" {
			status += " (minimum 2 words)"
		}
		s.WriteString(helpStyle.Render(status) + "\n")
		s.WriteString(helpStyle.Render("Press Enter to generate • Esc to quit"))

	case stateDisplay:
		s.WriteString(promptStyle.Render(fmt.Sprintf("Generated: %s", m.prompt)) + "\n")
		s.WriteString(objectStyle.Width(60).Height(15).Render(m.object3D) + "\n")
		
		likeBtn := buttonStyle.Render("[I] I like it")
		tryBtn := buttonStyle.Render("[T] Try again")
		
		buttons := lipgloss.JoinHorizontal(lipgloss.Top, likeBtn, tryBtn)
		s.WriteString(lipgloss.PlaceHorizontal(60, lipgloss.Center, buttons) + "\n\n")
		s.WriteString(helpStyle.Render("Press I to save • T to try again • Esc to quit"))
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