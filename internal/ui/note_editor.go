package ui

import (
	"strings"

	"markdown-note-taking-app/internal/models"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	editorTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFDF5")).
				Background(lipgloss.Color("#25A065")).
				Padding(0, 1)

	activeFieldStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("39")).
				Padding(0, 1)

	inactiveFieldStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("241")).
				Padding(0, 1)

	contentStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1)
)

// NoteEditorModel manages the note editor view
type NoteEditorModel struct {
	app     *App
	note    *models.Note
	title   string
	content string
	focused bool // true for content, false for title
	cursor  int
	mode    string // "create" or "edit"
	width   int
	height  int
}

// NewNoteEditorModel creates a new note editor model
func NewNoteEditorModel(app *App) *NoteEditorModel {
	return &NoteEditorModel{
		app:     app,
		note:    nil,
		title:   "",
		content: "",
		focused: false,
		cursor:  0,
		mode:    "create",
	}
}

// Init initializes the note editor
func (m *NoteEditorModel) Init(selectedNote *models.Note) tea.Cmd {
	if selectedNote != nil {
		m.SetNote(selectedNote)
		return nil
	}

	// Reset editor for new note
	m.title = ""
	m.content = ""
	m.focused = false
	m.mode = "create"
	return nil
}

// SetNote sets the editor to edit mode with an existing note
func (m *NoteEditorModel) SetNote(note *models.Note) {
	m.note = note
	m.title = note.Title
	m.content = note.Content
	m.focused = false
	m.mode = "edit"
}

// Update handles updates for the note editor
func (m *NoteEditorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+s":
			// Save note
			return m.app, m.saveNote()
		case "tab":
			// Switch focus between title and content
			m.focused = !m.focused
		case "esc":
			// Cancel and go back
			return m.app, m.app.SwitchToView(ViewNotesList)
		case "backspace":
			if m.focused {
				// Editing content
				if len(m.content) > 0 {
					m.content = m.content[:len(m.content)-1]
				}
			} else {
				// Editing title
				if len(m.title) > 0 {
					m.title = m.title[:len(m.title)-1]
				}
			}
		case "enter":
			if m.focused {
				// New line in content
				m.content += "\n"
			}
		default:
			// Regular character input
			char := msg.String()
			if len(char) == 1 {
				if m.focused {
					m.content += char
				} else {
					m.title += char
				}
			}
		}
	}
	return m.app, nil
}

// saveNote saves the current note
func (m *NoteEditorModel) saveNote() tea.Cmd {
	return func() tea.Msg {
		if strings.TrimSpace(m.title) == "" {
			// Don't save notes without titles
			return nil
		}

		if m.mode == "create" {
			_, err := m.app.GetStorage().CreateNote(m.title, m.content)
			if err != nil {
				// For now, just ignore errors
				return nil
			}
		} else {
			// Update existing note
			if m.note != nil {
				m.note.Title = m.title
				m.note.Content = m.content
				err := m.app.GetStorage().UpdateNote(m.note)
				if err != nil {
					// For now, just ignore errors
					return nil
				}
			}
		}

		// Go back to notes list
		return m.app.SwitchToView(ViewNotesList)()
	}
}

// View renders the note editor
func (m *NoteEditorModel) View() string {
	mode := "Create Note"
	if m.mode == "edit" {
		mode = "Edit Note"
	}

	s := editorTitleStyle.Render(mode) + "\n\n"

	// Title field
	titleLabel := "Title:"
	if !m.focused {
		titleLabel = "[*] " + titleLabel
	} else {
		titleLabel = "[ ] " + titleLabel
	}
	s += titleLabel + "\n"

	// Render title field with appropriate style
	var titleField string
	if !m.focused {
		// Title is active
		if m.title == "" {
			titleField = activeFieldStyle.Width(m.width - 6).Render("Enter title...")
		} else {
			titleField = activeFieldStyle.Width(m.width - 6).Render(m.title)
		}
	} else {
		// Title is inactive
		titleField = inactiveFieldStyle.Width(m.width - 6).Render(m.title)
	}
	s += titleField + "\n\n"

	// Content field
	contentLabel := "Content:"
	if m.focused {
		contentLabel = "[*] " + contentLabel
	} else {
		contentLabel = "[ ] " + contentLabel
	}
	s += contentLabel + "\n"

	// Calculate available height for content
	usedHeight := 15                           // title + spacing
	contentHeight := m.height - usedHeight - 4 // reserve space for controls
	if contentHeight < 5 {
		contentHeight = 5 // minimum height
	}

	// Render content area
	var displayContent string
	if m.content == "" && m.focused {
		displayContent = "Start writing your note..."
	} else if m.content == "" {
		displayContent = ""
	} else {
		displayContent = m.content
	}

	// Apply different border style based on focus
	var contentBoxStyle lipgloss.Style
	if m.focused {
		contentBoxStyle = contentStyle.Copy().BorderForeground(lipgloss.Color("39")) // Active color
	} else {
		contentBoxStyle = contentStyle.Copy().BorderForeground(lipgloss.Color("62")) // Inactive color
	}

	contentBox := contentBoxStyle.Width(m.width - 4).Height(contentHeight).Render(displayContent)
	s += contentBox

	// Controls
	s += "\n\nControls:\n"
	s += "  Tab - Switch between title/content\n"
	s += "  Ctrl+S - Save\n"
	s += "  Esc - Cancel\n"

	return s
}
