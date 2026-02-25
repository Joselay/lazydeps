package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/smaetongmenglay/lazydeps/internal/ui"
)

func main() {
	dir := "."
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	// Resolve to absolute path
	if abs, err := os.Getwd(); err == nil && dir == "." {
		dir = abs
	}

	m := ui.NewModel(dir)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
