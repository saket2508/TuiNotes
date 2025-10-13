package theme

import (
	"github.com/charmbracelet/lipgloss"
)

// Styles contains all the pre-defined styles for the application
type Styles struct {
	// Title styles
	TitleBar       lipgloss.Style
	SecondaryTitle lipgloss.Style

	// Input field styles
	ActiveField    lipgloss.Style
	InactiveField  lipgloss.Style
	SearchActive   lipgloss.Style
	SearchInactive lipgloss.Style

	// Content styles
	Content        lipgloss.Style
	ContentActive  lipgloss.Style
	ContentBox     lipgloss.Style

	// Tag styles
	Tag            []lipgloss.Style
	TagInput       lipgloss.Style
	TagInputActive lipgloss.Style
	TagLabel       lipgloss.Style

	// List styles
	ListItem       lipgloss.Style
	ListItemSelected lipgloss.Style
	ListCursor     lipgloss.Style

	// Button/Control styles
	ControlText    lipgloss.Style
	KeyBinding     lipgloss.Style
	Description    lipgloss.Style

	// Preview styles
	PreviewTitle   lipgloss.Style
	PreviewBox     lipgloss.Style
	PreviewContent lipgloss.Style

	// Pane styles for split view
	EditorPane     lipgloss.Style
	PreviewPane    lipgloss.Style

	// Message/Status styles
	SuccessText    lipgloss.Style
	ErrorText      lipgloss.Style
	WarningText    lipgloss.Style

	// Border styles
	BorderActive   lipgloss.Style
	BorderInactive lipgloss.Style
}

// NewStyles creates the complete style system
func NewStyles() *Styles {
	styles := &Styles{}

	// Title styles
	styles.TitleBar = lipgloss.NewStyle().
		Foreground(Colors.Text).
		Background(Colors.Primary).
		Bold(true).
		Padding(0, 1).
		MarginBottom(1)

	styles.SecondaryTitle = lipgloss.NewStyle().
		Foreground(Colors.Primary).
		Bold(true).
		MarginBottom(1)

	// Input field styles
	styles.ActiveField = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Colors.BorderActive).
		Foreground(Colors.Text).
		Padding(0, 1)

	styles.InactiveField = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(Colors.BorderInactive).
		Foreground(Colors.Muted).
		Padding(0, 1)

	styles.SearchActive = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Colors.Primary).
		Foreground(Colors.Text).
		Padding(0, 1)

	styles.SearchInactive = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(Colors.BorderInactive).
		Foreground(Colors.Muted).
		Padding(0, 1)

	// Content styles
	styles.Content = lipgloss.NewStyle().
		Foreground(Colors.Text).
		Padding(1)

	styles.ContentActive = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Colors.BorderActive).
		Foreground(Colors.Text).
		Padding(1)

	styles.ContentBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Colors.Border).
		Foreground(Colors.Text).
		Padding(1)

	// Tag styles
	styles.Tag = make([]lipgloss.Style, len(TagColors))
	for i, color := range TagColors {
		styles.Tag[i] = lipgloss.NewStyle().
			Foreground(color.Foreground).
			Background(color.Background).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(color.Border).
			Padding(0, 1).
			MarginRight(1)
	}

	styles.TagInput = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Colors.BorderInactive).
		Foreground(Colors.Muted).
		Padding(0, 1)

	styles.TagInputActive = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Colors.Primary).
		Foreground(Colors.Text).
		Padding(0, 1)

	styles.TagLabel = lipgloss.NewStyle().
		Foreground(Colors.Muted).
		Bold(true).
		MarginBottom(1)

	// List styles
	styles.ListItem = lipgloss.NewStyle().
		Foreground(Colors.Text).
		Padding(0, 1)

	styles.ListItemSelected = lipgloss.NewStyle().
		Foreground(Colors.Text).
		Background(lipgloss.Color("#1E293B")). // Subtle highlight
		Padding(0, 1)

	styles.ListCursor = lipgloss.NewStyle().
		Foreground(Colors.Primary).
		Bold(true)

	// Control styles
	styles.ControlText = lipgloss.NewStyle().
		Foreground(Colors.Muted).
		MarginTop(1)

	styles.KeyBinding = lipgloss.NewStyle().
		Foreground(Colors.Primary).
		Bold(true)

	styles.Description = lipgloss.NewStyle().
		Foreground(Colors.Muted)

	// Preview styles
	styles.PreviewTitle = lipgloss.NewStyle().
		Foreground(Colors.Text).
		Background(Colors.Primary).
		Bold(true).
		Padding(0, 1).
		MarginBottom(1)

	styles.PreviewBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Colors.Border).
		Padding(1)

	styles.PreviewContent = lipgloss.NewStyle().
		Foreground(Colors.Text)

	// Split pane styles
	styles.EditorPane = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Colors.Border).
		Padding(1)

	styles.PreviewPane = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Colors.Primary).
		Padding(1)

	// Message styles
	styles.SuccessText = lipgloss.NewStyle().
		Foreground(Colors.Success).
		Bold(true)

	styles.ErrorText = lipgloss.NewStyle().
		Foreground(Colors.Error).
		Bold(true)

	styles.WarningText = lipgloss.NewStyle().
		Foreground(Colors.Warning).
		Bold(true)

	// Border styles
	styles.BorderActive = lipgloss.NewStyle().
		BorderForeground(Colors.BorderActive)

	styles.BorderInactive = lipgloss.NewStyle().
		BorderForeground(Colors.BorderInactive)

	return styles
}