package ui

import (
	"strings"
	"testing"

	"wintop/internal/metrics"
)

func TestCPUViewPowerFan(t *testing.T) {
	// 8 physical cores, 16 threads
	coresPct := make([]float64, 16)
	for i := range coresPct {
		coresPct[i] = float64((i + 1) * 5)
	}

	snap := metrics.SystemSnapshot{
		CPUName:          "AMD Ryzen 7 5700X 8-Core Processor",
		CPUPhysicalCores: 8,
		CPULogicalCores:  16,
		CPUTotalPct:      15.4,
		CPUCoresPct:      coresPct,
		CPUPowerWatts:    52.8,
		CPUFanRPM:        1250,
	}

	m := Model{
		theme: DefaultTheme(),
		snapshot: SystemSnapshotWrapper{
			snap:       snap,
			cpuHistory: []float64{10, 12, 14, 15.4},
		},
	}

	rendered := m.renderCPUBox(60, 11)
	t.Logf("Rendered CPU Box (with 16 threads):\n%s", rendered)

	if !strings.Contains(rendered, "(8C/16T)") {
		t.Fatalf("expected (8C/16T) in CPU box, got:\n%s", rendered)
	}
	if !strings.Contains(rendered, "52.8 W") {
		t.Fatalf("expected power in CPU box, got:\n%s", rendered)
	}
	if !strings.Contains(rendered, "1250 RPM") {
		t.Fatalf("expected fan RPM in CPU box, got:\n%s", rendered)
	}

	// Verify all 16 threads appear (C00 through C15)
	for i := 0; i < 16; i++ {
		tag := strings.TrimSpace(formatCoreBar(i, coresPct[i], 2, DefaultTheme()))
		// Check for core index C00, C01...
		cPrefix := strings.Split(tag, " ")[0]
		if !strings.Contains(rendered, cPrefix) {
			t.Fatalf("expected core %s in rendered output", cPrefix)
		}
	}
}

func TestCPUViewFanNA(t *testing.T) {
	snap := metrics.SystemSnapshot{
		CPUName:          "AMD Ryzen 7 5700X 8-Core Processor",
		CPUPhysicalCores: 8,
		CPULogicalCores:  16,
		CPUTotalPct:      15.4,
		CPUFanRPM:        0, // No sensor running
	}

	m := Model{
		theme: DefaultTheme(),
		snapshot: SystemSnapshotWrapper{
			snap: snap,
		},
	}

	rendered := m.renderCPUBox(60, 8)
	t.Logf("Rendered CPU Box (Fan N/A):\n%s", rendered)

	if !strings.Contains(rendered, "FAN: N/A") {
		t.Fatalf("expected 'FAN: N/A' in CPU box, got:\n%s", rendered)
	}
}
