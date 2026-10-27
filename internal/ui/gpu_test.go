package ui

import (
	"strings"
	"testing"

	"wintop/internal/metrics"
)

func TestGPUView(t *testing.T) {
	snap := metrics.SystemSnapshot{
		GPUs: []metrics.GPUInfo{
			{
				Name:           "NVIDIA GeForce RTX 4070 Ti SUPER",
				UtilizationGPU: 25.5,
				MemUsedMB:      4096,
				MemTotalMB:     16384,
				MemUsedPercent: 25.0,
				FanSpeedPct:    35.0,
				Temperature:    42.0,
			},
		},
	}

	m := Model{
		theme: DefaultTheme(),
		snapshot: SystemSnapshotWrapper{
			snap:       snap,
			gpuHistory: []float64{10, 15, 20, 25, 25.5},
		},
	}

	rendered := m.renderGPUBox(60, 10)
	t.Logf("Rendered GPU Box:\n%s", rendered)

	if !strings.Contains(rendered, "RTX 4070") {
		t.Fatalf("expected GPU model in box, got:\n%s", rendered)
	}
	if !strings.Contains(rendered, "PROC") || !strings.Contains(rendered, "VRAM") || !strings.Contains(rendered, "FAN") {
		t.Fatalf("expected PROC, VRAM, and FAN in box, got:\n%s", rendered)
	}
}
