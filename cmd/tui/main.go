package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"go-with-go/tui"
)

func main() {
	p := tea.NewProgram(tui.New(), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "tui: %v\n", err)
		os.Exit(1)
	}
}
