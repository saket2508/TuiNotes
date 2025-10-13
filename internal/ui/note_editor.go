package ui

import (
	"strings"

	"markdown-note-taking-app/internal/models"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Note: Styles are now defined inline with responsive design and enhanced colors

// NoteEditorModel manages the note editor view
type NoteEditorModel struct {
	app     *App
	note    *models.Note
	title   string
	content string
	focused int // 0=title, 1=content, 2=tags
	cursor  int
	mode    string // "create" or "edit"
	width   int
	height  int

	// Tag management
	tags         []models.Tag
	tagInput     string
	availableTags []*models.Tag
	tagSuggestions []string
	showSuggestions bool
	suggestionCursor int

	// Markdown preview
	preview *MarkdownPreviewModel
	splitPane bool // true when showing split-pane view
}

// NewNoteEditorModel creates a new note editor model
func NewNoteEditorModel(app *App) *NoteEditorModel {
	return &NoteEditorModel{
		app:              app,
		note:             nil,
		title:            "",
		content:          "",
		focused:          0, // Start with title focused
		cursor:           0,
		mode:             "create",
		tags:             []models.Tag{},
		tagInput:         "",
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
		m.title = ""
		m.content = ""
		m.tags = []models.Tag{}
		m.focused = 0 // Start with title focused
		m.mode = "create"
	}

	// Load available tags and reset tag input
	m.tagInput = ""
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
	m.title = note.Title
	m.content = note.Content
	m.tags = make([]models.Tag, len(note.Tags))
	copy(m.tags, note.Tags)
	m.focused = 0 // Start with title focused
	m.mode = "edit"
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
			m.focused = (m.focused + 1) % 3 // Cycle through 0=title, 1=content, 2=tags
			m.showSuggestions = false
			m.suggestionCursor = 0
			return m.app, nil
		}

		// Handle input based on focused field
		switch m.focused {
		case 0: // Title field
			m.handleTitleInput(msg)
		case 1: // Content field
			m.handleContentInput(msg)
		case 2: // Tags field
			m.handleTagInput(msg)
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
		if strings.TrimSpace(m.title) == "" {
			// Don't save notes without titles
			return nil
		}

		var note *models.Note
		var err error

		if m.mode == "create" {
			note, err = m.app.GetStorage().CreateNote(m.title, m.content)
			if err != nil {
				return nil
			}
		} else {
			// Update existing note
			if m.note != nil {
				m.note.Title = m.title
				m.note.Content = m.content
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

// Input handlers
func (m *NoteEditorModel) handleTitleInput(msg tea.KeyMsg) {
	switch msg.String() {
	case "backspace":
		if len(m.title) > 0 {
			m.title = m.title[:len(m.title)-1]
		}
	default:
		// Regular character input
		char := msg.String()
		if len(char) == 1 {
			m.title += char
		}
	}
}

func (m *NoteEditorModel) handleContentInput(msg tea.KeyMsg) {
	switch msg.String() {
	case "backspace":
		if len(m.content) > 0 {
			m.content = m.content[:len(m.content)-1]
		}
	case "enter":
		// New line in content
		m.content += "\n"
	default:
		// Regular character input
		char := msg.String()
		if len(char) == 1 {
			m.content += char
		}
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
		default:
			// Any other input hides suggestions
			m.showSuggestions = false
			m.suggestionCursor = 0
			m.handleTagTextInput(msg)
		}
	} else {
		m.handleTagTextInput(msg)
	}
}

func (m *NoteEditorModel) handleTagTextInput(msg tea.KeyMsg) {
	switch msg.String() {
	case "backspace":
		if len(m.tagInput) > 0 {
			m.tagInput = m.tagInput[:len(m.tagInput)-1]
		}
	case "enter":
		if len(m.tagInput) > 0 {
			m.addTag(m.tagInput)
		}
	case " ":
		// Space separates tags
		if len(m.tagInput) > 0 {
			m.addTag(m.tagInput)
		}
	default:
		// Regular character input
		char := msg.String()
		if len(char) == 1 {
			m.tagInput += char
			m.updateTagSuggestions()
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
	m.tagInput = ""
	m.showSuggestions = false
	m.suggestionCursor = 0
}

func (m *NoteEditorModel) updateTagSuggestions() {
	if len(m.tagInput) < 2 {
		m.tagSuggestions = []string{}
		m.showSuggestions = false
		return
	}

	m.tagSuggestions = []string{}
	for _, tag := range m.availableTags {
		if strings.Contains(strings.ToLower(tag.Name), strings.ToLower(m.tagInput)) {
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
		m.preview.SetContent(m.content)
	} else {
		m.preview.ShowPreview(false)
	}
}

// UpdatePreview updates the markdown preview with current content
func (m *NoteEditorModel) UpdatePreview() {
	if m.preview != nil {
		m.preview.SetContent(m.content)
	}
}

// getTagStyle returns a tag style based on the tag index for variety
func (m *NoteEditorModel) getTagStyle(index int, _ string) lipgloss.Style {
	// Cycle through different colors based on index for variety
	switch index % 4 {
	case 0:
		// Cyan tag
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#38BDF8")).
			Background(lipgloss.Color("#0F172A")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#0EA5E9")).
			Padding(0, 1).
			MarginRight(1)
	case 1:
		// Green tag
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4ADE80")).
			Background(lipgloss.Color("#0F172A")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#22C55E")).
			Padding(0, 1).
			MarginRight(1)
	case 2:
		// Purple tag
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#C084FC")).
			Background(lipgloss.Color("#0F172A")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#A855F7")).
			Padding(0, 1).
			MarginRight(1)
	case 3:
		// Orange tag
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FB923C")).
			Background(lipgloss.Color("#0F172A")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#F97316")).
			Padding(0, 1).
			MarginRight(1)
	default:
		// Default cyan tag
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#38BDF8")).
			Background(lipgloss.Color("#0F172A")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#0EA5E9")).
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

// renderSinglePaneView renders the traditional single editor view
func (m *NoteEditorModel) renderSinglePaneView(mode string) string {
	// Enhanced responsive title style
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F1F5F9")).
		Background(lipgloss.Color("#38BDF8")).
		Bold(true).
		Padding(0, 1).
		MarginBottom(1)

	s := titleStyle.Render(mode) + "\n\n"

	// Enhanced field styles
	activeFieldStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#38BDF8")).
		Foreground(lipgloss.Color("#F1F5F9")).
		Padding(0, 1)

	inactiveFieldStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#475569")).
		Foreground(lipgloss.Color("#94A3B8")).
		Padding(0, 1)

	tagLabelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#94A3B8")).
		Bold(true).
		MarginBottom(1)

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

	// Title field
	titleLabel := "Title:"
	if m.focused == 0 {
		titleLabel = "[*] " + titleLabel
	} else {
		titleLabel = "[ ] " + titleLabel
	}
	s += titleLabel + "\n"

	var titleField string
	if m.focused == 0 {
		if m.title == "" {
			titleField = activeFieldStyle.Width(fieldWidth).Render("Enter title...")
		} else {
			titleField = activeFieldStyle.Width(fieldWidth).Render(m.title)
		}
	} else {
		titleField = inactiveFieldStyle.Width(fieldWidth).Render(m.title)
	}
	s += titleField + "\n\n"

	// Tags field
	s += tagLabelStyle.Render("Tags:") + "\n"

	// Display existing tags
	for i, tag := range m.tags {
		style := m.getTagStyle(i, tag.Name)
		s += style.Render(tag.Name)
	}

	// Responsive tag input width
	tagInputWidth := func() int {
		if m.width < 100 {
			width := m.width - 20
			if width < 20 {
				return 20
			}
			return width
		} else if m.width < 140 {
			return 40
		} else {
			return 50
		}
	}()

	// Enhanced tag input styles
	tagInputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#38BDF8")).
		Foreground(lipgloss.Color("#F1F5F9")).
		Padding(0, 1).
		Width(tagInputWidth)

	inactiveTagInputStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#475569")).
		Foreground(lipgloss.Color("#64748B")).
		Padding(0, 1).
		Width(tagInputWidth)

	var tagInputField string
	if m.focused == 2 {
		if m.tagInput == "" {
			tagInputField = tagInputStyle.Render("Add tags...")
		} else {
			tagInputField = tagInputStyle.Render(m.tagInput + "_")
		}
	} else {
		if m.tagInput == "" {
			tagInputField = inactiveTagInputStyle.Render("")
		} else {
			tagInputField = inactiveTagInputStyle.Render(m.tagInput)
		}
	}
	s += tagInputField + "\n"

	// Enhanced tag suggestions
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
			BorderForeground(lipgloss.Color("#38BDF8")).
			Background(lipgloss.Color("#0F172A")).
			Foreground(lipgloss.Color("#F1F5F9")).
			Padding(0, 1)
		s += suggestionStyle.Render(suggestionBox)
	}

	s += "\n"

	// Content field
	contentLabel := "Content:"
	if m.focused == 1 {
		contentLabel = "[*] " + contentLabel
	} else {
		contentLabel = "[ ] " + contentLabel
	}
	s += contentLabel + "\n"

	// Responsive content height calculation
	usedHeight := 20
	available := m.height - usedHeight - 4
	contentHeight := available
	if contentHeight < 5 {
		contentHeight = 5
	}

	var displayContent string
	if m.content == "" && m.focused == 1 {
		displayContent = "Start writing your note..."
	} else if m.content == "" {
		displayContent = ""
	} else {
		displayContent = m.content
	}

	// Enhanced content styles
	contentActiveStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#38BDF8")).
		Foreground(lipgloss.Color("#F1F5F9")).
		Padding(1)

	contentInactiveStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#475569")).
		Foreground(lipgloss.Color("#94A3B8")).
		Padding(1)

	var contentBoxStyle lipgloss.Style
	if m.focused == 1 {
		contentBoxStyle = contentActiveStyle
	} else {
		contentBoxStyle = contentInactiveStyle
	}

	contentBox := contentBoxStyle.Width(fieldWidth).Height(contentHeight).Render(displayContent)
	s += contentBox

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

	if m.focused == 2 {
		tagHelp := "Tags: Type to add • Space/Enter to confirm • ↑↓ to navigate suggestions"
		if m.width < 100 {
			tagHelp = "Tags: Type • Space/Enter to add • ↑↓ for suggestions"
		}
		s += controlsStyle.Render(tagHelp) + "\n"
	}

	return s
}

// renderSplitPaneView renders the split-pane editor view
func (m *NoteEditorModel) renderSplitPaneView(mode string) string {
	// Enhanced responsive title style
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F1F5F9")).
		Background(lipgloss.Color("#38BDF8")).
		Bold(true).
		Padding(0, 1).
		MarginBottom(1)

	s := titleStyle.Render(mode + " - Split View") + "\n\n"

	// Responsive pane width calculations
	editorWidth := (m.width - 8) / 2     // Account for borders and spacing
	previewWidth := m.width - editorWidth - 4 // Leave space for borders

	// Enhanced editor pane
	editorPane := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#475569")).
		Width(editorWidth).
		Height(m.height - 8).
		Padding(1)

	editorContent := m.renderEditorContent(editorWidth - 4, m.height - 10)
	editorBox := editorPane.Render(editorContent)

	// Enhanced preview pane with cyan accent
	previewPane := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#38BDF8")). // Cyan accent
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

// renderEditorContent renders the editor content for split-pane view
func (m *NoteEditorModel) renderEditorContent(width, height int) string {
	s := ""

	// Enhanced field styles for split-pane
	activeFieldStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#38BDF8")).
		Foreground(lipgloss.Color("#F1F5F9")).
		Padding(0, 1)

	inactiveFieldStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#475569")).
		Foreground(lipgloss.Color("#94A3B8")).
		Padding(0, 1)

	tagLabelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#94A3B8")).
		Bold(true).
		MarginBottom(1)

	// Title
	titleLabel := "Title:"
	if m.focused == 0 {
		titleLabel = "[*] " + titleLabel
	} else {
		titleLabel = "[ ] " + titleLabel
	}
	s += titleLabel + "\n"

	var titleField string
	if m.focused == 0 {
		if m.title == "" {
			titleField = activeFieldStyle.Width(width).Render("Enter title...")
		} else {
			titleField = activeFieldStyle.Width(width).Render(m.title)
		}
	} else {
		titleField = inactiveFieldStyle.Width(width).Render(m.title)
	}
	s += titleField + "\n\n"

	// Tags
	s += tagLabelStyle.Render("Tags:") + "\n"

	// Display tags
	for i, tag := range m.tags {
		style := m.getTagStyle(i, tag.Name)
		s += style.Render(tag.Name)
	}

	// Responsive tag input width for split-pane
	tagInputWidth := width - 8
	if tagInputWidth < 20 {
		tagInputWidth = 20
	}

	// Enhanced tag input styles
	tagInputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#38BDF8")).
		Foreground(lipgloss.Color("#F1F5F9")).
		Padding(0, 1).
		Width(tagInputWidth)

	inactiveTagInputStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#475569")).
		Foreground(lipgloss.Color("#64748B")).
		Padding(0, 1).
		Width(tagInputWidth)

	var tagInputField string
	if m.focused == 2 {
		if m.tagInput == "" {
			tagInputField = tagInputStyle.Render("Add tags...")
		} else {
			tagInputField = tagInputStyle.Render(m.tagInput + "_")
		}
	} else {
		if m.tagInput == "" {
			tagInputField = inactiveTagInputStyle.Render("")
		} else {
			tagInputField = inactiveTagInputStyle.Render(m.tagInput)
		}
	}
	s += tagInputField + "\n"

	// Content
	contentLabel := "Content:"
	if m.focused == 1 {
		contentLabel = "[*] " + contentLabel
	} else {
		contentLabel = "[ ] " + contentLabel
	}
	s += contentLabel + "\n"

	// Available height for content
	contentHeight := height - 20 // Account for other fields
	if contentHeight < 5 {
		contentHeight = 5
	}

	var displayContent string
	if m.content == "" && m.focused == 1 {
		displayContent = "Start writing..."
	} else if m.content == "" {
		displayContent = ""
	} else {
		displayContent = m.content
	}

	// Enhanced content styles for split-pane
	contentActiveStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#38BDF8")).
		Foreground(lipgloss.Color("#F1F5F9")).
		Padding(1)

	contentInactiveStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#475569")).
		Foreground(lipgloss.Color("#94A3B8")).
		Padding(1)

	var contentBoxStyle lipgloss.Style
	if m.focused == 1 {
		contentBoxStyle = contentActiveStyle
	} else {
		contentBoxStyle = contentInactiveStyle
	}

	contentBox := contentBoxStyle.Width(width).Height(contentHeight).Render(displayContent)
	s += contentBox

	return s
}
