package theme

import (
	"math"
)

// Breakpoint defines terminal size breakpoints
type Breakpoint int

const (
	BreakpointSmall Breakpoint = iota  // < 100 width
	BreakpointMedium                   // 100-140 width
	BreakpointLarge                    // > 140 width
)

// Responsive utilities
type Responsive struct {
	Width  int
	Height int
}

// NewResponsive creates a new responsive helper
func NewResponsive(width, height int) *Responsive {
	return &Responsive{
		Width:  width,
		Height: height,
	}
}

// GetBreakpoint returns the current breakpoint based on width
func (r *Responsive) GetBreakpoint() Breakpoint {
	if r.Width < 100 {
		return BreakpointSmall
	} else if r.Width < 140 {
		return BreakpointMedium
	}
	return BreakpointLarge
}

// IsSmall returns true for small terminals
func (r *Responsive) IsSmall() bool {
	return r.Width < 100
}

// IsMedium returns true for medium terminals
func (r *Responsive) IsMedium() bool {
	return r.Width >= 100 && r.Width < 140
}

// IsLarge returns true for large terminals
func (r *Responsive) IsLarge() bool {
	return r.Width >= 140
}

// WidthPercent returns width as percentage of terminal width
func (r *Responsive) WidthPercent(percent int) int {
	return int(float64(r.Width) * float64(percent) / 100.0)
}

// HeightPercent returns height as percentage of terminal height
func (r *Responsive) HeightPercent(percent int) int {
	return int(float64(r.Height) * float64(percent) / 100.0)
}

// MaxWidth returns the maximum of two values
func (r *Responsive) MaxWidth(a, b int) int {
	return int(math.Max(float64(a), float64(b)))
}

// MinWidth returns the minimum of two values
func (r *Responsive) MinWidth(a, b int) int {
	return int(math.Min(float64(a), float64(b)))
}

// ClampWidth constrains width between min and max values
func (r *Responsive) ClampWidth(width, min, max int) int {
	return int(math.Max(float64(min), math.Min(float64(width), float64(max))))
}

// Responsive width calculations for different components
func (r *Responsive) SearchWidth() int {
	switch r.GetBreakpoint() {
	case BreakpointSmall:
		return r.ClampWidth(r.Width-10, 40, 60)
	case BreakpointMedium:
		return r.ClampWidth(r.WidthPercent(70), 50, 80)
	default: // BreakpointLarge
		return r.ClampWidth(r.WidthPercent(60), 60, 100)
	}
}

func (r *Responsive) EditorWidth() int {
	switch r.GetBreakpoint() {
	case BreakpointSmall:
		return r.Width - 6
	case BreakpointMedium:
		return r.Width - 8
	default:
		return r.WidthPercent(90)
	}
}

func (r *Responsive) SplitPaneEditorWidth() int {
	return (r.Width - 8) / 2 // Account for borders and spacing
}

func (r *Responsive) SplitPanePreviewWidth() int {
	editorWidth := r.SplitPaneEditorWidth()
	return r.Width - editorWidth - 4 // Leave space for borders
}

func (r *Responsive) ContentHeight(usedHeight int) int {
	available := r.Height - usedHeight - 4 // Reserve space for controls
	return r.MaxWidth(available, 5) // Minimum height of 5
}

func (r *Responsive) TagInputWidth() int {
	switch r.GetBreakpoint() {
	case BreakpointSmall:
		return r.ClampWidth(r.Width-20, 20, 30)
	case BreakpointMedium:
		return 40
	default:
		return 50
	}
}

// Spacing utilities
func (r *Responsive) Padding() int {
	switch r.GetBreakpoint() {
	case BreakpointSmall:
		return 0
	case BreakpointMedium:
		return 1
	default:
		return 1
	}
}

func (r *Responsive) Margin() int {
	switch r.GetBreakpoint() {
	case BreakpointSmall:
		return 0
	case BreakpointMedium:
		return 1
	default:
		return 1
	}
}

// Truncate text with ellipsis if it exceeds maxLength
func TruncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	if maxLength < 3 {
		return text[:maxLength]
	}
	return text[:maxLength-3] + "..."
}

// Calculate maximum title length for list items
func (r *Responsive) MaxTitleLength() int {
	switch r.GetBreakpoint() {
	case BreakpointSmall:
		return r.ClampWidth(r.Width-8, 20, 40)
	case BreakpointMedium:
		return r.ClampWidth(r.Width-10, 30, 60)
	default:
		return r.ClampWidth(r.Width-12, 40, 80)
	}
}