package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Theme defines the color palette inspired by btop's Dracula / Tokyo-Night palettes.
type Theme struct {
	BorderNormal lipgloss.Color
	BorderActive lipgloss.Color

	TitleStyle lipgloss.Style
	BoxStyle   lipgloss.Style

	TextPrimary   lipgloss.Color
	TextSecondary lipgloss.Color
	TextMuted     lipgloss.Color

	GraphCPU lipgloss.Color
	GraphNet lipgloss.Color

	MeterGreen  lipgloss.Color
	MeterYellow lipgloss.Color
	MeterRed    lipgloss.Color

	SelectedBg lipgloss.Color
	SelectedFg lipgloss.Color
}

// DefaultTheme returns the default neon btop-like color scheme.
func DefaultTheme() Theme {
	return Theme{
		BorderNormal:  lipgloss.Color("#44475a"),
		BorderActive:  lipgloss.Color("#bd93f9"), // Purple
		TextPrimary:   lipgloss.Color("#f8f8f2"),
		TextSecondary: lipgloss.Color("#8be9fd"), // Cyan
		TextMuted:     lipgloss.Color("#6272a4"), // Gray

		GraphCPU: lipgloss.Color("#50fa7b"), // Green
		GraphNet: lipgloss.Color("#ff79c6"), // Pink

		MeterGreen:  lipgloss.Color("#50fa7b"),
		MeterYellow: lipgloss.Color("#f1fa8c"),
		MeterRed:    lipgloss.Color("#ff5555"),

		SelectedBg: lipgloss.Color("#bd93f9"),
		SelectedFg: lipgloss.Color("#282a36"),

		TitleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#bd93f9")),

		BoxStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#44475a")).
			Padding(0, 1),
	}
}

// FormatBytes converts raw bytes to human-readable strings.
func FormatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	suffixes := []string{"KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%.1f %s", float64(b)/float64(div), suffixes[exp])
}

// FormatBitrate converts bytes/sec to human readable rate.
func FormatBitrate(bytesPerSec float64) string {
	if bytesPerSec < 1024 {
		return fmt.Sprintf("%.0f B/s", bytesPerSec)
	} else if bytesPerSec < 1024*1024 {
		return fmt.Sprintf("%.1f KB/s", bytesPerSec/1024)
	} else if bytesPerSec < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB/s", bytesPerSec/(1024*1024))
	}
	return fmt.Sprintf("%.2f GB/s", bytesPerSec/(1024*1024*1024))
}

// FormatDuration converts seconds to d h m s string.
func FormatDuration(seconds uint64) string {
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	mins := (seconds % 3600) / 60
	secs := seconds % 60
	if days > 0 {
		return fmt.Sprintf("%dd %02dh %02dm", days, hours, mins)
	}
	if hours > 0 {
		return fmt.Sprintf("%02dh %02dm %02ds", hours, mins, secs)
	}
	return fmt.Sprintf("%02dm %02ds", mins, secs)
}

// RenderProgressBar draws a filled bar of given width with gradient colors.
func RenderProgressBar(pct float64, width int, th Theme) string {
	if width < 3 {
		width = 3
	}
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}

	filled := int((pct / 100.0) * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	var barColor lipgloss.Color
	if pct < 60 {
		barColor = th.MeterGreen
	} else if pct < 85 {
		barColor = th.MeterYellow
	} else {
		barColor = th.MeterRed
	}

	filledPart := lipgloss.NewStyle().Foreground(barColor).Render(strings.Repeat("■", filled))
	emptyPart := lipgloss.NewStyle().Foreground(th.BorderNormal).Render(strings.Repeat("⬝", empty))

	return filledPart + emptyPart
}
