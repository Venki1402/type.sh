// Command type.sh is a terminal typing-speed trainer.
//
// Configuration and progress are stored globally (via the OS config directory)
// so your name, preferences, and stats persist between runs.
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"typesh/internal/config"
	"typesh/internal/tui"
)

// version is stamped at build time via -ldflags "-X main.version=...".
// It defaults to "dev" for `go run`/`go build` without ldflags.
var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v", "version":
			fmt.Printf("type.sh %s\n", version)
			return
		}
	}

	cfg, firstRun, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "type.sh: could not load config: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(tui.New(cfg, firstRun), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "type.sh: %v\n", err)
		os.Exit(1)
	}
}
