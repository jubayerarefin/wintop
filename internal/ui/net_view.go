package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderNetBox(width, height int) string {
	innerW := width - 4
	if innerW < 30 {
		innerW = 30
	}
	textW := innerW - 2
	innerH := height - 2
	if innerH < 3 {
		innerH = 3
	}

	snap := m.snapshot.snap

	title := m.theme.TitleStyle.Render("net")

	rxStr := lipgloss.NewStyle().Foreground(m.theme.GraphNet).Bold(true).Render(
		fmt.Sprintf("▼ RX: %-10s (%s)", FormatBitrate(snap.NetBytesRecvPerSec), FormatBytes(snap.NetTotalRecv)),
	)
	txStr := lipgloss.NewStyle().Foreground(m.theme.TextSecondary).Bold(true).Render(
		fmt.Sprintf("▲ TX: %-10s (%s)", FormatBitrate(snap.NetBytesSentPerSec), FormatBytes(snap.NetTotalSent)),
	)

	header := fmt.Sprintf("%s   %s   %s", title, rxStr, txStr)

	// Inner space for braille graph
	graphH := innerH - 1
	if graphH < 2 {
		graphH = 2
	}
	graphW := textW
	if graphW < 10 {
		graphW = 10
	}

	var maxVal float64 = 1024 * 10 // at least 10 KB/s
	for _, v := range m.snapshot.netRecvHist {
		if v > maxVal {
			maxVal = v
		}
	}
	for _, v := range m.snapshot.netSentHist {
		if v > maxVal {
			maxVal = v
		}
	}

	canvas := NewBrailleCanvas(graphW, graphH)
	canvas.RenderAreaGraph(m.snapshot.netRecvHist, maxVal)
	graphStr := lipgloss.NewStyle().Foreground(m.theme.GraphNet).Render(canvas.String())

	content := strings.Join([]string{header, graphStr}, "\n")

	return m.theme.BoxStyle.
		Width(innerW).
		Height(innerH).
		Render(content)
}
