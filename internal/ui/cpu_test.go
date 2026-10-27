package ui

import (
	"strings"
	"testing"

	"wintop/internal/metrics"
)

func TestCPUViewPowerFan(t *testing.T) {
	snap := metrics.SystemSnapshot{
		CPUName:       "AMD Ryzen 7 5700X 8-Core Processor",
		CPUTotalPct:   15.4,
		CPUCoresPct:   []float64{10, 20, 15, 5, 8, 12, 14, 16},
		CPUPowerWatts: 52.8,
		CPUFanRPM:     1250,
	}

	m := Model{
		theme: DefaultTheme(),
		snapshot: SystemSnapshotWrapper{
			snap:       snap,
			cpuHistory: []float64{10, 12, 14, 15.4},
		},
	}

	rendered := m.renderCPUBox(60, 10)
	t.Logf("Rendered CPU Box:\n%s", rendered)

	if !strings.Contains(rendered, "52.8 W") {
		t.Fatalf("expected power in CPU box, got:\n%s", rendered)
	}
	if !strings.Contains(rendered, "1250 RPM") {
		t.Fatalf("expected fan RPM in CPU box, got:\n%s", rendered)
	}
}
