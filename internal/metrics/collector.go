package metrics

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

// ProcessInfo holds snapshot data for a running process.
type ProcessInfo struct {
	PID        int32
	Name       string
	CPUPercent float64
	MemPercent float32
	MemRSS     uint64 // bytes
	NumThreads int32
	Status     string
	Username   string
}

// DiskInfo holds disk drive stats.
type DiskInfo struct {
	Mountpoint  string
	Fstype      string
	Total       uint64
	Used        uint64
	Free        uint64
	UsedPercent float64
}

// SystemSnapshot contains a point-in-time snapshot of the system.
type SystemSnapshot struct {
	Uptime      uint64
	Hostname    string
	OSName      string
	CPUName     string
	CPUCores    int
	CPUTotalPct float64
	CPUCoresPct []float64

	MemTotal       uint64
	MemUsed        uint64
	MemAvailable   uint64
	MemUsedPercent float64

	SwapTotal       uint64
	SwapUsed        uint64
	SwapUsedPercent float64

	Disks []DiskInfo

	NetBytesRecvPerSec float64
	NetBytesSentPerSec float64
	NetTotalRecv       uint64
	NetTotalSent       uint64

	GPUs []GPUInfo

	Processes []ProcessInfo
}

// Collector continuously monitors system metrics and provides snapshots.
type Collector struct {
	mu           sync.RWMutex
	snapshot     SystemSnapshot
	cpuHistory   []float64
	netRecvHist  []float64
	netSentHist  []float64
	gpuHistory   []float64
	historyLimit int

	prevNetTime time.Time
	prevNetRecv uint64
	prevNetSent uint64

	stopChan chan struct{}
}

// NewCollector creates a collector with specified history length.
func NewCollector(historyLimit int) *Collector {
	if historyLimit < 30 {
		historyLimit = 120
	}
	return &Collector{
		historyLimit: historyLimit,
		cpuHistory:   make([]float64, 0, historyLimit),
		netRecvHist:  make([]float64, 0, historyLimit),
		netSentHist:  make([]float64, 0, historyLimit),
		gpuHistory:   make([]float64, 0, historyLimit),
		stopChan:     make(chan struct{}),
	}
}

// Start launches the polling loop at the specified interval.
func (c *Collector) Start(interval time.Duration) {
	// First immediate collection
	c.collect()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-c.stopChan:
				return
			case <-ticker.C:
				c.collect()
			}
		}
	}()
}

// Stop terminates the polling goroutine.
func (c *Collector) Stop() {
	select {
	case <-c.stopChan:
	default:
		close(c.stopChan)
	}
}

// GetSnapshot returns a copy of the current snapshot and history.
func (c *Collector) GetSnapshot() (SystemSnapshot, []float64, []float64, []float64, []float64) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cpuCopy := make([]float64, len(c.cpuHistory))
	copy(cpuCopy, c.cpuHistory)

	netRecvCopy := make([]float64, len(c.netRecvHist))
	copy(netRecvCopy, c.netRecvHist)

	netSentCopy := make([]float64, len(c.netSentHist))
	copy(netSentCopy, c.netSentHist)

	gpuCopy := make([]float64, len(c.gpuHistory))
	copy(gpuCopy, c.gpuHistory)

	return c.snapshot, cpuCopy, netRecvCopy, netSentCopy, gpuCopy
}

func (c *Collector) collect() {
	var snap SystemSnapshot

	// Host info
	if hInfo, err := host.Info(); err == nil {
		snap.Uptime = hInfo.Uptime
		snap.Hostname = hInfo.Hostname
		snap.OSName = fmt.Sprintf("%s %s", hInfo.Platform, hInfo.PlatformVersion)
	}

	// CPU info
	if cInfos, err := cpu.Info(); err == nil && len(cInfos) > 0 {
		snap.CPUName = cInfos[0].ModelName
	}
	if logicalCores, err := cpu.Counts(true); err == nil && logicalCores > 0 {
		snap.CPUCores = logicalCores
	}

	// CPU Percentages (total and per-core)
	if totalPcts, err := cpu.Percent(0, false); err == nil && len(totalPcts) > 0 {
		snap.CPUTotalPct = totalPcts[0]
	}
	if corePcts, err := cpu.Percent(0, true); err == nil {
		snap.CPUCoresPct = corePcts
		if snap.CPUCores == 0 {
			snap.CPUCores = len(corePcts)
		}
	}

	// Memory & Swap
	if vMem, err := mem.VirtualMemory(); err == nil {
		snap.MemTotal = vMem.Total
		snap.MemUsed = vMem.Used
		snap.MemAvailable = vMem.Available
		snap.MemUsedPercent = vMem.UsedPercent
	}
	if sMem, err := mem.SwapMemory(); err == nil {
		snap.SwapTotal = sMem.Total
		snap.SwapUsed = sMem.Used
		snap.SwapUsedPercent = sMem.UsedPercent
	}

	// Disks
	if parts, err := disk.Partitions(false); err == nil {
		for _, p := range parts {
			if strings.HasPrefix(p.Device, "CD-ROM") {
				continue
			}
			if u, err := disk.Usage(p.Mountpoint); err == nil && u.Total > 0 {
				snap.Disks = append(snap.Disks, DiskInfo{
					Mountpoint:  p.Mountpoint,
					Fstype:      p.Fstype,
					Total:       u.Total,
					Used:        u.Used,
					Free:        u.Free,
					UsedPercent: u.UsedPercent,
				})
			}
		}
	}

	// Network
	now := time.Now()
	if netCounters, err := net.IOCounters(false); err == nil && len(netCounters) > 0 {
		currentRecv := netCounters[0].BytesRecv
		currentSent := netCounters[0].BytesSent
		snap.NetTotalRecv = currentRecv
		snap.NetTotalSent = currentSent

		if !c.prevNetTime.IsZero() {
			dt := now.Sub(c.prevNetTime).Seconds()
			if dt > 0 {
				if currentRecv >= c.prevNetRecv {
					snap.NetBytesRecvPerSec = float64(currentRecv-c.prevNetRecv) / dt
				}
				if currentSent >= c.prevNetSent {
					snap.NetBytesSentPerSec = float64(currentSent-c.prevNetSent) / dt
				}
			}
		}
		c.prevNetTime = now
		c.prevNetRecv = currentRecv
		c.prevNetSent = currentSent
	}

	// GPUs
	snap.GPUs = collectGPUs()

	// Processes
	snap.Processes = c.collectProcesses()

	// Update collector state
	c.mu.Lock()
	c.snapshot = snap

	c.cpuHistory = append(c.cpuHistory, snap.CPUTotalPct)
	if len(c.cpuHistory) > c.historyLimit {
		c.cpuHistory = c.cpuHistory[1:]
	}

	c.netRecvHist = append(c.netRecvHist, snap.NetBytesRecvPerSec)
	if len(c.netRecvHist) > c.historyLimit {
		c.netRecvHist = c.netRecvHist[1:]
	}

	c.netSentHist = append(c.netSentHist, snap.NetBytesSentPerSec)
	if len(c.netSentHist) > c.historyLimit {
		c.netSentHist = c.netSentHist[1:]
	}

	var gpuPct float64
	if len(snap.GPUs) > 0 {
		gpuPct = snap.GPUs[0].UtilizationGPU
	}
	c.gpuHistory = append(c.gpuHistory, gpuPct)
	if len(c.gpuHistory) > c.historyLimit {
		c.gpuHistory = c.gpuHistory[1:]
	}
	c.mu.Unlock()
}

func (c *Collector) collectProcesses() []ProcessInfo {
	procs, err := process.Processes()
	if err != nil {
		return nil
	}

	var res []ProcessInfo
	for _, p := range procs {
		name, err := p.Name()
		if err != nil || name == "" {
			continue
		}

		cpuPct, _ := p.CPUPercent()
		memPct, _ := p.MemoryPercent()
		var rss uint64
		if memInfo, err := p.MemoryInfo(); err == nil && memInfo != nil {
			rss = memInfo.RSS
		}
		numThreads, _ := p.NumThreads()

		res = append(res, ProcessInfo{
			PID:        p.Pid,
			Name:       name,
			CPUPercent: cpuPct,
			MemPercent: memPct,
			MemRSS:     rss,
			NumThreads: numThreads,
		})
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].CPUPercent > res[j].CPUPercent
	})

	return res
}
