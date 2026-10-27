package ui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"wintop/internal/metrics"
)

func TestModelView(t *testing.T) {
	c := metrics.NewCollector(30)
	var m tea.Model = NewModel(c)

	// Simulate window size
	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 35})

	// Simulate tick
	m, _ = m.Update(TickMsg(time.Now()))

	output := m.View()
	if !strings.Contains(output, "wintop") {
		t.Fatalf("expected output to contain 'wintop', got:\n%s", output)
	}

	t.Logf("Rendered view preview (first 15 lines):\n")
	lines := strings.Split(output, "\n")
	for i := 0; i < len(lines) && i < 15; i++ {
		t.Log(lines[i])
	}
}
