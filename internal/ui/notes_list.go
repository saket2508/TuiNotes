package ui

import (
	"fmt"

	"markdown-note-taking-app/internal/models"
	"markdown-note-taking-app/internal/utils"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

	// Use enhanced colors but keep current structure for now
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F1F5F9")).
		Background(lipgloss.Color("#38BDF8")).
		Bold(true).
		Padding(0, 1).
		MarginBottom(1)

	searchActiveStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#38BDF8")).
		Foreground(lipgloss.Color("#F1F5F9")).
		Padding(0, 1)

	searchInactiveStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#475569")).
		Foreground(lipgloss.Color("#94A3B8")).
		Padding(0, 1)

	searchLabelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#38BDF8")).
		Bold(true)

	s := titleStyle.Render("Markdown Notes") + "\n\n"

	// Calculate responsive search width
	searchWidth := func() int {
		if m.width < 100 {
			if m.width-10 < 40 {
				return 40
			}
			return m.width - 10
		} else if m.width < 140 {
			width := int(float64(m.width) * 0.7)
			if width < 50 {
				return 50
			}
			if width > 80 {
				return 80
			}
			return width
		} else {
			width := int(float64(m.width) * 0.6)
			if width < 60 {
				return 60
			}
			if width > 100 {
				return 100
			}
			return width
		}
	}()

	// Search interface
	if m.searchMode {
		s += searchLabelStyle.Render("Search:") + " "
		if m.searchQuery == "" {
			s += searchActiveStyle.Width(searchWidth).Render("Type to search...")
		} else {
			s += searchActiveStyle.Width(searchWidth).Render(m.searchQuery + "_")
		}
	} else {
		s += searchLabelStyle.Render("Search:") + " "
		if m.searchQuery != "" {
			s += searchInactiveStyle.Width(searchWidth).Render(m.searchQuery)
			s += fmt.Sprintf(" (%d results)", len(m.filteredNotes))
		} else {
			s += searchInactiveStyle.Width(searchWidth).Render("Press Ctrl+S to search")
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
		// Calculate responsive max lines
		usedHeight := 8
		available := m.height - usedHeight - 4
		maxLines := available
		if maxLines < 5 {
			maxLines = 5
		}

		displayNotes := m.filteredNotes
		if len(displayNotes) > maxLines {
			displayNotes = displayNotes[:maxLines]
		}

		// Calculate responsive title length
		maxTitleLength := func() int {
			if m.width < 100 {
				length := m.width - 8
				if length < 20 {
					return 20
				}
				if length > 40 {
					return 40
				}
				return length
			} else if m.width < 140 {
				length := m.width - 10
				if length < 30 {
					return 30
				}
				if length > 60 {
					return 60
				}
				return length
			} else {
				length := m.width - 12
				if length < 40 {
					return 40
				}
				if length > 80 {
					return 80
				}
				return length
			}
		}()

		for i, note := range displayNotes {
			cursor := "  "
			if m.cursor == i {
				cursor = "> "
			}

			// Truncate title
			title := note.Title
			if len(title) > maxTitleLength {
				title = title[:maxTitleLength-3] + "..."
			}

			// Apply selection styling
			itemStyle := lipgloss.NewStyle()
			if m.cursor == i {
				itemStyle = itemStyle.Background(lipgloss.Color("#1E293B")).Padding(0, 1)
			}

			s += cursor + itemStyle.Render(title) + "\n"
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
		controls := "n (new) • e (edit) • d (delete) • ↑↓ (navigate) • Ctrl+S (search) • q (quit) • ? (help)"
		if m.width < 100 {
			controls = "n new • e edit • d delete • ↑↓ navigate • Ctrl+S search • q quit • ? help"
		}
		s += controls + "\n"
	}

	return s
}

// Messages
type notesLoadedMsg struct {
	notes []*models.Note
}
