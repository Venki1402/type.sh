// Package words provides the word bank and text generation for typing tests.
package words

import (
	"math/rand"
	"strings"
)

// Bank is the pool of common English words used to build test passages.
var Bank = []string{
	"the", "of", "and", "a", "to", "in", "is", "you", "that", "it",
	"he", "was", "for", "on", "are", "as", "with", "his", "they", "at",
	"be", "this", "have", "from", "or", "one", "had", "by", "word", "but",
	"not", "what", "all", "were", "we", "when", "your", "can", "said", "there",
	"each", "which", "she", "do", "how", "their", "time", "will", "about", "if",
	"up", "out", "many", "then", "them", "these", "so", "some", "her", "would",
	"make", "like", "into", "him", "has", "two", "more", "very", "after", "words",
	"first", "been", "who", "now", "find", "long", "down", "day", "did", "get",
	"come", "made", "may", "part", "over", "new", "sound", "take", "only", "little",
	"work", "know", "place", "year", "live", "me", "back", "give", "most", "good",
	"woman", "through", "just", "form", "great", "think", "help", "low", "line", "before",
	"turn", "cause", "same", "mean", "differ", "move", "right", "boy", "old", "too",
	"does", "tell", "sentence", "set", "three", "want", "air", "well", "also", "play",
	"small", "end", "put", "home", "read", "hand", "port", "large", "spell", "add",
	"even", "land", "here", "must", "big", "high", "such", "follow", "act", "why",
	"ask", "men", "change", "went", "light", "kind", "off", "need", "house", "picture",
}

// Options controls how a passage is generated.
type Options struct {
	Count          int
	IncludePunct   bool
	IncludeNumbers bool
	Rand           *rand.Rand
}

var punctuation = []string{",", ".", ";", ":", "!", "?"}

// Generate builds a space-separated passage of Count words honoring the options.
func Generate(opts Options) string {
	r := opts.Rand
	if r == nil {
		r = rand.New(rand.NewSource(1))
	}
	if opts.Count <= 0 {
		opts.Count = 30
	}

	out := make([]string, 0, opts.Count)
	for i := 0; i < opts.Count; i++ {
		w := Bank[r.Intn(len(Bank))]

		// Occasionally swap a word for a short number.
		if opts.IncludeNumbers && r.Intn(6) == 0 {
			w = number(r)
		}

		// Occasionally capitalize the first letter after sentence end.
		if opts.IncludePunct && i > 0 && r.Intn(8) == 0 {
			w = strings.Title(w)
		}

		// Occasionally append punctuation.
		if opts.IncludePunct && r.Intn(7) == 0 {
			w += punctuation[r.Intn(len(punctuation))]
		}

		out = append(out, w)
	}
	return strings.Join(out, " ")
}

func number(r *rand.Rand) string {
	n := r.Intn(4) + 1 // 1-4 digits
	digits := make([]byte, n)
	for i := range digits {
		digits[i] = byte('0' + r.Intn(10))
	}
	return string(digits)
}
