package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"wintop/internal/metrics"
	"wintop/internal/ui"
)

func main() {
	// Initialize metrics collector with 120-second history buffer
	collector := metrics.NewCollector(120)
	collector.Start(1 * time.Second)
	defer collector.Stop()

	// Initialize Bubbletea model
	model := ui.NewModel(collector)

	// Launch Bubbletea program with Alternate Screen Buffer
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running wintop: %v\n", err)
		os.Exit(1)
	}
}
