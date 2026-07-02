package tui

import "github.com/charmbracelet/lipgloss"

// Theme is a named color palette for the interface.
type Theme struct {
	Name      string
	Correct   lipgloss.Color // correctly typed characters
	Incorrect lipgloss.Color // mistyped characters
	Pending   lipgloss.Color // not-yet-typed characters
	Cursor    lipgloss.Color // caret background
	Accent    lipgloss.Color // headings / highlights
	Text      lipgloss.Color // primary text
	Dim       lipgloss.Color // secondary text
	Good      lipgloss.Color // positive status
	Warn      lipgloss.Color // caution status
	Bad       lipgloss.Color // negative status
}

// ThemeOrder is the cycle order for the settings screen.
var ThemeOrder = []string{"default", "warm", "mono"}

// themes holds the available palettes.
var themes = map[string]Theme{
	"default": {
		Name:      "default",
		Correct:   lipgloss.Color("#a6e3a1"),
		Incorrect: lipgloss.Color("#f38ba8"),
		Pending:   lipgloss.Color("#585b70"),
		Cursor:    lipgloss.Color("#f9e2af"),
		Accent:    lipgloss.Color("#89b4fa"),
		Text:      lipgloss.Color("#cdd6f4"),
		Dim:       lipgloss.Color("#7f849c"),
		Good:      lipgloss.Color("#a6e3a1"),
		Warn:      lipgloss.Color("#f9e2af"),
		Bad:       lipgloss.Color("#f38ba8"),
	},
	"warm": {
		Name:      "warm",
		Correct:   lipgloss.Color("#e0af68"),
		Incorrect: lipgloss.Color("#f7768e"),
		Pending:   lipgloss.Color("#565f89"),
		Cursor:    lipgloss.Color("#ff9e64"),
		Accent:    lipgloss.Color("#ff9e64"),
		Text:      lipgloss.Color("#c0caf5"),
		Dim:       lipgloss.Color("#787c99"),
		Good:      lipgloss.Color("#9ece6a"),
		Warn:      lipgloss.Color("#e0af68"),
		Bad:       lipgloss.Color("#f7768e"),
	},
	"mono": {
		Name:      "mono",
		Correct:   lipgloss.Color("#ffffff"),
		Incorrect: lipgloss.Color("#ff5f5f"),
		Pending:   lipgloss.Color("#5f5f5f"),
		Cursor:    lipgloss.Color("#ffffff"),
		Accent:    lipgloss.Color("#bcbcbc"),
		Text:      lipgloss.Color("#d0d0d0"),
		Dim:       lipgloss.Color("#808080"),
		Good:      lipgloss.Color("#d0d0d0"),
		Warn:      lipgloss.Color("#bcbcbc"),
		Bad:       lipgloss.Color("#ff5f5f"),
	},
}

// themeByName returns the named theme, falling back to default.
func themeByName(name string) Theme {
	if t, ok := themes[name]; ok {
		return t
	}
	return themes["default"]
}

// Styles bundles the reusable lipgloss styles derived from a Theme.
type Styles struct {
	theme Theme

	Title    lipgloss.Style
	Subtle   lipgloss.Style
	Accent   lipgloss.Style
	Text     lipgloss.Style
	Box      lipgloss.Style
	Selected lipgloss.Style
	Item     lipgloss.Style

	Correct   lipgloss.Style
	Incorrect lipgloss.Style
	Pending   lipgloss.Style
	Cursor    lipgloss.Style

	Good lipgloss.Style
	Warn lipgloss.Style
	Bad  lipgloss.Style
}

// newStyles builds a Styles set from a theme.
func newStyles(t Theme) Styles {
	return Styles{
		theme:     t,
		Title:     lipgloss.NewStyle().Foreground(t.Accent).Bold(true),
		Subtle:    lipgloss.NewStyle().Foreground(t.Dim),
		Accent:    lipgloss.NewStyle().Foreground(t.Accent),
		Text:      lipgloss.NewStyle().Foreground(t.Text),
		Box:       lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(t.Dim).Padding(1, 3),
		Selected:  lipgloss.NewStyle().Foreground(t.Accent).Bold(true),
		Item:      lipgloss.NewStyle().Foreground(t.Text),
		Correct:   lipgloss.NewStyle().Foreground(t.Correct),
		Incorrect: lipgloss.NewStyle().Foreground(t.Incorrect).Underline(true),
		Pending:   lipgloss.NewStyle().Foreground(t.Pending),
		Cursor:    lipgloss.NewStyle().Foreground(lipgloss.Color("#1e1e2e")).Background(t.Cursor),
		Good:      lipgloss.NewStyle().Foreground(t.Good).Bold(true),
		Warn:      lipgloss.NewStyle().Foreground(t.Warn).Bold(true),
		Bad:       lipgloss.NewStyle().Foreground(t.Bad).Bold(true),
	}
}

// progressBar renders a simple [████░░░░] style bar of the given width.
func (s Styles) progressBar(ratio float64, width int) string {
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	filled := int(ratio * float64(width))
	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += s.Accent.Render("█")
		} else {
			bar += s.Pending.Render("░")
		}
	}
	return bar
}
