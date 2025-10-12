package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	helpTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	helpSectionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("62")).
				Bold(true)

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
)

// HelpModel manages the help view
type HelpModel struct {
	app    *App
	width  int
	height int
}

// NewHelpModel creates a new help model
func NewHelpModel(app *App) *HelpModel {
	return &HelpModel{
		app: app,
	}
}

// Init initializes the help view
func (m *HelpModel) Init() tea.Cmd {
	return nil
}

// Update handles updates for the help view
func (m *HelpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q", "?":
			return m.app, m.app.SwitchToView(ViewNotesList)
		}
	}
	return m.app, nil
}

// View renders the help view
func (m *HelpModel) View() string {
	s := helpTitleStyle.Render("Help & Keyboard Shortcuts") + "\n\n"

	// Notes List shortcuts
	s += helpSectionStyle.Render("Notes List") + "\n"
	s += formatHelpItem("n", "Create new note")
	s += formatHelpItem("e, Enter", "Edit selected note")
	s += formatHelpItem("d", "Delete selected note")
	s += formatHelpItem("Ctrl+S", "Toggle search mode")
	s += formatHelpItem("↑, k", "Move cursor up")
	s += formatHelpItem("↓, j", "Move cursor down")
	s += formatHelpItem("?", "Show this help")
	s += "\n"

	// Search shortcuts
	s += helpSectionStyle.Render("Search Mode") + "\n"
	s += formatHelpItem("Ctrl+S", "Enter/exit search mode")
	s += formatHelpItem("Type", "Search notes (fuzzy matching)")
	s += formatHelpItem("Enter", "Confirm search")
	s += formatHelpItem("Esc", "Cancel search")
	s += formatHelpItem("Backspace", "Delete search character")
	s += "\n"

	// Editor shortcuts
	s += helpSectionStyle.Render("Note Editor") + "\n"
	s += formatHelpItem("Tab", "Switch between title and content")
	s += formatHelpItem("Ctrl+S", "Save note")
	s += formatHelpItem("Esc", "Cancel and return to notes list")
	s += formatHelpItem("Enter", "New line (in content)")
	s += formatHelpItem("Backspace", "Delete character")
	s += "\n"

	// General shortcuts
	s += helpSectionStyle.Render("General") + "\n"
	s += formatHelpItem("Esc", "Return to notes list (from any view)")
	s += formatHelpItem("q, Ctrl+C", "Quit application")
	s += "\n"

	// Footer
	s += "Press Esc, q, or ? to close help"

	return s
}

// formatHelpItem formats a key-description pair
func formatHelpItem(key, description string) string {
	keyPart := helpKeyStyle.Render(strings.Repeat(" ", 12-len(key)) + key)
	descPart := helpDescStyle.Render(description)
	return keyPart + " " + descPart + "\n"
}
