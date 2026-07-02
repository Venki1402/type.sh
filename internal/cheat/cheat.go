// Package cheat analyzes a typing session for signs of automation or pasting.
//
// The old detector only looked at final WPM and whether the text matched. Now
// that we capture per-keystroke timing, we can inspect the actual rhythm of
// typing, which is far harder to fake than a headline WPM number.
package cheat

import (
	"math"
	"sort"

	"typesh/internal/typing"
)

// Severity ranks how strong a signal is.
type Severity int

const (
	Low Severity = iota
	Medium
	High
)

func (s Severity) String() string {
	switch s {
	case High:
		return "HIGH"
	case Medium:
		return "MEDIUM"
	default:
		return "LOW"
	}
}

// Flag is a single detected anomaly.
type Flag struct {
	Code     string
	Title    string
	Detail   string
	Severity Severity
}

// Report is the outcome of an analysis.
type Report struct {
	Suspicious bool
	Flags      []Flag
}

// Thresholds (kept as named constants so they are easy to tune/test).
const (
	inhumanWPM        = 220.0 // sustained net WPM above this is not human
	fastKeyMS         = 25.0  // an interval faster than this is "very fast"
	fastKeyRatioFlag  = 0.55  // if this fraction of keys are very fast -> flag
	roboticStdevMS    = 8.0   // interval stdev below this is machine-like
	pasteRunLength    = 10    // N consecutive near-instant keys => paste burst
	pasteRunMS        = 8.0   // interval considered "instant" for a paste run
	minSamples        = 12    // need enough keystrokes to judge timing
	perfectSpeedWPM   = 140.0 // perfect accuracy above this is doubtful
)

// Analyze inspects a session's keystroke log and computed stats.
func Analyze(ks []typing.Keystroke, st typing.Stats) Report {
	var flags []Flag

	// 1. Headline speed sanity check.
	if st.WPM > inhumanWPM {
		flags = append(flags, Flag{
			Code:     "INHUMAN_SPEED",
			Title:    "Inhuman speed",
			Detail:   "Net WPM exceeds the fastest recorded human typists.",
			Severity: High,
		})
	}

	intervals := intervalsMS(ks)

	if len(ks) >= minSamples && len(intervals) >= 2 {
		// 2. Too many near-instant keystrokes (macro / injected input).
		if r := fastRatio(intervals, fastKeyMS); r >= fastKeyRatioFlag {
			flags = append(flags, Flag{
				Code:     "MACHINE_GUN_KEYS",
				Title:    "Machine-gun keystrokes",
				Detail:   "Most keys arrived faster than human fingers can move.",
				Severity: High,
			})
		}

		// 3. Suspiciously even rhythm (real typing always has jitter).
		mean, stdev := meanStdev(intervals)
		if stdev < roboticStdevMS && mean < 90 {
			flags = append(flags, Flag{
				Code:     "ROBOTIC_RHYTHM",
				Title:    "Robotic rhythm",
				Detail:   "Keystroke timing is unnaturally uniform.",
				Severity: High,
			})
		}

		// 4. A long run of instant keystrokes = a pasted chunk.
		if run := longestRun(intervals, pasteRunMS); run >= pasteRunLength {
			flags = append(flags, Flag{
				Code:     "PASTE_BURST",
				Title:    "Paste burst",
				Detail:   "A block of characters appeared all at once.",
				Severity: High,
			})
		}
	}

	// 5. Flawless accuracy at high speed with zero corrections is unusual.
	if st.Accuracy >= 100 && st.WPM > perfectSpeedWPM && st.TotalKeys == correctOnly(ks) {
		flags = append(flags, Flag{
			Code:     "FLAWLESS_HIGH_SPEED",
			Title:    "Flawless at high speed",
			Detail:   "100% accuracy at high speed with no corrections.",
			Severity: Medium,
		})
	}

	// Verdict: any HIGH flag, or two or more MEDIUM flags, is suspicious.
	high, med := 0, 0
	for _, f := range flags {
		switch f.Severity {
		case High:
			high++
		case Medium:
			med++
		}
	}
	suspicious := high > 0 || med >= 2

	sort.SliceStable(flags, func(i, j int) bool {
		return flags[i].Severity > flags[j].Severity
	})

	return Report{Suspicious: suspicious, Flags: flags}
}

// intervalsMS returns inter-keystroke gaps in milliseconds.
func intervalsMS(ks []typing.Keystroke) []float64 {
	if len(ks) < 2 {
		return nil
	}
	out := make([]float64, 0, len(ks)-1)
	for i := 1; i < len(ks); i++ {
		d := float64((ks[i].Elapsed - ks[i-1].Elapsed).Microseconds()) / 1000.0
		if d >= 0 {
			out = append(out, d)
		}
	}
	return out
}

func fastRatio(intervals []float64, thresholdMS float64) float64 {
	if len(intervals) == 0 {
		return 0
	}
	fast := 0
	for _, v := range intervals {
		if v < thresholdMS {
			fast++
		}
	}
	return float64(fast) / float64(len(intervals))
}

func meanStdev(v []float64) (mean, stdev float64) {
	if len(v) == 0 {
		return 0, 0
	}
	var sum float64
	for _, x := range v {
		sum += x
	}
	mean = sum / float64(len(v))
	var variance float64
	for _, x := range v {
		variance += (x - mean) * (x - mean)
	}
	variance /= float64(len(v))
	return mean, math.Sqrt(variance)
}

// longestRun returns the length of the longest streak of consecutive intervals
// below thresholdMS (measured in keystrokes, hence +1).
func longestRun(intervals []float64, thresholdMS float64) int {
	best, cur := 0, 0
	for _, v := range intervals {
		if v < thresholdMS {
			cur++
			if cur > best {
				best = cur
			}
		} else {
			cur = 0
		}
	}
	if best == 0 {
		return 0
	}
	return best + 1
}

func correctOnly(ks []typing.Keystroke) int {
	n := 0
	for _, k := range ks {
		if k.Correct {
			n++
		}
	}
	return n
}
