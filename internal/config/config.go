// Package config handles loading and persisting the user's global
// configuration and profile so settings and progress survive between runs.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// appDir is the sub-directory used inside the OS config directory.
const appDir = "type.sh"

// configFile is the name of the on-disk config file.
const configFile = "config.json"

// maxHistory bounds how many recent tests we keep on disk.
const maxHistory = 100

// Profile holds the persistent player progression data.
type Profile struct {
	Name            string  `json:"name"`
	BestWPM         float64 `json:"best_wpm"`
	BestAccuracy    float64 `json:"best_accuracy"`
	TotalTests      int     `json:"total_tests"`
	CleanTests      int     `json:"clean_tests"`
	SuspiciousTests int     `json:"suspicious_tests"`
	XP              int     `json:"xp"`
}

// Level is derived from XP (100 XP per level, starting at level 1).
func (p Profile) Level() int { return p.XP/100 + 1 }

// XPInLevel returns progress within the current level (0-99).
func (p Profile) XPInLevel() int { return p.XP % 100 }

// XPToNext returns XP remaining until the next level.
func (p Profile) XPToNext() int { return 100 - p.XPInLevel() }

// Preferences holds the user's chosen defaults so they are not re-entered.
type Preferences struct {
	Theme          string `json:"theme"`          // "default" | "mono" | "warm"
	DefaultMode    string `json:"default_mode"`   // "time" | "word"
	DefaultTime    int    `json:"default_time"`   // seconds
	DefaultWords   int    `json:"default_words"`  // word count
	IncludePunct   bool   `json:"include_punct"`  // add punctuation
	IncludeNumbers bool   `json:"include_numbers"` // add numbers
}

// HistoryEntry is a single recorded test.
type HistoryEntry struct {
	Time       time.Time `json:"time"`
	WPM        float64   `json:"wpm"`
	Accuracy   float64   `json:"accuracy"`
	Mode       string    `json:"mode"`
	Suspicious bool      `json:"suspicious"`
}

// Config is the full persisted document.
type Config struct {
	Profile     Profile        `json:"profile"`
	Preferences Preferences    `json:"preferences"`
	History     []HistoryEntry `json:"history"`

	path string // resolved on-disk location, not serialized
}

// Dir returns the directory where config is stored, creating it if needed.
func Dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, appDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// defaults returns a Config populated with sensible starting values.
func defaults(path string) *Config {
	return &Config{
		Preferences: Preferences{
			Theme:        "default",
			DefaultMode:  "time",
			DefaultTime:  30,
			DefaultWords: 30,
		},
		path: path,
	}
}

// Load reads the config from disk, returning defaults when none exists.
// The second return value reports whether this is a first run (no file yet).
func Load() (*Config, bool, error) {
	dir, err := Dir()
	if err != nil {
		return nil, false, err
	}
	path := filepath.Join(dir, configFile)

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return defaults(path), true, nil
	}
	if err != nil {
		return nil, false, err
	}

	cfg := defaults(path)
	if err := json.Unmarshal(data, cfg); err != nil {
		// Corrupt file: start fresh rather than crash, but keep the path.
		return defaults(path), true, nil
	}
	cfg.path = path
	firstRun := cfg.Profile.Name == ""
	return cfg, firstRun, nil
}

// Save writes the config back to disk atomically.
func (c *Config) Save() error {
	if c.path == "" {
		dir, err := Dir()
		if err != nil {
			return err
		}
		c.path = filepath.Join(dir, configFile)
	}

	if len(c.History) > maxHistory {
		c.History = c.History[len(c.History)-maxHistory:]
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	tmp := c.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, c.path)
}

// Path returns the resolved config file location (for display).
func (c *Config) Path() string { return c.path }

// AddHistory appends an entry and trims to the retention limit.
func (c *Config) AddHistory(e HistoryEntry) {
	c.History = append(c.History, e)
	if len(c.History) > maxHistory {
		c.History = c.History[len(c.History)-maxHistory:]
	}
}
