package cheat

import (
	"testing"
	"time"

	"typesh/internal/typing"
)

// buildKeystrokes creates a keystroke log with a fixed gap (ms) between keys.
func buildKeystrokes(n int, gapMS int, allCorrect bool) []typing.Keystroke {
	ks := make([]typing.Keystroke, n)
	var t time.Duration
	for i := 0; i < n; i++ {
		ks[i] = typing.Keystroke{
			Rune:    'a',
			Elapsed: t,
			Correct: allCorrect,
		}
		t += time.Duration(gapMS) * time.Millisecond
	}
	return ks
}

func hasFlag(r Report, code string) bool {
	for _, f := range r.Flags {
		if f.Code == code {
			return true
		}
	}
	return false
}

func TestHumanLikeTypingIsClean(t *testing.T) {
	// ~130ms per key with natural jitter → ~90 wpm, human.
	ks := make([]typing.Keystroke, 60)
	var elapsed time.Duration
	jitter := []int{90, 150, 110, 200, 80, 170, 120, 140}
	for i := range ks {
		ks[i] = typing.Keystroke{Rune: 'a', Elapsed: elapsed, Correct: i%13 != 0}
		elapsed += time.Duration(jitter[i%len(jitter)]) * time.Millisecond
	}
	st := typing.Stats{WPM: 90, Accuracy: 92, TotalKeys: 60}
	rep := Analyze(ks, st)
	if rep.Suspicious {
		t.Fatalf("expected clean, got flags: %+v", rep.Flags)
	}
}

func TestRoboticRhythmFlagged(t *testing.T) {
	// Perfectly even 40ms gaps → zero jitter, machine-like.
	ks := buildKeystrokes(60, 40, true)
	st := typing.Stats{WPM: 120, Accuracy: 100, TotalKeys: 60}
	rep := Analyze(ks, st)
	if !rep.Suspicious {
		t.Fatal("expected robotic rhythm to be flagged")
	}
	if !hasFlag(rep, "ROBOTIC_RHYTHM") {
		t.Fatalf("expected ROBOTIC_RHYTHM, got %+v", rep.Flags)
	}
}

func TestPasteBurstFlagged(t *testing.T) {
	// 30 keys at 2ms apart = an instant block.
	ks := buildKeystrokes(30, 2, true)
	st := typing.Stats{WPM: 300, Accuracy: 100, TotalKeys: 30}
	rep := Analyze(ks, st)
	if !rep.Suspicious {
		t.Fatal("expected paste burst to be flagged")
	}
	if !hasFlag(rep, "PASTE_BURST") && !hasFlag(rep, "MACHINE_GUN_KEYS") {
		t.Fatalf("expected paste/machine-gun flag, got %+v", rep.Flags)
	}
}

func TestInhumanSpeedFlagged(t *testing.T) {
	ks := buildKeystrokes(60, 100, true)
	st := typing.Stats{WPM: 250, Accuracy: 100, TotalKeys: 60}
	rep := Analyze(ks, st)
	if !hasFlag(rep, "INHUMAN_SPEED") {
		t.Fatalf("expected INHUMAN_SPEED, got %+v", rep.Flags)
	}
}
