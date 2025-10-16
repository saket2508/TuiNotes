package ui

import (
	"fmt"
	"strings"

	"markdown-note-taking-app/internal/models"
	"markdown-note-taking-app/internal/utils"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// NotesListModel manages the notes list view
type NotesListModel struct {
	app           *App
	allNotes      []*models.Note // Store all notes
	filteredNotes []*models.Note // Store filtered notes for search
	selectedNote  *models.Note
	cursor        int
	loaded        bool
	width         int
	height        int

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
			case "n", "N":
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
			case "h", "H":
				// Help
				return m.app, m.app.SwitchToView(ViewHelp)
			case "ctrl+c":
				// Quit
				return m.app, tea.Quit
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

// renderGradientHeader creates a beautiful gradient Noteshell header
func (m *NotesListModel) renderGradientHeader() string {
	// ASCII art for Noteshell with gradient colors
	asciiArt := []string{
		"██████╗  ██╗   ██╗██╗██╗     ██╗     ███╗   ██╗ ██████╗ ████████╗███████╗███████╗",
		"██╔═══██╗██║   ██║██║██║     ██║     ████╗  ██║██╔═══██╗╚══██╔══╝██╔════╝██╔════╝",
		"██║   ██║██║   ██║██║██║     ██║     ██╔██╗ ██║██║   ██║   ██║   █████╗  ███████╗",
		"██║▄▄ ██║██║   ██║██║██║     ██║     ██║╚██╗██║██║   ██║   ██║   ██╔══╝  ╚════██║",
		"╚██████╔╝╚██████╔╝██║███████╗███████╗██║ ╚████║╚██████╔╝   ██║   ███████╗███████║",
		" ╚══▀▀═╝  ╚═════╝ ╚═╝╚══════╝╚══════╝╚═╝  ╚═══╝ ╚═════╝    ╚═╝   ╚══════╝╚══════╝",
	}

	// Gradient colors from orange to amber to yellow (warm theme)
	colors := []string{
		"#EA580C", // Deep Orange
		"#F97316", // Orange
		"#FB923C", // Light Orange
		"#F59E0B", // Amber
		"#FBBF24", // Yellow
		"#FCD34D", // Light Yellow
	}

	// Apply gradient to each line
	var gradientLines []string
	for i, line := range asciiArt {
		color := colors[i%len(colors)]
		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color(color)).
			Bold(true)
		gradientLines = append(gradientLines, style.Render(line))
	}

	// Subtitle with elegant typography (reduced margin)
	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#94A3B8")).
		Italic(true).
		MarginTop(0).
		MarginBottom(0)

	subtitle := subtitleStyle.Render("  ── Your terminal-based markdown note-taking shell ──")

	// Combine all parts
	header := strings.Join(gradientLines, "\n")
	return header + "\n" + subtitle
}

// renderQuickActions creates minimal keyboard shortcuts info
func (m *NotesListModel) renderQuickActions() string {
	// Minimal shortcuts display
	shortcutsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#64748B")).
		Italic(true).
		MarginBottom(1)

	shortcuts := shortcutsStyle.Render("N: New • S: Search • ↑↓: Navigate • Enter: Edit • Ctrl+C: Quit")
	return shortcuts
}


// View renders the notes list with centered layout and orange/yellow highlighting
func (m *NotesListModel) View() string {
	if !m.loaded {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#94A3B8")).
			Bold(true).
			Render("Loading notes...")
	}

	// Define warm colors for highlighting
	orangeHighlight := "#EA580C" // Orange

	// Search styling - redesigned to look like an input field
	searchActiveStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(orangeHighlight)).
		Background(lipgloss.Color("#1F2937")). // Dark background like input fields
		Foreground(lipgloss.Color("#F1F5F9")).
		Padding(0, 2).
		Width(40) // Fixed width like a proper input field

	searchInactiveStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#475569")).
		Background(lipgloss.Color("#0F172A")). // Subtle background
		Foreground(lipgloss.Color("#94A3B8")).
		Padding(0, 2).
		Width(40) // Consistent width

	searchLabelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(orangeHighlight)).
		Bold(true)

	// Build the content
	content := m.renderGradientHeader() + "\n\n"

	// Minimal shortcuts
	content += m.renderQuickActions() + "\n\n"

	// Search interface - redesigned as an input field
	content += searchLabelStyle.Render("Search:") + "\n"
	if m.searchMode {
		if m.searchQuery == "" {
			// Active state with placeholder
			placeholderStyle := searchActiveStyle.
				Foreground(lipgloss.Color("#64748B")) // Dimmed placeholder text
			content += placeholderStyle.Render("Type your search query...")
		} else {
			// Active state with cursor
			cursorStyle := searchActiveStyle.
				Foreground(lipgloss.Color("#F1F5F9"))
			content += cursorStyle.Render(m.searchQuery + "▏") // Better cursor indicator
		}
	} else {
		if m.searchQuery != "" {
			// Show search query with results count
			content += searchInactiveStyle.Render(m.searchQuery)
			content += lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F59E0B")).
				Render(fmt.Sprintf(" (%d results)", len(m.filteredNotes)))
		} else {
			// Inactive state with prompt
			promptStyle := searchInactiveStyle.
				Foreground(lipgloss.Color("#64748B"))
			content += promptStyle.Render("Press Ctrl+S to search")
		}
	}

	content += "\n\n"

	// Notes list with orange/yellow highlighting
	if len(m.filteredNotes) == 0 {
		if m.searchQuery != "" {
			content += lipgloss.NewStyle().
				Foreground(lipgloss.Color("#94A3B8")).
				Italic(true).
				Render("No notes found matching \"" + m.searchQuery + "\"")
		} else {
			content += lipgloss.NewStyle().
				Foreground(lipgloss.Color("#94A3B8")).
				Italic(true).
				Render("No notes yet. Press 'n' to create your first note.")
		}
	} else {
		// Calculate responsive max lines
		usedHeight := 6 // Reduced from 8
		available := m.height - usedHeight - 4
		maxLines := max(available, 5)

		displayNotes := m.filteredNotes
		if len(displayNotes) > maxLines {
			displayNotes = displayNotes[:maxLines]
		}

		// Calculate responsive title length (more generous)
		maxTitleLength := func() int {
			if m.width < 80 {
				return m.width - 15
			} else if m.width < 120 {
				return m.width - 20
			} else {
				return 60 // Cap at 60 for readability
			}
		}()

		for i, note := range displayNotes {
			// Orange/amber cursor for selected item
			cursor := "  "
			if m.cursor == i {
				cursor = lipgloss.NewStyle().
					Foreground(lipgloss.Color(orangeHighlight)).
					Bold(true).
					Render("▶ ")
			}

			// Truncate title
			title := note.Title
			if len(title) > maxTitleLength {
				title = title[:maxTitleLength-3] + "..."
			}

			// Apply orange/yellow highlighting for selected notes
			itemStyle := lipgloss.NewStyle()
			if m.cursor == i {
				// Orange to amber gradient background
				itemStyle = itemStyle.
					Background(lipgloss.Color(orangeHighlight)).
					Foreground(lipgloss.Color("#0F172A")).
					Bold(true).
					Padding(0, 1).
					MarginLeft(1).
					MarginRight(1)
			} else {
				// Subtle yellow background for non-selected
				itemStyle = itemStyle.
					Background(lipgloss.Color("#1F2937")). // Dark background
					Foreground(lipgloss.Color("#F1F5F9")).
					Padding(0, 1).
					MarginLeft(1).
					MarginRight(1)
			}

			content += cursor + itemStyle.Render(title) + "\n"
		}

		if len(m.filteredNotes) > maxLines {
			content += lipgloss.NewStyle().
				Foreground(lipgloss.Color("#64748B")).
				Italic(true).
				Render(fmt.Sprintf("... and %d more", len(m.filteredNotes)-maxLines))
		}
	}

	// Wrap everything in a centered container
	containerWidth := min(m.width-4, 100) // Max 100 chars width
	containerStyle := lipgloss.NewStyle().
		Width(containerWidth).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#334155")).
		Padding(2, 2).
		Background(lipgloss.Color("#0F172A"))

	centeredContent := lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			containerStyle.Render(content),
	)

	return centeredContent
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Messages
type notesLoadedMsg struct {
	notes []*models.Note
}
