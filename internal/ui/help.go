package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Note: Styles are now defined inline with enhanced colors and responsive design

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
	// Enhanced responsive title style
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F1F5F9")).
		Background(lipgloss.Color("#38BDF8")).
		Bold(true).
		Padding(0, 1).
		MarginBottom(1)

	s := titleStyle.Render("Help & Keyboard Shortcuts") + "\n\n"

	// Enhanced section styles
	sectionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#10B981")).
		Bold(true).
		MarginTop(1).
		MarginBottom(1)

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#38BDF8")).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#94A3B8"))

	// Responsive layout based on terminal width
	useCompactLayout := m.width < 120

	// Notes List shortcuts
	s += sectionStyle.Render("📝 Notes List") + "\n"
	if useCompactLayout {
		s += formatHelpItemCompact("n", "New note", keyStyle, descStyle)
		s += formatHelpItemCompact("e, Enter", "Edit note", keyStyle, descStyle)
		s += formatHelpItemCompact("d", "Delete note", keyStyle, descStyle)
		s += formatHelpItemCompact("Ctrl+S", "Search mode", keyStyle, descStyle)
		s += formatHelpItemCompact("↑, k", "Move up", keyStyle, descStyle)
		s += formatHelpItemCompact("↓, j", "Move down", keyStyle, descStyle)
		s += formatHelpItemCompact("?", "Help", keyStyle, descStyle)
	} else {
		s += formatHelpItem("n", "Create new note", keyStyle, descStyle)
		s += formatHelpItem("e, Enter", "Edit selected note", keyStyle, descStyle)
		s += formatHelpItem("d", "Delete selected note", keyStyle, descStyle)
		s += formatHelpItem("Ctrl+S", "Toggle search mode", keyStyle, descStyle)
		s += formatHelpItem("↑, k", "Move cursor up", keyStyle, descStyle)
		s += formatHelpItem("↓, j", "Move cursor down", keyStyle, descStyle)
		s += formatHelpItem("?", "Show this help", keyStyle, descStyle)
	}
	s += "\n"

	// Search shortcuts
	s += sectionStyle.Render("🔍 Search Mode") + "\n"
	if useCompactLayout {
		s += formatHelpItemCompact("Ctrl+S", "Enter/exit search", keyStyle, descStyle)
		s += formatHelpItemCompact("Type", "Fuzzy search", keyStyle, descStyle)
		s += formatHelpItemCompact("Enter", "Confirm search", keyStyle, descStyle)
		s += formatHelpItemCompact("Esc", "Cancel search", keyStyle, descStyle)
		s += formatHelpItemCompact("Backspace", "Delete char", keyStyle, descStyle)
	} else {
		s += formatHelpItem("Ctrl+S", "Enter/exit search mode", keyStyle, descStyle)
		s += formatHelpItem("Type", "Search notes (fuzzy matching)", keyStyle, descStyle)
		s += formatHelpItem("Enter", "Confirm search", keyStyle, descStyle)
		s += formatHelpItem("Esc", "Cancel search", keyStyle, descStyle)
		s += formatHelpItem("Backspace", "Delete search character", keyStyle, descStyle)
	}
	s += "\n"

	// Editor shortcuts
	s += sectionStyle.Render("✏️ Note Editor") + "\n"
	if useCompactLayout {
		s += formatHelpItemCompact("Tab", "Switch fields", keyStyle, descStyle)
		s += formatHelpItemCompact("Ctrl+S", "Save note", keyStyle, descStyle)
		s += formatHelpItemCompact("Ctrl+P", "Toggle preview", keyStyle, descStyle)
		s += formatHelpItemCompact("Esc", "Cancel", keyStyle, descStyle)
		s += formatHelpItemCompact("Enter", "New line / Confirm", keyStyle, descStyle)
		s += formatHelpItemCompact("Space", "Separate tags", keyStyle, descStyle)
	} else {
		s += formatHelpItem("Tab", "Switch between title/content/tags", keyStyle, descStyle)
		s += formatHelpItem("Ctrl+S", "Save note", keyStyle, descStyle)
		s += formatHelpItem("Ctrl+P", "Toggle preview", keyStyle, descStyle)
		s += formatHelpItem("Esc", "Cancel and return to notes list", keyStyle, descStyle)
		s += formatHelpItem("Enter", "New line (in content) / Confirm tag", keyStyle, descStyle)
		s += formatHelpItem("Space", "Separate tags", keyStyle, descStyle)
	}
	s += "\n"

	// Tag management shortcuts
	s += sectionStyle.Render("🏷️ Tag Management") + "\n"
	if useCompactLayout {
		s += formatHelpItemCompact("Tab to Tags", "Switch to tags", keyStyle, descStyle)
		s += formatHelpItemCompact("Type", "Add tags", keyStyle, descStyle)
		s += formatHelpItemCompact("Space/Enter", "Confirm tag", keyStyle, descStyle)
		s += formatHelpItemCompact("↑/↓", "Navigate suggestions", keyStyle, descStyle)
		s += formatHelpItemCompact("Esc", "Close suggestions", keyStyle, descStyle)
	} else {
		s += formatHelpItem("Tab to Tags", "Switch to tag input field", keyStyle, descStyle)
		s += formatHelpItem("Type", "Add new tags (auto-suggests existing)", keyStyle, descStyle)
		s += formatHelpItem("Space/Enter", "Confirm tag addition", keyStyle, descStyle)
		s += formatHelpItem("↑/↓", "Navigate tag suggestions", keyStyle, descStyle)
		s += formatHelpItem("Esc", "Close tag suggestions", keyStyle, descStyle)
	}
	s += "\n"

	// General shortcuts
	s += sectionStyle.Render("⚙️ General") + "\n"
	if useCompactLayout {
		s += formatHelpItemCompact("Esc", "Return to notes list", keyStyle, descStyle)
		s += formatHelpItemCompact("q, Ctrl+C", "Quit application", keyStyle, descStyle)
	} else {
		s += formatHelpItem("Esc", "Return to notes list (from any view)", keyStyle, descStyle)
		s += formatHelpItem("q, Ctrl+C", "Quit application", keyStyle, descStyle)
	}
	s += "\n"

	// Enhanced footer
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#64748B")).
		Italic(true).
		MarginTop(1)
	s += footerStyle.Render("Press Esc, q, or ? to close help")

	return s
}

// formatHelpItem formats a key-description pair with responsive layout
func formatHelpItem(key, description string, keyStyle, descStyle lipgloss.Style) string {
	keyPart := keyStyle.Render(strings.Repeat(" ", 12-len(key)) + key)
	descPart := descStyle.Render(description)
	return keyPart + " " + descPart + "\n"
}

// formatHelpItemCompact formats a key-description pair for small terminals
func formatHelpItemCompact(key, description string, keyStyle, descStyle lipgloss.Style) string {
	keyPart := keyStyle.Render(strings.Repeat(" ", 8-len(key)) + key)
	descPart := descStyle.Render(description)
	return keyPart + " " + descPart + "\n"
}
