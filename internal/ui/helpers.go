package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 1 {
		return "…"
	}
	return s[:maxLen-1] + "…"
}

func formatCoreBar(coreIdx int, pct float64, barW int, th Theme) string {
	lbl := lipgloss.NewStyle().Foreground(th.TextMuted).Render(fmt.Sprintf("C%02d", coreIdx))
	bar := RenderProgressBar(pct, barW, th)
	val := lipgloss.NewStyle().Foreground(th.TextPrimary).Render(fmt.Sprintf("%3.0f%%", pct))
	return fmt.Sprintf("%s %s %s", lbl, bar, val)
}
