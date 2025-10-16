package ui

import (
	"strings"

	"markdown-note-taking-app/internal/models"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Note: Styles are now defined inline with responsive design and enhanced colors

// NoteEditorModel manages the note editor view
type NoteEditorModel struct {
	app     *App
	note    *models.Note
	focused int    // 0=title, 1=content, 2=tags
	mode    string // "create" or "edit"
	width   int
	height  int

	// Text inputs for proper cursor management
	titleInput   textinput.Model
	contentInput textarea.Model
	tagInput     textinput.Model

	// Tag management
	tags             []models.Tag
	availableTags    []*models.Tag
	tagSuggestions   []string
	showSuggestions  bool
	suggestionCursor int

	// Markdown preview
	preview   *MarkdownPreviewModel
	splitPane bool // true when showing split-pane view
}

// NewNoteEditorModel creates a new note editor model
func NewNoteEditorModel(app *App) *NoteEditorModel {
	// Create text inputs with proper styling
	titleInput := textinput.New()
	titleInput.Placeholder = "Enter title..."
	titleInput.CharLimit = 100
	titleInput.Cursor.SetMode(cursor.CursorBlink)
	titleInput.Focus()
	// titleInput.PromptStyle.Foreground(lipgloss.Color("#38BDF8"))
	// titleInput.TextStyle.Foreground(lipgloss.Color("#F1F5F9"))

	contentInput := textarea.New()
	contentInput.Placeholder = "Start writing your note..."
	contentInput.CharLimit = 10000
	contentInput.Cursor.SetMode(cursor.CursorBlink)

	// Style the textarea when focused
	contentInput.FocusedStyle.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color("#64748B"))
	contentInput.FocusedStyle.Text = lipgloss.NewStyle().Foreground(lipgloss.Color("#F1F5F9"))

	// Style when unfocused
	contentInput.BlurredStyle.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color("#64748B"))
	contentInput.BlurredStyle.Text = lipgloss.NewStyle().Foreground(lipgloss.Color("#94A3B8"))

	tagInput := textinput.New()
	tagInput.Placeholder = "Add tags..."
	tagInput.CharLimit = 50
	// tagInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#38BDF8"))
	// tagInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F1F5F9"))

	return &NoteEditorModel{
		app:              app,
		note:             nil,
		focused:          0, // Start with title focused
		mode:             "create",
		titleInput:       titleInput,
		contentInput:     contentInput,
		tagInput:         tagInput,
		tags:             []models.Tag{},
		availableTags:    []*models.Tag{},
		tagSuggestions:   []string{},
		showSuggestions:  false,
		suggestionCursor: 0,
		preview:          NewMarkdownPreviewModel(),
		splitPane:        false,
	}
}

// Init initializes the note editor
func (m *NoteEditorModel) Init(selectedNote *models.Note) tea.Cmd {
	if selectedNote != nil {
		m.SetNote(selectedNote)
	} else {
		// Reset editor for new note
		m.titleInput.SetValue("")
		m.contentInput.SetValue("")
		m.tagInput.SetValue("")
		m.tags = []models.Tag{}
		m.focused = 0 // Start with title focused
		m.mode = "create"

		// Focus the title input
		m.titleInput.Focus()
		m.contentInput.Blur()
		m.tagInput.Blur()
	}

	// Reset tag suggestions
	m.showSuggestions = false
	m.suggestionCursor = 0
	return m.loadAvailableTags()
}

// loadAvailableTags loads all available tags from storage
func (m *NoteEditorModel) loadAvailableTags() tea.Cmd {
	return func() tea.Msg {
		tags, err := m.app.GetStorage().GetAllTags()
		if err != nil {
			return tagsLoadedMsg{tags: []*models.Tag{}}
		}
		return tagsLoadedMsg{tags: tags}
	}
}

// SetNote sets the editor to edit mode with an existing note
func (m *NoteEditorModel) SetNote(note *models.Note) {
	m.note = note
	m.titleInput.SetValue(note.Title)
	m.contentInput.SetValue(note.Content)
	m.tags = make([]models.Tag, len(note.Tags))
	copy(m.tags, note.Tags)
	m.focused = 0 // Start with title focused
	m.mode = "edit"

	// Focus the title input for editing
	m.titleInput.Focus()
	m.contentInput.Blur()
	m.tagInput.Blur()
}

// Update handles updates for the note editor
func (m *NoteEditorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.preview != nil {
			m.preview.Update(msg)
		}

	case tagsLoadedMsg:
		m.availableTags = msg.tags
		return m.app, nil

	case tea.KeyMsg:
		// Handle escape key
		if msg.String() == "esc" {
			if m.showSuggestions {
				m.showSuggestions = false
				m.suggestionCursor = 0
			} else {
				return m.app, m.app.SwitchToView(ViewNotesList)
			}
			return m.app, nil
		}

		// Handle save key
		if msg.String() == "ctrl+s" {
			return m.app, m.saveNote()
		}

		// Handle preview toggle
		if msg.String() == "ctrl+p" {
			m.ToggleSplitPane()
			return m.app, nil
		}

		// Handle tab navigation between fields
		if msg.String() == "tab" {
			// Cycle through 0=title, 1=tags, 2=content (reordered)
			m.focused = (m.focused + 1) % 3
			m.updateFocus()
			m.showSuggestions = false
			m.suggestionCursor = 0
			return m.app, nil
		}

		// Handle input based on focused field
		switch m.focused {
		case 0: // Title field
			m.titleInput, _ = m.titleInput.Update(msg)
		case 1: // Tags field (moved from position 2)
			m.handleTagInput(msg)
		case 2: // Content field (moved from position 1)
			m.contentInput, _ = m.contentInput.Update(msg)
		}

		// Update preview if split pane is active
		if m.splitPane {
			m.UpdatePreview()
		}
	}
	return m.app, nil
}

// saveNote saves the current note
func (m *NoteEditorModel) saveNote() tea.Cmd {
	return func() tea.Msg {
		if strings.TrimSpace(m.titleInput.Value()) == "" {
			// Don't save notes without titles
			return nil
		}

		var note *models.Note
		var err error

		if m.mode == "create" {
			note, err = m.app.GetStorage().CreateNote(m.titleInput.Value(), m.contentInput.Value())
			if err != nil {
				return nil
			}
		} else {
			// Update existing note
			if m.note != nil {
				m.note.Title = m.titleInput.Value()
				m.note.Content = m.contentInput.Value()
				err = m.app.GetStorage().UpdateNote(m.note)
				if err != nil {
					return nil
				}
				note = m.note
			}
		}

		// Save tags
		if note != nil {
			// Clear existing tags for this note
			if m.mode == "edit" && m.note != nil {
				for _, tag := range m.tags {
					m.app.GetStorage().RemoveTagFromNote(note.ID, tag.ID)
				}
			}

			// Add new tags
			for _, tag := range m.tags {
				err := m.app.GetStorage().AddTagToNote(note.ID, tag.Name)
				if err != nil {
					// For now, just ignore tag errors
					continue
				}
			}
		}

		// Go back to notes list
		return m.app.SwitchToView(ViewNotesList)()
	}
}

// Messages
type tagsLoadedMsg struct {
	tags []*models.Tag
}

// updateFocus updates the focus state of text inputs based on current focused field
func (m *NoteEditorModel) updateFocus() {
	switch m.focused {
	case 0: // Title field
		m.titleInput.Focus()
		m.tagInput.Blur()
		m.contentInput.Blur()
	case 1: // Tags field (moved from position 2)
		m.titleInput.Blur()
		m.tagInput.Focus()
		m.contentInput.Blur()
	case 2: // Content field (moved from position 1)
		m.titleInput.Blur()
		m.tagInput.Blur()
		m.contentInput.Focus()
	}
}

func (m *NoteEditorModel) handleTagInput(msg tea.KeyMsg) {
	if m.showSuggestions {
		// Handle suggestion navigation
		switch msg.String() {
		case "up":
			if m.suggestionCursor > 0 {
				m.suggestionCursor--
			}
		case "down":
			if m.suggestionCursor < len(m.tagSuggestions)-1 {
				m.suggestionCursor++
			}
		case "enter":
			// Select suggestion
			if m.suggestionCursor < len(m.tagSuggestions) {
				m.addTag(m.tagSuggestions[m.suggestionCursor])
			}
			m.showSuggestions = false
			m.suggestionCursor = 0
		case "backspace":
			// Handle backspace when suggestions are shown
			m.tagInput, _ = m.tagInput.Update(msg)
			m.updateTagSuggestions()
		default:
			// Any other input hides suggestions and goes to textinput
			m.showSuggestions = false
			m.suggestionCursor = 0
			m.tagInput, _ = m.tagInput.Update(msg)
			m.updateTagSuggestions()
		}
	} else {
		// Update textinput and check for special keys
		prevValue := m.tagInput.Value()
		m.tagInput, _ = m.tagInput.Update(msg)
		newValue := m.tagInput.Value()

		// Handle special keys that don't go through textinput normally
		switch msg.String() {
		case "enter":
			if len(newValue) > 0 {
				m.addTag(newValue)
				m.tagInput.SetValue("")
			}
		case " ":
			// Space separates tags
			if len(newValue) > 0 {
				m.addTag(newValue)
				m.tagInput.SetValue("")
			}
		default:
			// Check if value changed for suggestions
			if prevValue != newValue {
				m.updateTagSuggestions()
			}
		}
	}
}

func (m *NoteEditorModel) addTag(tagName string) {
	tagName = strings.TrimSpace(tagName)
	if tagName == "" {
		return
	}

	// Check if tag already exists
	for _, tag := range m.tags {
		if strings.EqualFold(tag.Name, tagName) {
			return // Tag already added
		}
	}

	// Add tag to current tags
	newTag := models.Tag{Name: tagName}
	m.tags = append(m.tags, newTag)

	// Clear input
	m.tagInput.SetValue("")
	m.showSuggestions = false
	m.suggestionCursor = 0
}

func (m *NoteEditorModel) updateTagSuggestions() {
	tagInputValue := m.tagInput.Value()
	if len(tagInputValue) < 2 {
		m.tagSuggestions = []string{}
		m.showSuggestions = false
		return
	}

	m.tagSuggestions = []string{}
	for _, tag := range m.availableTags {
		if strings.Contains(strings.ToLower(tag.Name), strings.ToLower(tagInputValue)) {
			// Check if tag is already added
			alreadyAdded := false
			for _, existingTag := range m.tags {
				if existingTag.ID == tag.ID {
					alreadyAdded = true
					break
				}
			}
			if !alreadyAdded {
				m.tagSuggestions = append(m.tagSuggestions, tag.Name)
			}
		}
	}

	m.showSuggestions = len(m.tagSuggestions) > 0
	m.suggestionCursor = 0
}

// ToggleSplitPane toggles the split-pane preview view
func (m *NoteEditorModel) ToggleSplitPane() {
	m.splitPane = !m.splitPane
	if m.splitPane {
		m.preview.ShowPreview(true)
		m.preview.SetContent(m.contentInput.Value())
	} else {
		m.preview.ShowPreview(false)
	}
}

// UpdatePreview updates the markdown preview with current content
func (m *NoteEditorModel) UpdatePreview() {
	if m.preview != nil {
		m.preview.SetContent(m.contentInput.Value())
	}
}

// getTagBadgeStyle returns a badge style for tags (no borders, colored backgrounds)
func (m *NoteEditorModel) getTagBadgeStyle(index int, _ string) lipgloss.Style {
	// Cycle through different background colors for variety
	switch index % 4 {
	case 0:
		// Cyan badge
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#0F172A")). // Dark text
			Background(lipgloss.Color("#38BDF8")). // Cyan background
			Padding(0, 1).
			MarginRight(1)
	case 1:
		// Green badge
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#0F172A")). // Dark text
			Background(lipgloss.Color("#4ADE80")). // Green background
			Padding(0, 1).
			MarginRight(1)
	case 2:
		// Purple badge
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#0F172A")). // Dark text
			Background(lipgloss.Color("#C084FC")). // Purple background
			Padding(0, 1).
			MarginRight(1)
	case 3:
		// Orange badge
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#0F172A")). // Dark text
			Background(lipgloss.Color("#FB923C")). // Orange background
			Padding(0, 1).
			MarginRight(1)
	default:
		// Default cyan badge
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#0F172A")). // Dark text
			Background(lipgloss.Color("#38BDF8")). // Cyan background
			Padding(0, 1).
			MarginRight(1)
	}
}

// View renders the note editor
func (m *NoteEditorModel) View() string {
	mode := "Create Note"
	if m.mode == "edit" {
		mode = "Edit Note"
	}

	if m.splitPane {
		// Split-pane view
		return m.renderSplitPaneView(mode)
	} else {
		// Single pane view
		return m.renderSinglePaneView(mode)
	}
}

// renderSinglePaneView renders the traditional single editor view with orange highlights
func (m *NoteEditorModel) renderSinglePaneView(mode string) string {
	// Define warm colors for highlighting (matching notes list)
	orangeHighlight := "#EA580C" // Orange

	// Enhanced responsive title style with warm colors
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F1F5F9")).
		Background(lipgloss.Color(orangeHighlight)).
		Bold(true).
		Padding(0, 1).
		MarginBottom(1)

	s := titleStyle.Render(mode) + "\n\n"

	// Responsive field width calculations
	fieldWidth := func() int {
		if m.width < 100 {
			return m.width - 6
		} else if m.width < 140 {
			return m.width - 8
		} else {
			return int(float64(m.width) * 0.9)
		}
	}()

	// Title field width (60-70% of available width for better balance)
	titleFieldWidth := func() int {
		if m.width < 100 {
			return int(float64(m.width) * 0.65)
		} else if m.width < 140 {
			return int(float64(m.width) * 0.6)
		} else {
			return int(float64(m.width) * 0.65)
		}
	}()

	// Title field
	titleLabel := "Title:"
	if m.focused == 0 {
		titleLabel = "[*] " + titleLabel
	} else {
		titleLabel = "[ ] " + titleLabel
	}
	s += titleLabel + "\n"

	// Set width for title input
	m.titleInput.Width = titleFieldWidth - 4 // Account for padding and border
	titleField := m.titleInput.View()

	// Apply orange border styling to title input
	titleBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(func() lipgloss.Color {
			if m.focused == 0 {
				return lipgloss.Color(orangeHighlight)
			}
			return lipgloss.Color("#475569")
		}()).
		Foreground(func() lipgloss.Color {
			if m.focused == 0 {
				return lipgloss.Color("#F1F5F9")
			}
			return lipgloss.Color("#94A3B8")
		}()).
		Padding(0, 1).
		Width(titleFieldWidth)

	s += titleBorderStyle.Render(titleField) + "\n"

	// Tags field (moved before content)
	tagsLabel := "Tags:"
	if m.focused == 1 {
		tagsLabel = "[*] " + tagsLabel
	} else {
		tagsLabel = "[ ] " + tagsLabel
	}
	s += tagsLabel + "\n"

	// Display existing tags as horizontal badges
	if len(m.tags) > 0 {
		s += " " // Start with space for better spacing
		for i, tag := range m.tags {
			badgeStyle := m.getTagBadgeStyle(i, tag.Name)
			s += badgeStyle.Render(tag.Name) + " "
		}
		s += "\n"
	}

	// Tag input without border (inline with badges)
	tagInputWidth := fieldWidth
	m.tagInput.Width = tagInputWidth - 4
	tagInputField := m.tagInput.View()

	// Simple styling for tag input with orange highlight when focused
	tagInputStyle := lipgloss.NewStyle().
		Foreground(func() lipgloss.Color {
			if m.focused == 1 {
				return lipgloss.Color("#F1F5F9")
			}
			return lipgloss.Color("#94A3B8")
		}()).
		Width(tagInputWidth)

	s += tagInputStyle.Render(tagInputField) + "\n\n"

	// Content field (moved to position 2)
	contentLabel := "Content:"
	if m.focused == 2 {
		contentLabel = "[*] " + contentLabel
	} else {
		contentLabel = "[ ] " + contentLabel
	}
	s += contentLabel + "\n"

	// Responsive content height calculation
	usedHeight := 20
	available := m.height - usedHeight - 4
	contentHeight := max(available, 5)

	// Set content textarea dimensions and get view
	contentField := m.contentInput.View()
	// Note: textarea dimensions are controlled via styling, not direct width/height assignment

	// Apply orange border styling to content area
	contentBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(func() lipgloss.Color {
			if m.focused == 2 {
				return lipgloss.Color(orangeHighlight)
			}
			return lipgloss.Color("#475569")
		}()).
		Foreground(func() lipgloss.Color {
			if m.focused == 2 {
				return lipgloss.Color("#F1F5F9")
			}
			return lipgloss.Color("#94A3B8")
		}()).
		Padding(1).
		Width(fieldWidth).
		Height(contentHeight)

	s += contentBorderStyle.Render(contentField)

	// Enhanced controls with responsive layout
	s += "\n\n"
	controlsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#94A3B8")).
		MarginTop(1)

	controls := "Tab - Switch fields • Ctrl+S - Save • Ctrl+P - Toggle preview • Esc - Cancel"
	if m.width < 100 {
		controls = "Tab: Switch • Ctrl+S: Save • Ctrl+P: Preview • Esc: Cancel"
	}
	s += controlsStyle.Render(controls) + "\n"

	if m.focused == 1 {
		tagHelp := "Tags: Type to add • Space/Enter to confirm • ↑↓ to navigate suggestions"
		if m.width < 100 {
			tagHelp = "Tags: Type • Space/Enter to add • ↑↓ for suggestions"
		}
		s += controlsStyle.Render(tagHelp) + "\n"
	}

	// Enhanced tag suggestions with orange accent
	if m.showSuggestions && len(m.tagSuggestions) > 0 {
		suggestionBox := ""
		maxSuggestions := 5
		for i, suggestion := range m.tagSuggestions {
			if i >= maxSuggestions {
				break
			}
			prefix := "  "
			if i == m.suggestionCursor {
				prefix = "> "
			}
			suggestionBox += prefix + suggestion + "\n"
		}
		suggestionStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(orangeHighlight)).
			Background(lipgloss.Color("#0F172A")).
			Foreground(lipgloss.Color("#F1F5F9")).
			Padding(0, 1)
		s += suggestionStyle.Render(suggestionBox)
	}

	return s
}

// renderSplitPaneView renders the split-pane editor view with orange highlights
func (m *NoteEditorModel) renderSplitPaneView(mode string) string {
	// Define warm colors for highlighting (matching notes list)
	orangeHighlight := "#EA580C" // Orange

	// Enhanced responsive title style with warm colors
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F1F5F9")).
		Background(lipgloss.Color(orangeHighlight)).
		Bold(true).
		Padding(0, 1).
		MarginBottom(1)

	s := titleStyle.Render(mode+" - Split View") + "\n\n"

	// Responsive pane width calculations
	editorWidth := (m.width - 8) / 2          // Account for borders and spacing
	previewWidth := m.width - editorWidth - 4 // Leave space for borders

	// Enhanced editor pane with orange accent
	editorPane := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(orangeHighlight)). // Orange accent
		Width(editorWidth).
		Height(m.height - 8).
		Padding(1)

	editorContent := m.renderEditorContent(editorWidth-4, m.height-10)
	editorBox := editorPane.Render(editorContent)

	// Enhanced preview pane with orange accent
	previewPane := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(orangeHighlight)). // Orange accent
		Width(previewWidth).
		Height(m.height - 8).
		Padding(1)

	previewContent := m.preview.View()
	previewBox := previewPane.Render(previewContent)

	// Combine panes side by side
	s += lipgloss.JoinHorizontal(lipgloss.Left, editorBox, previewBox)

	// Enhanced controls with responsive layout
	s += "\n\n"
	controlsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#94A3B8")).
		MarginTop(1)

	controls := "Tab: Switch fields • Ctrl+S: Save • Ctrl+P: Exit preview • Esc: Cancel"
	if m.width < 120 {
		controls = "Tab: Switch • Ctrl+S: Save • Ctrl+P: Exit • Esc: Cancel"
	}
	s += controlsStyle.Render(controls)

	return s
}

// renderEditorContent renders the editor content for split-pane view with orange highlights
func (m *NoteEditorModel) renderEditorContent(width, height int) string {
	// Define warm colors for highlighting (matching notes list)
	orangeHighlight := "#EA580C" // Orange

	s := ""

	// Label style
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#94A3B8")).
		Bold(true).
		MarginBottom(0).Padding(0, 1)

	// Responsive field dimensions
	fieldWidth := width - 4 // Account for padding

	// Title section
	titleLabel := "Title:"
	if m.focused == 0 {
		titleLabel = "[*] " + titleLabel
	} else {
		titleLabel = "[ ] " + titleLabel
	}
	s += labelStyle.Render(titleLabel) + "\n"

	// Title input with border
	m.titleInput.Width = fieldWidth
	titleField := m.titleInput.View()

	// Apply orange border styling to title input
	titleBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(func() lipgloss.Color {
			if m.focused == 0 {
				return lipgloss.Color(orangeHighlight)
			}
			return lipgloss.Color("#475569")
		}()).
		Foreground(func() lipgloss.Color {
			if m.focused == 0 {
				return lipgloss.Color("#F1F5F9")
			}
			return lipgloss.Color("#94A3B8")
		}()).
		Width(fieldWidth + 2) // Account for border padding

	s += titleBorderStyle.Render(titleField) + "\n"

	// Tags section (moved before content)
	tagsLabel := "Tags:"
	if m.focused == 1 {
		tagsLabel = "[*] " + tagsLabel
	} else {
		tagsLabel = "[ ] " + tagsLabel
	}
	s += labelStyle.Render(tagsLabel) + "\n"

	// Display existing tags as horizontal badges
	if len(m.tags) > 0 {
		s += " " // Start with space for better spacing
		for i, tag := range m.tags {
			badgeStyle := m.getTagBadgeStyle(i, tag.Name)
			s += badgeStyle.Render(tag.Name) + " "
		}
		s += "\n"
	}

	// Tag input without border (inline with badges)
	tagInputWidth := fieldWidth
	m.tagInput.Width = tagInputWidth - 4
	tagInputField := m.tagInput.View()

	// Simple styling for tag input with orange highlight when focused
	tagInputStyle := lipgloss.NewStyle().
		Foreground(func() lipgloss.Color {
			if m.focused == 1 {
				return lipgloss.Color("#F1F5F9")
			}
			return lipgloss.Color("#94A3B8")
		}()).
		Width(tagInputWidth)

	s += tagInputStyle.Render(tagInputField) + "\n\n"

	// Content section (moved to position 2)
	contentLabel := "Content:"
	if m.focused == 2 {
		contentLabel = "[*] " + contentLabel
	} else {
		contentLabel = "[ ] " + contentLabel
	}
	s += labelStyle.Render(contentLabel) + "\n"

	// Calculate content height (remaining space after other fields)
	usedHeight := 12 // Approximate height used by title, tags, labels
	contentHeight := max(height-usedHeight, 5)

	// Content input with border and responsive height
	contentField := m.contentInput.View()
	// Note: textarea dimensions are controlled via styling, not direct width/height assignment

	// Apply orange border styling to content area
	contentBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(func() lipgloss.Color {
			if m.focused == 2 {
				return lipgloss.Color(orangeHighlight)
			}
			return lipgloss.Color("#475569")
		}()).
		Foreground(func() lipgloss.Color {
			if m.focused == 2 {
				return lipgloss.Color("#F1F5F9")
			}
			return lipgloss.Color("#94A3B8")
		}()).
		Padding(1).
		Width(fieldWidth + 2).    // Account for border padding
		Height(contentHeight + 2) // Account for border padding

	s += contentBorderStyle.Render(contentField)

	return s
}
