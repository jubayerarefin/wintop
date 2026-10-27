package metrics

import (
	"testing"
	"time"
)

func TestMetricsCollector(t *testing.T) {
	c := NewCollector(30)
	c.collect()

	snap, cpuHist, _, _ := c.GetSnapshot()
	if snap.OSName == "" {
		t.Log("Warning: OSName is empty")
	}
	t.Logf("Snapshot collected successfully: OS=%s, Host=%s, CPU=%s, Cores=%d, TotalCPU=%.1f%%, RAM=%.1f%%, Procs=%d",
		snap.OSName, snap.Hostname, snap.CPUName, snap.CPUCores, snap.CPUTotalPct, snap.MemUsedPercent, len(snap.Processes))

	if len(cpuHist) == 0 {
		t.Fatal("expected at least one cpu history point")
	}

	time.Sleep(100 * time.Millisecond)
}
