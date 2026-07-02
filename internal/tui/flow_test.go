package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"typesh/internal/config"
)

func key(s string) tea.KeyMsg {
	switch s {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case " ":
		return tea.KeyMsg{Type: tea.KeySpace}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func send(m tea.Model, msg tea.Msg) Model {
	next, _ := m.Update(msg)
	return next.(Model)
}

// TestFullWordTestFlow drives menu → word test → typing → result and asserts
// that progress is recorded and persisted.
func TestFullWordTestFlow(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg, _, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	cfg.Profile.Name = "Tester"
	cfg.Preferences.DefaultWords = 10

	m := New(cfg, false)
	m = send(m, tea.WindowSizeMsg{Width: 100, Height: 40})

	// Menu: move to "Word test" (index 1) and select.
	m = send(m, key("down"))
	m = send(m, key("enter"))
	if m.state != stateModeSelect || m.modeType != "word" {
		t.Fatalf("expected word mode select, got state=%d type=%s", m.state, m.modeType)
	}

	// Start the test → countdown.
	m = send(m, key("enter"))
	if m.state != stateCountdown {
		t.Fatalf("expected countdown, got %d", m.state)
	}

	// Fast-forward the countdown.
	m.countdownStart = time.Now().Add(-4 * time.Second)
	m = send(m, tickMsg(time.Now()))
	if m.state != stateTyping {
		t.Fatalf("expected typing, got %d", m.state)
	}

	// Type the whole passage correctly, with a small delay for realism.
	target := string(m.session.Target)
	for _, r := range target {
		if r == ' ' {
			m = send(m, key(" "))
		} else {
			m = send(m, key(string(r)))
		}
		time.Sleep(2 * time.Millisecond)
	}

	if m.state != stateResult {
		t.Fatalf("expected result after completing passage, got %d", m.state)
	}
	if cfg.Profile.TotalTests != 1 {
		t.Fatalf("expected 1 test recorded, got %d", cfg.Profile.TotalTests)
	}
	if m.res.stats.Accuracy < 99 {
		t.Fatalf("expected near-perfect accuracy, got %.1f", m.res.stats.Accuracy)
	}
	if len(cfg.History) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(cfg.History))
	}

	// Verify persistence: reload from disk.
	reloaded, _, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.Profile.TotalTests != 1 {
		t.Fatalf("persisted total tests = %d, want 1", reloaded.Profile.TotalTests)
	}
	if reloaded.Profile.Name != "Tester" {
		t.Fatalf("persisted name = %q, want Tester", reloaded.Profile.Name)
	}
}

// TestConfigRoundTrip checks the first-run detection and save/load cycle.
func TestConfigRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg, firstRun, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if !firstRun {
		t.Fatal("expected first run on empty config dir")
	}
	cfg.Profile.Name = "Neo"
	cfg.Preferences.Theme = "warm"
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}

	again, firstRun, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if firstRun {
		t.Fatal("expected not-first-run after save")
	}
	if again.Profile.Name != "Neo" || again.Preferences.Theme != "warm" {
		t.Fatalf("roundtrip mismatch: %+v", again.Profile)
	}
}
