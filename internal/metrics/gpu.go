package metrics

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// GPUInfo holds telemetry for a graphic processing unit.
type GPUInfo struct {
	Name           string
	UtilizationGPU float64 // %
	MemUsedMB      uint64  // MB
	MemTotalMB     uint64  // MB
	MemUsedPercent float64 // %
	FanSpeedPct    float64 // %
	Temperature    float64 // Celsius
}

func findNvidiaSmi() string {
	if p, err := exec.LookPath("nvidia-smi"); err == nil {
		return p
	}
	candidates := []string{
		filepath.Join(os.Getenv("SystemRoot"), "System32", "nvidia-smi.exe"),
		filepath.Join(os.Getenv("ProgramFiles"), "NVIDIA Corporation", "NVSMI", "nvidia-smi.exe"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}

func collectGPUs() []GPUInfo {
	smiPath := findNvidiaSmi()
	if smiPath == "" {
		return nil
	}

	cmd := exec.Command(smiPath,
		"--query-gpu=name,utilization.gpu,memory.used,memory.total,fan.speed,temperature.gpu",
		"--format=csv,noheader,nounits",
	)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	var gpus []GPUInfo
	for _, line := range lines {
		parts := strings.Split(line, ",")
		if len(parts) < 6 {
			continue
		}
		name := strings.TrimSpace(parts[0])
		utilGPU, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		memUsed, _ := strconv.ParseUint(strings.TrimSpace(parts[2]), 10, 64)
		memTotal, _ := strconv.ParseUint(strings.TrimSpace(parts[3]), 10, 64)
		fanSpeed, _ := strconv.ParseFloat(strings.TrimSpace(parts[4]), 64)
		temp, _ := strconv.ParseFloat(strings.TrimSpace(parts[5]), 64)

		var memPct float64
		if memTotal > 0 {
			memPct = (float64(memUsed) / float64(memTotal)) * 100.0
		}

		gpus = append(gpus, GPUInfo{
			Name:           name,
			UtilizationGPU: utilGPU,
			MemUsedMB:      memUsed,
			MemTotalMB:     memTotal,
			MemUsedPercent: memPct,
			FanSpeedPct:    fanSpeed,
			Temperature:    temp,
		})
	}
	return gpus
}
