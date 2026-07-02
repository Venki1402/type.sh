// Package typing implements the core typing-test session: keystroke capture,
// live progress, and WPM/accuracy/consistency calculations.
package typing

import (
	"math"
	"time"
)

// Keystroke records a single character press with its timing and correctness.
type Keystroke struct {
	Rune    rune          // the character that was typed
	Elapsed time.Duration // time since the session started
	Correct bool          // whether it matched the target at that position
}

// Session tracks the state of an in-progress typing test.
type Session struct {
	Target     []rune
	Input      []rune
	Keystrokes []Keystroke // append-only log of character presses (no backspaces)

	startedAt time.Time
	started   bool
	ended     bool
	endedAt   time.Time
}

// New creates a session for the given target text.
func New(target string) *Session {
	return &Session{Target: []rune(target)}
}

// Started reports whether the first keystroke has been registered.
func (s *Session) Started() bool { return s.started }

// start marks the session as begun on the first keystroke.
func (s *Session) start(now time.Time) {
	if !s.started {
		s.started = true
		s.startedAt = now
	}
}

// Type registers a character press. It is a no-op once the session has ended.
func (s *Session) Type(r rune) {
	if s.ended {
		return
	}
	now := time.Now()
	s.start(now)

	pos := len(s.Input)
	correct := pos < len(s.Target) && s.Target[pos] == r

	s.Input = append(s.Input, r)
	s.Keystrokes = append(s.Keystrokes, Keystroke{
		Rune:    r,
		Elapsed: now.Sub(s.startedAt),
		Correct: correct,
	})
}

// Backspace removes the last typed character (the keystroke log is preserved
// so that accuracy reflects every press the user actually made).
func (s *Session) Backspace() {
	if s.ended || len(s.Input) == 0 {
		return
	}
	s.Input = s.Input[:len(s.Input)-1]
}

// End freezes the session, recording the finish time.
func (s *Session) End() {
	if s.ended {
		return
	}
	s.ended = true
	if s.started {
		s.endedAt = time.Now()
	}
}

// Ended reports whether the session is finished.
func (s *Session) Ended() bool { return s.ended }

// Complete reports whether the whole target has been typed (word-mode finish).
func (s *Session) Complete() bool { return len(s.Input) >= len(s.Target) }

// Elapsed returns how long the session has been running.
func (s *Session) Elapsed() time.Duration {
	if !s.started {
		return 0
	}
	if s.ended {
		return s.endedAt.Sub(s.startedAt)
	}
	return time.Since(s.startedAt)
}

// Stats is a snapshot of computed performance metrics.
type Stats struct {
	WPM         float64       // net WPM: correct chars / 5 / minutes
	RawWPM      float64       // gross WPM: all chars / 5 / minutes
	Accuracy    float64       // correct presses / total presses (0-100)
	Errors      int           // incorrect presses
	CorrectKeys int           // correct presses
	TotalKeys   int           // total presses (excludes backspaces)
	Consistency float64       // 0-100, higher = steadier rhythm
	Elapsed     time.Duration // duration used for the calculation
}

// Stats computes metrics from the keystroke log.
func (s *Session) Stats() Stats {
	elapsed := s.Elapsed()
	minutes := elapsed.Minutes()

	total := len(s.Keystrokes)
	correct := 0
	for _, k := range s.Keystrokes {
		if k.Correct {
			correct++
		}
	}
	errors := total - correct

	var wpm, raw, acc float64
	if minutes > 0 {
		wpm = (float64(correct) / 5.0) / minutes
		raw = (float64(total) / 5.0) / minutes
	}
	if total > 0 {
		acc = float64(correct) / float64(total) * 100
	}

	return Stats{
		WPM:         wpm,
		RawWPM:      raw,
		Accuracy:    acc,
		Errors:      errors,
		CorrectKeys: correct,
		TotalKeys:   total,
		Consistency: consistency(s.Keystrokes),
		Elapsed:     elapsed,
	}
}

// consistency measures rhythm steadiness from inter-keystroke intervals using
// the coefficient of variation, mapped to a 0-100 score (100 = perfectly even).
func consistency(ks []Keystroke) float64 {
	if len(ks) < 3 {
		return 0
	}
	intervals := make([]float64, 0, len(ks)-1)
	for i := 1; i < len(ks); i++ {
		d := (ks[i].Elapsed - ks[i-1].Elapsed).Seconds()
		if d > 0 {
			intervals = append(intervals, d)
		}
	}
	if len(intervals) < 2 {
		return 0
	}

	var sum float64
	for _, v := range intervals {
		sum += v
	}
	mean := sum / float64(len(intervals))
	if mean == 0 {
		return 0
	}

	var variance float64
	for _, v := range intervals {
		variance += (v - mean) * (v - mean)
	}
	variance /= float64(len(intervals))
	cv := math.Sqrt(variance) / mean

	score := 100 * (1 - cv)
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	return score
}
