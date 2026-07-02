// Package tui implements the Bubble Tea interface for the typing test.
package tui

import (
	"math/rand"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"typesh/internal/cheat"
	"typesh/internal/config"
	"typesh/internal/typing"
	"typesh/internal/words"
)

type state int

const (
	stateName state = iota
	stateMenu
	stateModeSelect
	stateCountdown
	stateTyping
	stateResult
	stateStats
	stateSettings
)

// tickMsg drives live updates during the countdown and the typing test.
type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// testSpec describes the test the user is about to run.
type testSpec struct {
	kind    string // "time" | "word"
	seconds int
	words   int
}

// menuItems are the top-level choices.
var menuItems = []string{"Time test", "Word test", "Stats", "Settings", "Quit"}

// mode option values shared by time and word selection.
var timeOptions = []int{15, 30, 45, 60}
var wordOptions = []int{10, 25, 50, 100}

// settings rows.
const (
	setName = iota
	setTheme
	setPunct
	setNumbers
	setBack
	settingsRows
)

// Model is the root Bubble Tea model.
type Model struct {
	cfg    *config.Config
	styles Styles
	state  state

	width, height int
	rng           *rand.Rand

	// menu
	menuIndex int

	// mode selection
	modeType  string // "time" | "word"
	modeIndex int

	// typing test
	spec           testSpec
	session        *typing.Session
	testDuration   time.Duration
	countdownStart time.Time

	// text input (onboarding + settings name)
	input textinput.Model

	// settings
	settingsIndex int

	// result snapshot
	res struct {
		stats     typing.Stats
		report    cheat.Report
		xp        int
		newBest   bool
		leveledUp bool
		saveErr   error
	}

	quitting bool
}

// New builds the root model from a loaded config. firstRun forces onboarding.
func New(cfg *config.Config, firstRun bool) Model {
	ti := textinput.New()
	ti.Placeholder = "your name"
	ti.CharLimit = 24
	ti.Prompt = "› "
	ti.SetValue(cfg.Profile.Name)

	m := Model{
		cfg:    cfg,
		styles: newStyles(themeByName(cfg.Preferences.Theme)),
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
		input:  ti,
	}

	if firstRun {
		m.state = stateName
		m.input.SetValue("")
		m.input.Focus()
	} else {
		m.state = stateMenu
	}
	return m
}

func (m Model) Init() tea.Cmd {
	if m.state == stateName {
		return textinput.Blink
	}
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case tickMsg:
		return m.onTick()

	case tea.KeyMsg:
		// Global quit shortcut.
		if msg.Type == tea.KeyCtrlC {
			m.quitting = true
			return m, tea.Quit
		}
		switch m.state {
		case stateName:
			return m.updateName(msg)
		case stateMenu:
			return m.updateMenu(msg)
		case stateModeSelect:
			return m.updateModeSelect(msg)
		case stateTyping:
			return m.updateTyping(msg)
		case stateResult:
			return m.updateResult(msg)
		case stateStats:
			m.state = stateMenu
			return m, nil
		case stateSettings:
			return m.updateSettings(msg)
		}
	}
	return m, nil
}

// onTick advances the countdown and time-limited tests.
func (m Model) onTick() (tea.Model, tea.Cmd) {
	switch m.state {
	case stateCountdown:
		if time.Since(m.countdownStart) >= 3*time.Second {
			m.state = stateTyping
		}
		return m, tick()
	case stateTyping:
		if m.spec.kind == "time" && m.session.Started() &&
			m.session.Elapsed() >= m.testDuration {
			return m.finishTest()
		}
		return m, tick()
	}
	return m, nil
}

// --- onboarding ---

func (m Model) updateName(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		name := trimmed(m.input.Value())
		if name == "" {
			return m, nil
		}
		m.cfg.Profile.Name = name
		_ = m.cfg.Save()
		m.input.Blur()
		m.state = stateMenu
		return m, nil
	case tea.KeyEsc:
		// Allow skipping with a default name on first run.
		if m.cfg.Profile.Name == "" {
			m.cfg.Profile.Name = "Anonymous"
			_ = m.cfg.Save()
		}
		m.state = stateMenu
		return m, nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// --- menu ---

func (m Model) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		m.menuIndex = wrap(m.menuIndex-1, len(menuItems))
	case "down", "j":
		m.menuIndex = wrap(m.menuIndex+1, len(menuItems))
	case "q":
		m.quitting = true
		return m, tea.Quit
	case "enter", " ":
		switch m.menuIndex {
		case 0:
			m.modeType = "time"
			m.modeIndex = indexOf(timeOptions, m.cfg.Preferences.DefaultTime)
			m.state = stateModeSelect
		case 1:
			m.modeType = "word"
			m.modeIndex = indexOf(wordOptions, m.cfg.Preferences.DefaultWords)
			m.state = stateModeSelect
		case 2:
			m.state = stateStats
		case 3:
			m.input.SetValue(m.cfg.Profile.Name)
			m.settingsIndex = setName
			m.input.Blur()
			m.state = stateSettings
		case 4:
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

// --- mode selection ---

func (m Model) updateModeSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	opts := m.currentOptions()
	switch msg.String() {
	case "esc":
		m.state = stateMenu
	case "left", "h", "up", "k":
		m.modeIndex = wrap(m.modeIndex-1, len(opts))
	case "right", "l", "down", "j":
		m.modeIndex = wrap(m.modeIndex+1, len(opts))
	case "enter", " ":
		return m.startTest()
	}
	return m, nil
}

func (m Model) currentOptions() []int {
	if m.modeType == "word" {
		return wordOptions
	}
	return timeOptions
}

// startTest generates text and enters the countdown.
func (m Model) startTest() (tea.Model, tea.Cmd) {
	opts := m.currentOptions()
	val := opts[m.modeIndex]

	var count int
	if m.modeType == "time" {
		m.spec = testSpec{kind: "time", seconds: val}
		m.testDuration = time.Duration(val) * time.Second
		count = val*4 + 40 // generous buffer of words for the time limit
		m.cfg.Preferences.DefaultTime = val
		m.cfg.Preferences.DefaultMode = "time"
	} else {
		m.spec = testSpec{kind: "word", words: val}
		count = val
		m.cfg.Preferences.DefaultWords = val
		m.cfg.Preferences.DefaultMode = "word"
	}

	text := words.Generate(words.Options{
		Count:          count,
		IncludePunct:   m.cfg.Preferences.IncludePunct,
		IncludeNumbers: m.cfg.Preferences.IncludeNumbers,
		Rand:           m.rng,
	})
	m.session = typing.New(text)
	m.countdownStart = time.Now()
	m.state = stateCountdown
	return m, tick()
}

// --- typing ---

func (m Model) updateTyping(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		// Abandon the test without recording it.
		m.session = nil
		m.state = stateMenu
		return m, nil
	case tea.KeyBackspace, tea.KeyCtrlH:
		m.session.Backspace()
		return m, nil
	case tea.KeySpace:
		m.session.Type(' ')
	case tea.KeyRunes:
		for _, r := range msg.Runes {
			m.session.Type(r)
		}
	default:
		return m, nil
	}

	// Word mode finishes when the whole passage is typed.
	if m.spec.kind == "word" && m.session.Complete() {
		return m.finishTest()
	}
	return m, nil
}

// finishTest freezes the session, records results, and persists progress.
func (m Model) finishTest() (tea.Model, tea.Cmd) {
	m.session.End()
	stats := m.session.Stats()
	report := cheat.Analyze(m.session.Keystrokes, stats)

	m.res.stats = stats
	m.res.report = report
	m.res.newBest = false
	m.res.leveledUp = false
	m.res.xp = 0

	p := &m.cfg.Profile
	p.TotalTests++
	modeLabel := m.spec.kind

	if report.Suspicious {
		p.SuspiciousTests++
	} else {
		p.CleanTests++

		xp := int(stats.WPM)
		if stats.Accuracy >= 95 {
			xp += 20
		} else if stats.Accuracy >= 85 {
			xp += 10
		}
		m.res.xp = xp

		prevLevel := p.Level()
		p.XP += xp
		if p.Level() > prevLevel {
			m.res.leveledUp = true
		}
		if stats.WPM > p.BestWPM {
			p.BestWPM = stats.WPM
			m.res.newBest = true
		}
		if stats.Accuracy > p.BestAccuracy {
			p.BestAccuracy = stats.Accuracy
		}
	}

	m.cfg.AddHistory(config.HistoryEntry{
		Time:       time.Now(),
		WPM:        stats.WPM,
		Accuracy:   stats.Accuracy,
		Mode:       modeLabel,
		Suspicious: report.Suspicious,
	})
	m.res.saveErr = m.cfg.Save()

	m.state = stateResult
	return m, nil
}

// --- result ---

func (m Model) updateResult(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "r":
		// Retry the same spec.
		return m.startTest()
	default:
		m.state = stateMenu
		return m, nil
	}
}

// --- settings ---

func (m Model) updateSettings(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// When editing the name, route most keys to the text input.
	if m.settingsIndex == setName && m.input.Focused() {
		switch msg.Type {
		case tea.KeyEnter:
			m.commitName()
			m.input.Blur()
			return m, nil
		case tea.KeyEsc:
			m.input.SetValue(m.cfg.Profile.Name)
			m.input.Blur()
			return m, nil
		}
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	switch msg.String() {
	case "esc", "q":
		_ = m.cfg.Save()
		m.state = stateMenu
	case "up", "k":
		m.settingsIndex = wrap(m.settingsIndex-1, settingsRows)
	case "down", "j":
		m.settingsIndex = wrap(m.settingsIndex+1, settingsRows)
	case "enter", " ", "left", "h", "right", "l":
		return m.applySetting(msg.String())
	}
	return m, nil
}

func (m Model) applySetting(key string) (tea.Model, tea.Cmd) {
	switch m.settingsIndex {
	case setName:
		m.input.SetValue(m.cfg.Profile.Name)
		m.input.CursorEnd()
		return m, m.input.Focus()
	case setTheme:
		dir := 1
		if key == "left" || key == "h" {
			dir = -1
		}
		i := wrap(indexOfStr(ThemeOrder, m.cfg.Preferences.Theme)+dir, len(ThemeOrder))
		m.cfg.Preferences.Theme = ThemeOrder[i]
		m.styles = newStyles(themeByName(m.cfg.Preferences.Theme))
	case setPunct:
		m.cfg.Preferences.IncludePunct = !m.cfg.Preferences.IncludePunct
	case setNumbers:
		m.cfg.Preferences.IncludeNumbers = !m.cfg.Preferences.IncludeNumbers
	case setBack:
		_ = m.cfg.Save()
		m.state = stateMenu
	}
	return m, nil
}

func (m *Model) commitName() {
	if n := trimmed(m.input.Value()); n != "" {
		m.cfg.Profile.Name = n
		_ = m.cfg.Save()
	}
}

// --- helpers ---

func trimmed(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

func wrap(i, n int) int {
	if n == 0 {
		return 0
	}
	return ((i % n) + n) % n
}

func indexOf(s []int, v int) int {
	for i, x := range s {
		if x == v {
			return i
		}
	}
	return 0
}

func indexOfStr(s []string, v string) int {
	for i, x := range s {
		if x == v {
			return i
		}
	}
	return 0
}
