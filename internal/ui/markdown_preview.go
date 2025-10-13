package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	previewStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1).
		MarginLeft(1)

	previewTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1).
			MarginBottom(1)

	previewContentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F1F5F9"))
)

// MarkdownPreviewModel manages the markdown preview view
type MarkdownPreviewModel struct {
	content    string
	rendered   string
	width      int
	height     int
	scrollPos  int
	showPreview bool
}

// NewMarkdownPreviewModel creates a new markdown preview model
func NewMarkdownPreviewModel() *MarkdownPreviewModel {
	return &MarkdownPreviewModel{
		content:     "",
		rendered:    "",
		width:       80,
		height:      24,
		scrollPos:   0,
		showPreview: false,
	}
}

// SetContent updates the markdown content and re-renders it
func (m *MarkdownPreviewModel) SetContent(content string) {
	m.content = content
	m.renderMarkdown()
}

// TogglePreview toggles the preview visibility
func (m *MarkdownPreviewModel) TogglePreview() {
	m.showPreview = !m.showPreview
}

// ShowPreview sets the preview visibility
func (m *MarkdownPreviewModel) ShowPreview(show bool) {
	m.showPreview = show
}

// IsShowing returns whether the preview is currently visible
func (m *MarkdownPreviewModel) IsShowing() bool {
	return m.showPreview
}

// renderMarkdown converts markdown content to terminal-friendly format
func (m *MarkdownPreviewModel) renderMarkdown() {
	if m.content == "" {
		m.rendered = ""
		return
	}

	// Simple markdown to terminal format conversion
	lines := strings.Split(m.content, "\n")
	var result []string

	for _, line := range lines {
		processedLine := m.processMarkdownLine(line)
		result = append(result, processedLine...)
	}

	m.rendered = strings.Join(result, "\n")
}

// processMarkdownLine processes a single line of markdown
func (m *MarkdownPreviewModel) processMarkdownLine(line string) []string {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return []string{""}
	}

	// Handle headings
	if strings.HasPrefix(trimmed, "#") {
		return m.processHeading(trimmed)
	}

	// Handle lists
	if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
		return []string{m.styleListItem(trimmed)}
	}

	// Handle code blocks
	if strings.HasPrefix(trimmed, "```") {
		return []string{m.styleCodeBlock(trimmed)}
	}

	// Handle blockquotes
	if strings.HasPrefix(trimmed, "> ") {
		return []string{m.styleBlockquote(trimmed)}
	}

	// Regular paragraph
	return []string{m.styleParagraph(trimmed)}
}

// processHeading processes heading lines
func (m *MarkdownPreviewModel) processHeading(line string) []string {
	level := 0
	for i, char := range line {
		if char == '#' {
			level = i + 1
		} else {
			break
		}
	}

	text := strings.TrimSpace(line[level:])

	var color string
	switch level {
	case 1:
		color = "#38BDF8" // Bright blue
	case 2:
		color = "#4ADE80" // Bright green
	case 3:
		color = "#F59E0B" // Bright yellow
	default:
		color = "#C084FC" // Bright purple
	}

	style := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(true)
	prefix := strings.Repeat("#", level) + " "
	return []string{style.Render(prefix + text)}
}

// styleListItem styles a list item
func (m *MarkdownPreviewModel) styleListItem(line string) string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("#94A3B8"))
	content := strings.TrimSpace(line[2:]) // Remove "- " or "* "
	return style.Render("• " + content)
}

// styleCodeBlock styles a code block
func (m *MarkdownPreviewModel) styleCodeBlock(line string) string {
	style := lipgloss.NewStyle().
		Background(lipgloss.Color("#1F2937")).
		Foreground(lipgloss.Color("#10B981")).
		Padding(0, 1)
	return style.Render(line)
}

// styleBlockquote styles a blockquote
func (m *MarkdownPreviewModel) styleBlockquote(line string) string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9CA3AF")).
		Italic(true)
	content := strings.TrimSpace(line[2:]) // Remove "> "
	return style.Render("│ " + content)
}

// styleParagraph styles regular text
func (m *MarkdownPreviewModel) styleParagraph(line string) string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("#F1F5F9"))
	return style.Render(line)
}

// Update handles updates for the markdown preview
func (m *MarkdownPreviewModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.renderMarkdown() // Re-render to adapt to new dimensions
	}
	return nil
}

// ScrollUp scrolls the preview content up
func (m *MarkdownPreviewModel) ScrollUp() {
	if m.scrollPos > 0 {
		m.scrollPos--
	}
}

// ScrollDown scrolls the preview content down
func (m *MarkdownPreviewModel) ScrollDown() {
	lines := strings.Split(m.rendered, "\n")
	if m.scrollPos < len(lines)-m.getMaxVisibleLines() {
		m.scrollPos++
	}
}

// ScrollToTop scrolls to the top of the preview
func (m *MarkdownPreviewModel) ScrollToTop() {
	m.scrollPos = 0
}

// ScrollToBottom scrolls to the bottom of the preview
func (m *MarkdownPreviewModel) ScrollToBottom() {
	lines := strings.Split(m.rendered, "\n")
	maxLines := m.getMaxVisibleLines()
	if len(lines) > maxLines {
		m.scrollPos = len(lines) - maxLines
	} else {
		m.scrollPos = 0
	}
}

// getMaxVisibleLines calculates how many lines can be displayed
func (m *MarkdownPreviewModel) getMaxVisibleLines() int {
	// Reserve space for title and borders
	return m.height - 6
}

// View renders the markdown preview
func (m *MarkdownPreviewModel) View() string {
	if !m.showPreview {
		return ""
	}

	title := previewTitleStyle.Render("Preview")

	if m.rendered == "" {
		return title + "\n" + previewStyle.Render("No content to preview")
	}

	// Split rendered content into lines and apply scrolling
	lines := strings.Split(m.rendered, "\n")
	maxLines := m.getMaxVisibleLines()

	if m.scrollPos < 0 {
		m.scrollPos = 0
	}
	if m.scrollPos > len(lines)-maxLines && len(lines) > maxLines {
		m.scrollPos = len(lines) - maxLines
	}

	// Get visible lines
	var visibleLines []string
	if len(lines) <= maxLines {
		visibleLines = lines
	} else {
		end := m.scrollPos + maxLines
		if end > len(lines) {
			end = len(lines)
		}
		visibleLines = lines[m.scrollPos:end]
	}

	content := strings.Join(visibleLines, "\n")
	renderedContent := previewContentStyle.Render(content)

	// Add scroll indicator if needed
	scrollIndicator := ""
	if len(lines) > maxLines {
		percentage := float64(m.scrollPos) / float64(len(lines)-maxLines) * 100
		scrollIndicator = lipgloss.NewStyle().
			Foreground(lipgloss.Color("62")).
			Render(fmt.Sprintf(" [%d%%] ", int(percentage)))
	}

	return title + "\n" + previewStyle.Render(renderedContent+scrollIndicator)
}