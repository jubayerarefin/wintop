package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderGPUBox(width, height int) string {
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
	if len(snap.GPUs) == 0 {
		return m.theme.BoxStyle.
			Width(innerW).
			Height(innerH).
			Render(m.theme.TitleStyle.Render("gpu") + "\n" + lipgloss.NewStyle().Foreground(m.theme.TextMuted).Render("No dedicated GPU detected"))
	}

	gpu := snap.GPUs[0]

	// Title line
	title := m.theme.TitleStyle.Render("gpu")
	displayName := strings.TrimPrefix(gpu.Name, "NVIDIA ")
	gpuName := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Render(truncate(displayName, 28))

	pctColor := m.theme.MeterGreen
	if gpu.UtilizationGPU > 60 {
		pctColor = m.theme.MeterYellow
	}
	if gpu.UtilizationGPU > 85 {
		pctColor = m.theme.MeterRed
	}
	utilStr := lipgloss.NewStyle().Foreground(pctColor).Bold(true).Render(fmt.Sprintf("%5.1f%%", gpu.UtilizationGPU))

	tempStr := ""
	if gpu.Temperature > 0 {
		tempStr = lipgloss.NewStyle().Foreground(m.theme.TextSecondary).Render(fmt.Sprintf("%.0f°C", gpu.Temperature))
	}

	header := fmt.Sprintf("%s %s %s %s", title, gpuName, tempStr, utilStr)

	// Available space
	remH := innerH - 1

	// Meters: PROC %, VRAM, FAN %
	barW := textW - 32
	if barW < 5 {
		barW = 5
	}

	procLbl := lipgloss.NewStyle().Foreground(m.theme.GraphGPU).Bold(true).Render("PROC")
	procBar := RenderProgressBar(gpu.UtilizationGPU, barW, m.theme)
	procVal := lipgloss.NewStyle().Foreground(m.theme.TextPrimary).Render(fmt.Sprintf("%5.1f%%", gpu.UtilizationGPU))
	procLine := fmt.Sprintf("%s %s %s", procLbl, procBar, procVal)

	vramLbl := lipgloss.NewStyle().Foreground(m.theme.TextSecondary).Bold(true).Render("VRAM")
	vramBar := RenderProgressBar(gpu.MemUsedPercent, barW, m.theme)
	vramVal := lipgloss.NewStyle().Foreground(m.theme.TextPrimary).Render(
		fmt.Sprintf("%s / %s (%3.0f%%)",
			FormatBytes(gpu.MemUsedMB*1024*1024),
			FormatBytes(gpu.MemTotalMB*1024*1024),
			gpu.MemUsedPercent),
	)
	vramLine := fmt.Sprintf("%s %s %s", vramLbl, vramBar, vramVal)

	fanLbl := lipgloss.NewStyle().Foreground(m.theme.TextMuted).Bold(true).Render("FAN ")
	fanBar := RenderProgressBar(gpu.FanSpeedPct, barW, m.theme)
	fanVal := lipgloss.NewStyle().Foreground(m.theme.TextPrimary).Render(fmt.Sprintf("%5.1f%%", gpu.FanSpeedPct))
	fanLine := fmt.Sprintf("%s %s %s", fanLbl, fanBar, fanVal)

	meters := []string{procLine, vramLine, fanLine}

	// If we have enough vertical space, render a Braille graph for GPU processing %
	graphH := remH - len(meters)
	var contentLines []string
	contentLines = append(contentLines, header)

	if graphH >= 2 {
		canvas := NewBrailleCanvas(textW, graphH)
		canvas.RenderAreaGraph(m.snapshot.gpuHistory, 100.0)
		graphStr := lipgloss.NewStyle().Foreground(m.theme.GraphGPU).Render(canvas.String())
		contentLines = append(contentLines, graphStr)
	}

	contentLines = append(contentLines, meters...)

	return m.theme.BoxStyle.
		Width(innerW).
		Height(innerH).
		Render(strings.Join(contentLines, "\n"))
}
