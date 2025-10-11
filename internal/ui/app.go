package ui

import (
	"fmt"

	"markdown-note-taking-app/internal/storage"

	tea "github.com/charmbracelet/bubbletea"
)

// View represents different application views
type View int

const (
	ViewNotesList View = iota
	ViewNoteEditor
	ViewHelp
)

// App represents the main application
type App struct {
	storage     *storage.Service
	currentView View
	notesList   *NotesListModel
	noteEditor  *NoteEditorModel
	help        *HelpModel
	width       int
	height      int
}

// NewApp creates a new application instance
func NewApp(dbPath string) (*App, error) {
	// Initialize storage
	storageService, err := storage.NewService(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	app := &App{
		storage:     storageService,
		currentView: ViewNotesList,
	}

	// Initialize view models
	app.notesList = NewNotesListModel(app)
	app.noteEditor = NewNoteEditorModel(app)
	app.help = NewHelpModel(app)

	return app, nil
}

// Close closes the application and cleans up resources
func (a *App) Close() error {
	return a.storage.Close()
}

// Init initializes the application
func (a *App) Init() tea.Cmd {
	return a.notesList.Init()
}

// Update handles application-wide updates and view switching
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		// Update all views with new dimensions
		a.notesList.Update(msg)
		a.noteEditor.Update(msg)
		a.help.Update(msg)
		return a, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return a, tea.Quit
		case "?":
			a.currentView = ViewHelp
			return a, nil
		case "esc":
			// Go back to notes list from any view
			if a.currentView != ViewNotesList {
				a.currentView = ViewNotesList
				return a, a.notesList.Init()
			}
		}
	}

	// Route updates to current view
	switch a.currentView {
	case ViewNotesList:
		return a.notesList.Update(msg)
	case ViewNoteEditor:
		return a.noteEditor.Update(msg)
	case ViewHelp:
		return a.help.Update(msg)
	default:
		return a, nil
	}
}

// View renders the current view
func (a *App) View() string {
	switch a.currentView {
	case ViewNotesList:
		return a.notesList.View()
	case ViewNoteEditor:
		return a.noteEditor.View()
	case ViewHelp:
		return a.help.View()
	default:
		return "Unknown view"
	}
}

// SwitchToView switches to a different view
func (a *App) SwitchToView(view View) tea.Cmd {
	a.currentView = view
	switch view {
	case ViewNotesList:
		return a.notesList.Init()
	case ViewNoteEditor:
		return a.noteEditor.Init(a.notesList.selectedNote)
	case ViewHelp:
		return a.help.Init()
	default:
		return nil
	}
}

// GetStorage returns the storage service
func (a *App) GetStorage() *storage.Service {
	return a.storage
}
