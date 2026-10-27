package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderProcBox(width, height int) string {
	innerW := width - 4
	if innerW < 30 {
		innerW = 30
	}
	textW := innerW - 2
	innerH := height - 2
	if innerH < 4 {
		innerH = 4
	}

	procs := m.getFilteredProcesses()
	totalProcs := len(m.snapshot.snap.Processes)

	// Title line
	title := m.theme.TitleStyle.Render("processes")
	procCountStr := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Render(
		fmt.Sprintf("(%d/%d)", len(procs), totalProcs),
	)

	titleLine := fmt.Sprintf("%s %s", title, procCountStr)
	if m.filterText != "" {
		filterBadge := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#282a36")).
			Background(lipgloss.Color("#ffb86c")).
			Padding(0, 1).
			Render("filter: " + m.filterText)
		titleLine += " " + filterBadge
	}

	// Table column widths
	pidW := 7
	cpuW := 7
	memPctW := 7
	rssW := 9
	thrW := 5
	// Remaining space for Program Name
	nameW := textW - (pidW + cpuW + memPctW + rssW + thrW + 5)
	if nameW < 10 {
		nameW = 10
	}

	// Column headers with sort arrows
	sortArrow := "▼"
	if m.sortAsc {
		sortArrow = "▲"
	}

	pidHdr := "PID"
	nameHdr := "NAME"
	cpuHdr := "CPU%"
	memHdr := "MEM%"
	switch m.sortBy {
	case SortByPID:
		pidHdr += sortArrow
	case SortByName:
		nameHdr += sortArrow
	case SortByCPU:
		cpuHdr += sortArrow
	case SortByMem:
		memHdr += sortArrow
	}

	hdrStyle := lipgloss.NewStyle().Foreground(m.theme.TextSecondary).Bold(true)
	headerRow := hdrStyle.Render(fmt.Sprintf("%-*s %-*s %*s %*s %*s %*s",
		pidW, pidHdr,
		nameW, nameHdr,
		cpuW, cpuHdr,
		memPctW, memHdr,
		rssW, "RSS",
		thrW, "THR",
	))

	// Available rows for process table
	tableRows := innerH - 2
	if tableRows < 1 {
		tableRows = 1
	}

	// Clamp selected index
	if m.selectedProcIdx >= len(procs) {
		m.selectedProcIdx = len(procs) - 1
	}
	if m.selectedProcIdx < 0 {
		m.selectedProcIdx = 0
	}

	// Calculate scrolling window
	scrollOffset := 0
	if m.selectedProcIdx >= tableRows {
		scrollOffset = m.selectedProcIdx - tableRows + 1
	}

	var rows []string
	rows = append(rows, titleLine, headerRow)

	for i := 0; i < tableRows; i++ {
		idx := scrollOffset + i
		if idx >= len(procs) {
			break
		}
		p := procs[idx]

		isSelected := idx == m.selectedProcIdx

		nameStr := truncate(p.Name, nameW)
		cpuStr := fmt.Sprintf("%5.1f", p.CPUPercent)
		memStr := fmt.Sprintf("%5.1f", p.MemPercent)
		rssStr := FormatBytes(p.MemRSS)
		thrStr := fmt.Sprintf("%3d", p.NumThreads)

		line := fmt.Sprintf("%-*d %-*s %*s %*s %*s %*s",
			pidW, p.PID,
			nameW, nameStr,
			cpuW, cpuStr,
			memPctW, memStr,
			rssW, rssStr,
			thrW, thrStr,
		)

		if isSelected {
			line = lipgloss.NewStyle().
				Background(m.theme.SelectedBg).
				Foreground(m.theme.SelectedFg).
				Bold(true).
				Render(line)
		} else {
			line = lipgloss.NewStyle().Foreground(m.theme.TextPrimary).Render(line)
		}

		rows = append(rows, line)
	}

	return m.theme.BoxStyle.
		Width(innerW).
		Height(innerH).
		Render(strings.Join(rows, "\n"))
}
