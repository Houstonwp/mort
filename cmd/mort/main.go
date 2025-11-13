package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"mort/internal/xtbmlcli"
	"mort/tui"
)

func main() {
	// Allow running the converter via `mort --convert`.
	if len(os.Args) > 1 && os.Args[1] == "--convert" {
		code := xtbmlcli.Run(os.Args[2:], os.Stdout, os.Stderr)
		os.Exit(code)
	}

	jsonDir := os.Getenv("MORT_JSON_DIR")
	model := tui.NewModel(jsonDir)
	if err := tea.NewProgram(model).Start(); err != nil {
		fmt.Fprintf(os.Stderr, "tui error: %v\n", err)
		os.Exit(1)
	}
}
