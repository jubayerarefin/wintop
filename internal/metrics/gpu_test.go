package metrics

import (
	"testing"
)

func TestCollectGPUs(t *testing.T) {
	gpus := collectGPUs()
	if len(gpus) == 0 {
		t.Log("No NVIDIA GPU detected or nvidia-smi unavailable")
		return
	}
	for i, g := range gpus {
		t.Logf("GPU %d: %s | Utilization: %.1f%% | VRAM: %d/%d MB (%.1f%%) | Fan: %.1f%% | Temp: %.1f°C",
			i, g.Name, g.UtilizationGPU, g.MemUsedMB, g.MemTotalMB, g.MemUsedPercent, g.FanSpeedPct, g.Temperature)
	}
}
