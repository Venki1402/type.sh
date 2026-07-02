package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"typesh/internal/cheat"
	"typesh/internal/typing"
)

// textWidth is the max width of the typing passage for readability.
const textWidth = 60

// View renders the current screen.
func (m Model) View() string {
	if m.quitting {
		name := m.cfg.Profile.Name
		if name == "" {
			name = "friend"
		}
		return m.styles.Subtle.Render(fmt.Sprintf("\n  Keep practicing, %s. 👋\n\n", name))
	}

	var body string
	switch m.state {
	case stateName:
		body = m.viewName()
	case stateMenu:
		body = m.viewMenu()
	case stateModeSelect:
		body = m.viewModeSelect()
	case stateCountdown:
		body = m.viewCountdown()
	case stateTyping:
		body = m.viewTyping()
	case stateResult:
		body = m.viewResult()
	case stateStats:
		body = m.viewStats()
	case stateSettings:
		body = m.viewSettings()
	}

	return m.center(body)
}

// center places the body in the middle of the terminal when sizes are known.
func (m Model) center(body string) string {
	if m.width <= 0 || m.height <= 0 {
		return body
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, body)
}

func (m Model) logo() string {
	return m.styles.Title.Render("type.sh") + m.styles.Subtle.Render("  ·  terminal typing trainer")
}

// --- onboarding ---

func (m Model) viewName() string {
	b := &strings.Builder{}
	fmt.Fprintln(b, m.logo())
	fmt.Fprintln(b)
	fmt.Fprintln(b, m.styles.Text.Render("Welcome! What should we call you?"))
	fmt.Fprintln(b)
	fmt.Fprintln(b, m.input.View())
	fmt.Fprintln(b)
	fmt.Fprint(b, m.styles.Subtle.Render("enter confirm · esc skip"))
	return m.styles.Box.Render(b.String())
}

// --- menu ---

func (m Model) viewMenu() string {
	b := &strings.Builder{}
	fmt.Fprintln(b, m.logo())
	fmt.Fprintln(b)
	fmt.Fprintln(b, m.profileLine())
	fmt.Fprintln(b, m.xpLine())
	fmt.Fprintln(b)

	for i, item := range menuItems {
		cursor := "  "
		render := m.styles.Item
		if i == m.menuIndex {
			cursor = m.styles.Accent.Render("› ")
			render = m.styles.Selected
		}
		fmt.Fprintf(b, "%s%s\n", cursor, render.Render(item))
	}
	fmt.Fprintln(b)
	fmt.Fprint(b, m.styles.Subtle.Render("↑↓ move · enter select · q quit"))
	return m.styles.Box.Render(b.String())
}

func (m Model) profileLine() string {
	p := m.cfg.Profile
	trust := m.trustBadge()
	return fmt.Sprintf("%s  %s  %s  %s",
		m.styles.Accent.Render(orDefault(p.Name, "Anonymous")),
		m.styles.Subtle.Render(fmt.Sprintf("Lv %d", p.Level())),
		m.styles.Text.Render(fmt.Sprintf("best %.0f wpm", p.BestWPM)),
		trust,
	)
}

func (m Model) xpLine() string {
	p := m.cfg.Profile
	bar := m.styles.progressBar(float64(p.XPInLevel())/100.0, 24)
	return fmt.Sprintf("%s %s", bar, m.styles.Subtle.Render(fmt.Sprintf("%d/100 xp", p.XPInLevel())))
}

func (m Model) trustBadge() string {
	p := m.cfg.Profile
	if p.TotalTests == 0 {
		return m.styles.Subtle.Render("new")
	}
	pct := float64(p.SuspiciousTests) / float64(p.TotalTests) * 100
	switch {
	case pct == 0:
		return m.styles.Good.Render("clean")
	case pct < 20:
		return m.styles.Warn.Render("caution")
	default:
		return m.styles.Bad.Render("flagged")
	}
}

// --- mode selection ---

func (m Model) viewModeSelect() string {
	b := &strings.Builder{}
	title := "Time test"
	unit := "seconds"
	if m.modeType == "word" {
		title = "Word test"
		unit = "words"
	}
	fmt.Fprintln(b, m.styles.Title.Render(title))
	fmt.Fprintln(b)

	opts := m.currentOptions()
	cells := make([]string, len(opts))
	for i, v := range opts {
		label := fmt.Sprintf(" %d ", v)
		if i == m.modeIndex {
			cells[i] = m.styles.Cursor.Render(label)
		} else {
			cells[i] = m.styles.Pending.Render(label)
		}
	}
	fmt.Fprintln(b, strings.Join(cells, "  "))
	fmt.Fprintln(b)
	fmt.Fprintln(b, m.styles.Subtle.Render(fmt.Sprintf("test length: %s", unit)))
	fmt.Fprintln(b)
	fmt.Fprint(b, m.styles.Subtle.Render("←→ choose · enter start · esc back"))
	return m.styles.Box.Render(b.String())
}

// --- countdown ---

func (m Model) viewCountdown() string {
	remaining := 3 - int(time.Since(m.countdownStart).Seconds())
	if remaining < 1 {
		remaining = 1
	}
	big := m.styles.Title.Render(fmt.Sprintf("%d", remaining))
	return m.styles.Box.Render(
		m.styles.Subtle.Render("get ready") + "\n\n   " + big + "   \n",
	)
}

// --- typing ---

func (m Model) viewTyping() string {
	st := m.session.Stats()

	// Live header: timer + wpm + accuracy.
	var timer string
	if m.spec.kind == "time" {
		remaining := m.testDuration - m.session.Elapsed()
		if remaining < 0 {
			remaining = 0
		}
		timer = fmt.Sprintf("%2.0fs", remaining.Seconds())
	} else {
		timer = fmt.Sprintf("%2.0fs", m.session.Elapsed().Seconds())
	}

	header := fmt.Sprintf("%s   %s   %s",
		m.styles.Accent.Render(timer),
		m.styles.Text.Render(fmt.Sprintf("%.0f wpm", st.WPM)),
		m.styles.Subtle.Render(fmt.Sprintf("%.0f%% acc", liveAcc(st))),
	)

	passage := m.renderPassage()

	b := &strings.Builder{}
	fmt.Fprintln(b, header)
	fmt.Fprintln(b)
	fmt.Fprintln(b, passage)
	fmt.Fprintln(b)
	fmt.Fprint(b, m.styles.Subtle.Render("esc cancel"))
	return m.styles.Box.Width(textWidth + 6).Render(b.String())
}

// renderPassage draws the target text with per-character coloring, a cursor,
// and word-wrapping to textWidth.
func (m Model) renderPassage() string {
	target := m.session.Target
	input := m.session.Input
	pos := len(input)
	breaks := wrapBreaks(target, textWidth)

	var b strings.Builder
	for i, r := range target {
		if breaks[i] {
			b.WriteByte('\n')
			continue
		}
		ch := string(r)
		switch {
		case i == pos:
			b.WriteString(m.styles.Cursor.Render(ch))
		case i < pos:
			if input[i] == r {
				b.WriteString(m.styles.Correct.Render(ch))
			} else {
				// Show the expected char, underlined in the error color.
				b.WriteString(m.styles.Incorrect.Render(ch))
			}
		default:
			b.WriteString(m.styles.Pending.Render(ch))
		}
	}

	// Characters typed past the end of the passage.
	if pos > len(target) {
		for _, r := range input[len(target):] {
			b.WriteString(m.styles.Incorrect.Render(string(r)))
		}
	}
	return b.String()
}

// wrapBreaks returns the set of target indices (spaces) to replace with a
// newline so that no line exceeds width columns.
func wrapBreaks(target []rune, width int) map[int]bool {
	breaks := make(map[int]bool)
	lineStart := 0
	lastSpace := -1
	for i, r := range target {
		if r == ' ' {
			lastSpace = i
		}
		if i-lineStart >= width && lastSpace > lineStart {
			breaks[lastSpace] = true
			lineStart = lastSpace + 1
			lastSpace = -1
		}
	}
	return breaks
}

// --- result ---

func (m Model) viewResult() string {
	st := m.res.stats
	rep := m.res.report
	b := &strings.Builder{}

	if rep.Suspicious {
		fmt.Fprintln(b, m.styles.Bad.Render("⚠  test flagged"))
	} else {
		fmt.Fprintln(b, m.styles.Good.Render("✓  test complete"))
	}
	fmt.Fprintln(b)

	fmt.Fprintf(b, "%s  %s\n",
		m.styles.Title.Render(fmt.Sprintf("%.0f", st.WPM)),
		m.styles.Subtle.Render("wpm"),
	)
	fmt.Fprintf(b, "%s   %s   %s\n",
		m.styles.Text.Render(fmt.Sprintf("%.0f%% acc", st.Accuracy)),
		m.styles.Text.Render(fmt.Sprintf("%.0f%% consistency", st.Consistency)),
		m.styles.Subtle.Render(fmt.Sprintf("%d errors", st.Errors)),
	)
	fmt.Fprintf(b, "%s\n",
		m.styles.Subtle.Render(fmt.Sprintf("raw %.0f · %.1fs · %d keys",
			st.RawWPM, st.Elapsed.Seconds(), st.TotalKeys)),
	)

	if rep.Suspicious {
		fmt.Fprintln(b)
		fmt.Fprintln(b, m.styles.Bad.Render("anti-cheat flags:"))
		for _, f := range rep.Flags {
			fmt.Fprintf(b, "  %s %s — %s\n",
				m.severityTag(f.Severity),
				m.styles.Text.Render(f.Title),
				m.styles.Subtle.Render(f.Detail),
			)
		}
		fmt.Fprintln(b, m.styles.Subtle.Render("  no xp or records awarded"))
	} else {
		fmt.Fprintln(b)
		xpLine := m.styles.Good.Render(fmt.Sprintf("+%d xp", m.res.xp))
		if m.res.newBest {
			xpLine += "   " + m.styles.Accent.Render("★ new best!")
		}
		if m.res.leveledUp {
			xpLine += "   " + m.styles.Accent.Render(fmt.Sprintf("⇧ level %d!", m.cfg.Profile.Level()))
		}
		fmt.Fprintln(b, xpLine)
	}

	if m.res.saveErr != nil {
		fmt.Fprintln(b, m.styles.Warn.Render("  (could not save progress)"))
	}

	fmt.Fprintln(b)
	fmt.Fprint(b, m.styles.Subtle.Render("r retry · any key menu"))
	return m.styles.Box.Render(b.String())
}

func (m Model) severityTag(s cheat.Severity) string {
	switch s {
	case cheat.High:
		return m.styles.Bad.Render("[HIGH]")
	case cheat.Medium:
		return m.styles.Warn.Render("[MED ]")
	default:
		return m.styles.Subtle.Render("[LOW ]")
	}
}

// --- stats ---

func (m Model) viewStats() string {
	p := m.cfg.Profile
	b := &strings.Builder{}
	fmt.Fprintln(b, m.styles.Title.Render("your stats"))
	fmt.Fprintln(b)
	fmt.Fprintf(b, "%s %s\n", pad("player"), m.styles.Text.Render(orDefault(p.Name, "Anonymous")))
	fmt.Fprintf(b, "%s %s\n", pad("level"), m.styles.Text.Render(fmt.Sprintf("%d  (%d xp)", p.Level(), p.XP)))
	fmt.Fprintf(b, "%s %s\n", pad("best wpm"), m.styles.Text.Render(fmt.Sprintf("%.1f", p.BestWPM)))
	fmt.Fprintf(b, "%s %s\n", pad("best acc"), m.styles.Text.Render(fmt.Sprintf("%.1f%%", p.BestAccuracy)))
	fmt.Fprintf(b, "%s %s\n", pad("tests"), m.styles.Text.Render(fmt.Sprintf("%d", p.TotalTests)))
	fmt.Fprintf(b, "%s %s / %s\n", pad("clean/flag"),
		m.styles.Good.Render(fmt.Sprintf("%d", p.CleanTests)),
		m.styles.Bad.Render(fmt.Sprintf("%d", p.SuspiciousTests)))

	// Recent history sparkline-ish list.
	if n := len(m.cfg.History); n > 0 {
		fmt.Fprintln(b)
		fmt.Fprintln(b, m.styles.Subtle.Render("recent:"))
		start := n - 5
		if start < 0 {
			start = 0
		}
		for _, h := range m.cfg.History[start:] {
			mark := m.styles.Good.Render("·")
			if h.Suspicious {
				mark = m.styles.Bad.Render("✗")
			}
			fmt.Fprintf(b, "  %s %s\n", mark,
				m.styles.Text.Render(fmt.Sprintf("%.0f wpm  %.0f%%  %s", h.WPM, h.Accuracy, h.Mode)))
		}
	}

	fmt.Fprintln(b)
	fmt.Fprint(b, m.styles.Subtle.Render(fmt.Sprintf("config: %s", m.cfg.Path())))
	fmt.Fprintln(b)
	fmt.Fprint(b, m.styles.Subtle.Render("any key back"))
	return m.styles.Box.Render(b.String())
}

// --- settings ---

func (m Model) viewSettings() string {
	b := &strings.Builder{}
	fmt.Fprintln(b, m.styles.Title.Render("settings"))
	fmt.Fprintln(b)

	rows := []struct {
		label string
		value string
	}{
		{"name", m.nameField()},
		{"theme", "‹ " + m.cfg.Preferences.Theme + " ›"},
		{"punctuation", onOff(m.cfg.Preferences.IncludePunct)},
		{"numbers", onOff(m.cfg.Preferences.IncludeNumbers)},
		{"", "save & back"},
	}

	for i, row := range rows {
		cursor := "  "
		labelStyle := m.styles.Subtle
		valStyle := m.styles.Text
		if i == m.settingsIndex {
			cursor = m.styles.Accent.Render("› ")
			valStyle = m.styles.Selected
		}
		if row.label == "" {
			fmt.Fprintf(b, "%s%s\n", cursor, valStyle.Render(row.value))
		} else {
			fmt.Fprintf(b, "%s%s %s\n", cursor, labelStyle.Render(pad(row.label)), valStyle.Render(row.value))
		}
	}

	fmt.Fprintln(b)
	if m.settingsIndex == setName && m.input.Focused() {
		fmt.Fprint(b, m.styles.Subtle.Render("type name · enter save · esc cancel"))
	} else {
		fmt.Fprint(b, m.styles.Subtle.Render("↑↓ move · ←→/enter change · esc back"))
	}
	return m.styles.Box.Render(b.String())
}

func (m Model) nameField() string {
	if m.settingsIndex == setName && m.input.Focused() {
		return m.input.View()
	}
	return orDefault(m.cfg.Profile.Name, "Anonymous")
}

// --- small helpers ---

func pad(s string) string { return fmt.Sprintf("%-11s", s) }

func onOff(b bool) string {
	if b {
		return "on"
	}
	return "off"
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

// liveAcc returns accuracy, defaulting to 100% before any keystroke is made.
func liveAcc(st typing.Stats) float64 {
	if st.TotalKeys == 0 {
		return 100
	}
	return st.Accuracy
}
