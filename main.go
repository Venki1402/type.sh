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

func main() {
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
