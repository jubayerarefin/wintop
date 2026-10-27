package metrics

import (
	"testing"
)

func TestCollectGPUs(t *testing.T) {
	gpus := collectGPUs()
	if len(gpus) == 0 {
		t.Log("No GPU detected")
		return
	}
	for i, g := range gpus {
		t.Logf("GPU %d: %s | Util: %.1f%% | VRAM: %d/%d MB (%.1f%%) | Fan: %.1f%% (%d RPM) | Temp: %.1f°C | Power: %.1f W / %.1f W",
			i, g.Name, g.UtilizationGPU, g.MemUsedMB, g.MemTotalMB, g.MemUsedPercent, g.FanSpeedPct, g.FanSpeedRPM, g.Temperature, g.PowerWatts, g.PowerLimitW)
	}
}
