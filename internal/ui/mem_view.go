package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderMemBox(width, height int) string {
	innerW := width - 4
	if innerW < 30 {
		innerW = 30
	}
	textW := innerW - 2
	innerH := height - 2
	if innerH < 4 {
		innerH = 4
	}

	snap := m.snapshot.snap

	title := m.theme.TitleStyle.Render("mem / disks")

	var lines []string
	lines = append(lines, title)

	// Available bar width
	barW := textW - 32
	if barW < 5 {
		barW = 5
	}

	// RAM Line
	ramLbl := lipgloss.NewStyle().Foreground(m.theme.TextSecondary).Bold(true).Render("RAM ")
	ramBar := RenderProgressBar(snap.MemUsedPercent, barW, m.theme)
	ramVal := lipgloss.NewStyle().Foreground(m.theme.TextPrimary).Render(
		fmt.Sprintf("%s / %s (%3.0f%%)", FormatBytes(snap.MemUsed), FormatBytes(snap.MemTotal), snap.MemUsedPercent),
	)
	lines = append(lines, fmt.Sprintf("%s %s %s", ramLbl, ramBar, ramVal))

	// Swap / Pagefile Line
	if snap.SwapTotal > 0 {
		swapLbl := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Bold(true).Render("PAGE")
		swapBar := RenderProgressBar(snap.SwapUsedPercent, barW, m.theme)
		swapVal := lipgloss.NewStyle().Foreground(m.theme.TextPrimary).Render(
			fmt.Sprintf("%s / %s (%3.0f%%)", FormatBytes(snap.SwapUsed), FormatBytes(snap.SwapTotal), snap.SwapUsedPercent),
		)
		lines = append(lines, fmt.Sprintf("%s %s %s", swapLbl, swapBar, swapVal))
	}

	// Disks
	if len(snap.Disks) > 0 {
		lines = append(lines, "")
		diskHeader := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Render("DRIVE   USED / TOTAL         USAGE")
		lines = append(lines, diskHeader)

		for _, d := range snap.Disks {
			dBarW := textW - 36
			if dBarW < 4 {
				dBarW = 4
			}
			mount := lipgloss.NewStyle().Foreground(m.theme.TextSecondary).Bold(true).Render(fmt.Sprintf("%-6s", d.Mountpoint))
			usedTotal := fmt.Sprintf("%-18s", fmt.Sprintf("%s / %s", FormatBytes(d.Used), FormatBytes(d.Total)))
			bar := RenderProgressBar(d.UsedPercent, dBarW, m.theme)
			pctStr := fmt.Sprintf("%3.0f%%", d.UsedPercent)

			lines = append(lines, fmt.Sprintf("%s  %s  %s %s", mount, usedTotal, bar, pctStr))
		}
	}

	return m.theme.BoxStyle.
		Width(innerW).
		Height(innerH).
		Render(strings.Join(lines, "\n"))
}
