package ui

import (
	"fmt"

	"markdown-note-taking-app/internal/models"
	"markdown-note-taking-app/internal/utils"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFDF5")).
		Background(lipgloss.Color("#25A065")).
		Padding(0, 1)

	searchStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(0, 1).
		Width(60)

	inactiveSearchStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("241")).
		Padding(0, 1).
		Width(60)

	searchLabelStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true)
)

// NotesListModel manages the notes list view
type NotesListModel struct {
	app          *App
	allNotes     []*models.Note // Store all notes
	filteredNotes []*models.Note // Store filtered notes for search
	selectedNote *models.Note
	cursor       int
	loaded       bool
	width        int
	height       int

	// Search functionality
	searchQuery string
	searchMode  bool // true when in search mode
}

// NewNotesListModel creates a new notes list model
func NewNotesListModel(app *App) *NotesListModel {
	return &NotesListModel{
		app:           app,
		allNotes:      []*models.Note{},
		filteredNotes: []*models.Note{},
		selectedNote:  nil,
		cursor:        0,
		loaded:        false,
		searchQuery:   "",
		searchMode:    false,
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

// filterNotes filters notes based on the current search query
func (m *NotesListModel) filterNotes() {
	if m.searchQuery == "" {
		// If no search query, show all notes
		m.filteredNotes = make([]*models.Note, len(m.allNotes))
		copy(m.filteredNotes, m.allNotes)
		return
	}

	// Perform fuzzy search
	searchTerms := utils.SplitWords(m.searchQuery)
	m.filteredNotes = []*models.Note{}

	for _, note := range m.allNotes {
		// Search in title and content
		titleWords := utils.SplitWords(note.Title)
		contentWords := utils.SplitWords(note.Content)

		// Check if any search term matches title or content
		if utils.ContainsAnyWord(searchTerms, titleWords) || utils.ContainsAnyWord(searchTerms, contentWords) {
			m.filteredNotes = append(m.filteredNotes, note)
		}
	}

	// Reset cursor if it's out of bounds
	if m.cursor >= len(m.filteredNotes) {
		m.cursor = 0
	}
}

// setSearchMode enables/disables search mode
func (m *NotesListModel) setSearchMode(enabled bool) {
	m.searchMode = enabled
	if enabled {
		m.cursor = 0
	} else {
		m.searchQuery = ""
		m.filterNotes() // Reset filter when exiting search mode
	}
}

// Update handles updates for the notes list
func (m *NotesListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case notesLoadedMsg:
		m.allNotes = msg.notes
		m.filterNotes() // Apply current search filter to loaded notes
		m.loaded = true
		return m.app, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+s":
			// Toggle search mode
			m.setSearchMode(!m.searchMode)
		}

		// Handle search mode input
		if m.searchMode {
			switch msg.String() {
			case "escape":
				// Exit search mode
				m.setSearchMode(false)
			case "backspace":
				if len(m.searchQuery) > 0 {
					m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
					m.filterNotes()
				}
			case "enter":
				// Exit search mode on enter
				m.setSearchMode(false)
			default:
				// Regular character input for search
				char := msg.String()
				if len(char) == 1 {
					m.searchQuery += char
					m.filterNotes()
				}
			}
		} else {
			// Normal navigation mode
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.filteredNotes)-1 {
					m.cursor++
				}
			case "n":
				// New note
				m.selectedNote = nil
				return m.app, m.app.SwitchToView(ViewNoteEditor)
			case "e", "enter":
				// Edit selected note
				if len(m.filteredNotes) > 0 {
					m.selectedNote = m.filteredNotes[m.cursor]
					return m.app, m.app.SwitchToView(ViewNoteEditor)
				}
			case "d":
				// Delete selected note
				if len(m.filteredNotes) > 0 {
					m.selectedNote = nil
					return m.app, m.deleteNote()
				}
			}
		}
	}
	return m.app, nil
}

// deleteNote deletes the currently selected note
func (m *NotesListModel) deleteNote() tea.Cmd {
	if len(m.filteredNotes) == 0 {
		return nil
	}

	selectedNote := m.filteredNotes[m.cursor]
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

	s := titleStyle.Render("Markdown Notes") + "\n\n"

	// Search interface
	if m.searchMode {
		s += searchLabelStyle.Render("Search:") + " "
		if m.searchQuery == "" {
			s += searchStyle.Render("Type to search...")
		} else {
			s += searchStyle.Render(m.searchQuery + "_") // Show cursor
		}
	} else {
		if m.searchQuery != "" {
			s += searchLabelStyle.Render("Search:") + " " + inactiveSearchStyle.Render(m.searchQuery)
			s += fmt.Sprintf(" (%d results)", len(m.filteredNotes))
		} else {
			s += searchLabelStyle.Render("Search:") + " " + inactiveSearchStyle.Render("Press Ctrl+S to search")
		}
	}

	s += "\n\n"

	// Notes list
	if len(m.filteredNotes) == 0 {
		if m.searchQuery != "" {
			s += "No notes found matching \"" + m.searchQuery + "\"\n\n"
		} else {
			s += "No notes yet. Press 'n' to create your first note.\n\n"
		}
	} else {
		// Calculate max lines for notes
		maxLines := m.height - 10 // Reserve space for header, search, controls
		if maxLines < 5 {
			maxLines = 5
		}

		displayNotes := m.filteredNotes
		if len(displayNotes) > maxLines {
			displayNotes = displayNotes[:maxLines]
		}

		for i, note := range displayNotes {
			cursor := "  "
			if m.cursor == i {
				cursor = "> "
			}

			// Truncate title if too long
			title := note.Title
			maxTitleLength := m.width - 10 // Reserve space for cursor and border
			if maxTitleLength < 20 {
				maxTitleLength = 20
			}
			if len(title) > maxTitleLength {
				title = title[:maxTitleLength-3] + "..."
			}

			s += fmt.Sprintf("%s%s\n", cursor, title)
		}

		if len(m.filteredNotes) > maxLines {
			s += fmt.Sprintf("... and %d more\n", len(m.filteredNotes)-maxLines)
		}
	}

	// Controls
	s += "\n"
	if m.searchMode {
		s += "Search Mode: Type to search • Enter to confirm • Esc to exit\n"
	} else {
		s += "Controls: n (new) • e (edit) • d (delete) • ↑↓ (navigate) • Ctrl+S (search) • q (quit) • ? (help)"
	}

	return s
}

// Messages
type notesLoadedMsg struct {
	notes []*models.Note
}
