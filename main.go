package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"wintop/internal/metrics"
	"wintop/internal/ui"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-v", "--version", "version":
			fmt.Printf("wintop %s\n", version)
			return
		case "-h", "--help", "help":
			fmt.Printf("wintop %s - terminal resource and process monitor for Windows\n\nUsage:\n  wintop [options]\n\nOptions:\n  -v, --version  Print version and exit\n  -h, --help     Print help and exit\n", version)
			return
		}
	}

	// Initialize metrics collector with 120-second history buffer
	collector := metrics.NewCollector(120)
	collector.Start(1 * time.Second)
	defer collector.Stop()

	// Initialize Bubbletea model
	model := ui.NewModel(collector, version)

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
