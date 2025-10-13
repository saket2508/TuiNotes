package theme

import (
	"github.com/charmbracelet/lipgloss"
)

// Color defines the unified color palette for the application
type Color struct {
	Background    lipgloss.Color
	Primary       lipgloss.Color
	Secondary     lipgloss.Color
	Accent        lipgloss.Color
	Text          lipgloss.Color
	Muted         lipgloss.Color
	Subtle        lipgloss.Color
	Success       lipgloss.Color
	Warning       lipgloss.Color
	Error         lipgloss.Color
	Border        lipgloss.Color
	BorderActive  lipgloss.Color
	BorderInactive lipgloss.Color
}

// Colors contains the unified color scheme
var Colors = Color{
	Background:     lipgloss.Color("#0F172A"), // Deep slate background
	Primary:        lipgloss.Color("#38BDF8"), // Bright cyan for primary actions
	Secondary:      lipgloss.Color("#10B981"), // Emerald green for secondary elements
	Accent:         lipgloss.Color("#F59E0B"), // Amber for highlights
	Text:           lipgloss.Color("#F1F5F9"), // Light slate for primary text
	Muted:          lipgloss.Color("#94A3B8"), // Slate gray for secondary text
	Subtle:         lipgloss.Color("#64748B"), // Muted slate for subtle elements
	Success:        lipgloss.Color("#22C55E"), // Green for success states
	Warning:        lipgloss.Color("#F59E0B"), // Amber for warnings
	Error:          lipgloss.Color("#F43F5E"), // Rose for error states
	Border:         lipgloss.Color("#334155"), // Border slate
	BorderActive:   lipgloss.Color("#38BDF8"), // Cyan for active borders
	BorderInactive: lipgloss.Color("#475569"), // Dimmer border for inactive elements
}

// Tag colors for variety and visual hierarchy
var TagColors = []struct {
	Foreground lipgloss.Color
	Background lipgloss.Color
	Border     lipgloss.Color
}{
	{Colors.Primary, Colors.Background, lipgloss.Color("#0EA5E9")},     // Cyan
	{Colors.Secondary, Colors.Background, lipgloss.Color("#22C55E")},   // Green
	{lipgloss.Color("#C084FC"), Colors.Background, lipgloss.Color("#A855F7")}, // Purple
	{lipgloss.Color("#FB923C"), Colors.Background, lipgloss.Color("#F97316")}, // Orange
}

// Heading colors for markdown preview
var HeadingColors = []lipgloss.Color{
	Colors.Primary,        // H1 - Cyan
	Colors.Secondary,      // H2 - Green
	Colors.Accent,         // H3 - Amber
	lipgloss.Color("#C084FC"), // H4+ - Purple
}