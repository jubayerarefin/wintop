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
	displayName = strings.TrimSpace(displayName)

	var topoStr string
	if snap.CPUPhysicalCores > 0 && snap.CPULogicalCores > 0 {
		topoStr = fmt.Sprintf("(%dC/%dT)", snap.CPUPhysicalCores, snap.CPULogicalCores)
	} else if snap.CPUCores > 0 {
		topoStr = fmt.Sprintf("(%dC)", snap.CPUCores)
	}

	cpuName := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Render(truncate(displayName, 18))
	topoStyled := lipgloss.NewStyle().Foreground(m.theme.TextSecondary).Render(topoStr)

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
	if topoStr != "" {
		headerParts = append(headerParts, topoStyled)
	}

	if snap.CPUPowerWatts > 0 {
		pwrStr := lipgloss.NewStyle().Foreground(m.theme.MeterYellow).Render(fmt.Sprintf("%.1f W", snap.CPUPowerWatts))
		headerParts = append(headerParts, pwrStr)
	} else {
		headerParts = append(headerParts, lipgloss.NewStyle().Foreground(m.theme.TextMuted).Render("PWR: --"))
	}

	if snap.CPUFanRPM > 0 {
		fanStr := lipgloss.NewStyle().Foreground(m.theme.TextSecondary).Render(fmt.Sprintf("%d RPM", snap.CPUFanRPM))
		headerParts = append(headerParts, fanStr)
	} else {
		headerParts = append(headerParts, lipgloss.NewStyle().Foreground(m.theme.TextMuted).Render("FAN: N/A"))
	}

	headerParts = append(headerParts, totalPctStr)
	header := strings.Join(headerParts, " ")

	// Available vertical space for content inside box
	remH := innerH - 1

	// Determine column layout for all CPU cores/threads
	totalCores := len(snap.CPUCoresPct)
	numCols := 1
	if textW >= 50 && totalCores >= 4 {
		numCols = 4
	} else if textW >= 28 && totalCores >= 2 {
		numCols = 2
	}

	var coreLines []string
	if totalCores > 0 {
		rowsNeeded := (totalCores + numCols - 1) / numCols
		maxRows := remH - 2 // Leave at least 2 rows for graph
		if maxRows < 1 {
			maxRows = 1
		}
		rowsToDisplay := rowsNeeded
		if rowsToDisplay > maxRows {
			rowsToDisplay = maxRows
		}

		colWidth := (textW - (numCols - 1)) / numCols
		barWidth := colWidth - 9
		if barWidth < 2 {
			barWidth = 2
		}

		for r := 0; r < rowsToDisplay; r++ {
			var colStrs []string
			for c := 0; c < numCols; c++ {
				idx := r + c*rowsNeeded
				if idx < totalCores {
					colStrs = append(colStrs, formatCoreBar(idx, snap.CPUCoresPct[idx], barWidth, m.theme))
				} else {
					colStrs = append(colStrs, strings.Repeat(" ", colWidth))
				}
			}
			coreLines = append(coreLines, strings.Join(colStrs, " "))
		}
	}

	// Braille Graph occupies remaining space
	graphH := remH - len(coreLines)
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

	content := header + "\n" + graphStr
	if len(coreLines) > 0 {
		content += "\n" + strings.Join(coreLines, "\n")
	}

	return m.theme.BoxStyle.
		Width(innerW).
		Height(innerH).
		Render(content)
}
