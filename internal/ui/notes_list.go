package ui

import (
	"fmt"

	"markdown-note-taking-app/internal/models"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFDF5")).
		Background(lipgloss.Color("#25A065")).
		Padding(0, 1)
)

// NotesListModel manages the notes list view
type NotesListModel struct {
	app          *App
	notes        []*models.Note
	selectedNote *models.Note
	cursor       int
	loaded       bool
	width        int
	height       int
}

// NewNotesListModel creates a new notes list model
func NewNotesListModel(app *App) *NotesListModel {
	return &NotesListModel{
		app:          app,
		notes:        []*models.Note{},
		selectedNote: nil,
		cursor:       0,
		loaded:       false,
	}
}

// Init initializes the notes list
func (m *NotesListModel) Init() tea.Cmd {
	return m.loadNotes()
}

// loadNotes loads notes from storage
func (m *NotesListModel) loadNotes() tea.Cmd {
	return func() tea.Msg {
		notes, err := m.app.GetStorage().GetAllNotes(models.NoteFilter{Limit: 100})
		if err != nil {
			// For now, just return empty list on error
			return notesLoadedMsg{notes: []*models.Note{}}
		}
		return notesLoadedMsg{notes: notes}
	}
}

// Update handles updates for the notes list
func (m *NotesListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case notesLoadedMsg:
		m.notes = msg.notes
		m.loaded = true
		return m.app, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.notes)-1 {
				m.cursor++
			}
		case "n":
			// New note
			m.selectedNote = nil
			return m.app, m.app.SwitchToView(ViewNoteEditor)
		case "e", "enter":
			// Edit selected note
			if len(m.notes) > 0 {
				m.selectedNote = m.notes[m.cursor]
				return m.app, m.app.SwitchToView(ViewNoteEditor)
			}
		case "d":
			// Delete selected note
			if len(m.notes) > 0 {
				m.selectedNote = nil
				return m.app, m.deleteNote()
			}
		}
	}
	return m.app, nil
}

// deleteNote deletes the currently selected note
func (m *NotesListModel) deleteNote() tea.Cmd {
	if len(m.notes) == 0 {
		return nil
	}

	selectedNote := m.notes[m.cursor]
	return func() tea.Msg {
		err := m.app.GetStorage().DeleteNote(selectedNote.ID)
		if err != nil {
			// For now, just ignore errors
			return nil
		}
		// Reload notes after deletion
		return m.loadNotes()()
	}
}

// View renders the notes list
func (m *NotesListModel) View() string {
	if !m.loaded {
		return "Loading notes..."
	}

	if len(m.notes) == 0 {
		return titleStyle.Render("Markdown Notes") + "\n\nNo notes yet. Press 'n' to create your first note.\n\n" +
			"Controls:\n" +
			"  n - New note\n" +
			"  e - Edit note\n" +
			"  d - Delete note\n" +
			"  ↑/↓ - Navigate\n" +
			"  q - Quit\n" +
			"  ? - Help"
	}

	s := titleStyle.Render("Markdown Notes") + "\n\n"

	for i, note := range m.notes {
		cursor := "  "
		if m.cursor == i {
			cursor = "> "
		}

		// Truncate title if too long
		title := note.Title
		if len(title) > 50 {
			title = title[:47] + "..."
		}

		s += fmt.Sprintf("%s%s\n", cursor, title)
	}

	s += "\nControls: n (new) • e (edit) • d (delete) • ↑↓ (navigate) • q (quit) • ? (help)"
	return s
}

// Messages
type notesLoadedMsg struct {
	notes []*models.Note
}
