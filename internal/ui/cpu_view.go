package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderCPUBox(width, height int) string {
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

	// Title line
	title := m.theme.TitleStyle.Render("cpu")
	displayName := snap.CPUName
	displayName = strings.ReplaceAll(displayName, " 8-Core Processor", "")
	displayName = strings.ReplaceAll(displayName, " Processor", "")
	cpuName := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Render(truncate(displayName, 22))

	pctColor := m.theme.MeterGreen
	if snap.CPUTotalPct > 60 {
		pctColor = m.theme.MeterYellow
	}
	if snap.CPUTotalPct > 85 {
		pctColor = m.theme.MeterRed
	}
	totalPctStr := lipgloss.NewStyle().Foreground(pctColor).Bold(true).Render(fmt.Sprintf("%5.1f%%", snap.CPUTotalPct))

	var headerParts []string
	headerParts = append(headerParts, title, cpuName)

	if snap.CPUPowerWatts > 0 {
		pwrStr := lipgloss.NewStyle().Foreground(m.theme.MeterYellow).Render(fmt.Sprintf("%.1f W", snap.CPUPowerWatts))
		headerParts = append(headerParts, pwrStr)
	}
	if snap.CPUFanRPM > 0 {
		fanStr := lipgloss.NewStyle().Foreground(m.theme.TextSecondary).Render(fmt.Sprintf("%d RPM", snap.CPUFanRPM))
		headerParts = append(headerParts, fanStr)
	}
	headerParts = append(headerParts, totalPctStr)

	header := strings.Join(headerParts, " ")

	// Available vertical space for content inside box
	remH := innerH - 1

	// Braille Graph
	graphH := remH / 2
	if graphH < 2 {
		graphH = 2
	}
	graphW := textW
	if graphW < 10 {
		graphW = 10
	}

	canvas := NewBrailleCanvas(graphW, graphH)
	canvas.RenderAreaGraph(m.snapshot.cpuHistory, 100.0)
	graphStr := lipgloss.NewStyle().Foreground(m.theme.GraphCPU).Render(canvas.String())

	// Mini core bars
	coresRemH := remH - graphH - 1
	var coreLines []string
	if coresRemH > 0 && len(snap.CPUCoresPct) > 0 {
		halfCores := (len(snap.CPUCoresPct) + 1) / 2
		coreLinesToDisplay := coresRemH
		if coreLinesToDisplay > halfCores {
			coreLinesToDisplay = halfCores
		}

		colWidth := (textW - 2) / 2
		barWidth := colWidth - 14
		if barWidth < 3 {
			barWidth = 3
		}

		for line := 0; line < coreLinesToDisplay; line++ {
			c1 := line
			c2 := line + halfCores

			str1 := formatCoreBar(c1, snap.CPUCoresPct[c1], barWidth, m.theme)
			str2 := ""
			if c2 < len(snap.CPUCoresPct) {
				str2 = formatCoreBar(c2, snap.CPUCoresPct[c2], barWidth, m.theme)
			}
			coreLines = append(coreLines, fmt.Sprintf("%-*s  %s", colWidth, str1, str2))
		}
	}

	content := header + "\n" + graphStr
	if len(coreLines) > 0 {
		content += "\n" + strings.Join(coreLines, "\n")
	}

	return m.theme.BoxStyle.
		Width(innerW).
		Height(innerH).
		Render(content)
}
