package metrics

import (
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"github.com/yusufpapurcu/wmi"
)

type CPUPowerCollector struct {
	mu         sync.Mutex
	hQuery     uintptr
	hCounters  []uintptr
	openQuery  *syscall.Proc
	addCounter *syscall.Proc
	collect    *syscall.Proc
	getValue   *syscall.Proc
	closeQuery *syscall.Proc
	ready      bool
}

func NewCPUPowerCollector() *CPUPowerCollector {
	pdh, err := syscall.LoadDLL("pdh.dll")
	if err != nil {
		return &CPUPowerCollector{}
	}

	openQuery, err := pdh.FindProc("PdhOpenQueryW")
	if err != nil {
		return &CPUPowerCollector{}
	}
	expandPath, _ := pdh.FindProc("PdhExpandWildCardPathW")
	addCounter, err := pdh.FindProc("PdhAddEnglishCounterW")
	if err != nil {
		addCounter, _ = pdh.FindProc("PdhAddCounterW")
	}
	collect, _ := pdh.FindProc("PdhCollectQueryData")
	getValue, _ := pdh.FindProc("PdhGetFormattedCounterValue")
	closeQuery, _ := pdh.FindProc("PdhCloseQuery")

	c := &CPUPowerCollector{
		openQuery:  openQuery,
		addCounter: addCounter,
		collect:    collect,
		getValue:   getValue,
		closeQuery: closeQuery,
	}

	// Open persistent query
	var hQuery uintptr
	if ret, _, _ := openQuery.Call(0, 0, uintptr(unsafe.Pointer(&hQuery))); ret != 0 {
		return c
	}
	c.hQuery = hQuery

	// Discover counters
	var paths []string
	if expandPath != nil {
		wildcard, _ := syscall.UTF16PtrFromString(`\Energy Meter(*_pkg)\Power`)
		var pathListLen uint32
		expandPath.Call(0, uintptr(unsafe.Pointer(wildcard)), 0, uintptr(unsafe.Pointer(&pathListLen)), 0)
		if pathListLen > 0 {
			buf := make([]uint16, pathListLen)
			if ret, _, _ := expandPath.Call(0, uintptr(unsafe.Pointer(wildcard)), uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&pathListLen)), 0); ret == 0 {
				start := 0
				for i, v := range buf {
					if v == 0 {
						if i > start {
							paths = append(paths, syscall.UTF16ToString(buf[start:i]))
						}
						start = i + 1
					}
				}
			}
		}
	}

	// Fallback path if expansion was empty
	if len(paths) == 0 {
		paths = append(paths, `\Energy Meter(rapl_package0_pkg)\Power`)
	}

	for _, p := range paths {
		pPtr, _ := syscall.UTF16PtrFromString(p)
		var hC uintptr
		if r, _, _ := addCounter.Call(hQuery, uintptr(unsafe.Pointer(pPtr)), 0, uintptr(unsafe.Pointer(&hC))); r == 0 {
			c.hCounters = append(c.hCounters, hC)
		}
	}

	if len(c.hCounters) > 0 && collect != nil {
		collect.Call(hQuery)
		c.ready = true
	}

	return c
}

type pdhFmtValue struct {
	CStatus   uint32
	Pad       uint32
	DoubleVal float64
}

const pdhFmtDouble = 0x00000200

// QueryPower returns the current CPU package power in Watts.
func (c *CPUPowerCollector) QueryPower() float64 {
	if c == nil || !c.ready || c.collect == nil || c.getValue == nil {
		return 0
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	ret, _, _ := c.collect.Call(c.hQuery)
	if ret != 0 {
		return 0
	}

	var totalPowerWatts float64
	for _, hC := range c.hCounters {
		var val pdhFmtValue
		if r, _, _ := c.getValue.Call(hC, pdhFmtDouble, 0, uintptr(unsafe.Pointer(&val))); r == 0 {
			// Energy Meter reports in milliwatts
			if val.DoubleVal > 0 {
				totalPowerWatts += val.DoubleVal / 1000.0
			}
		}
	}

	return totalPowerWatts
}

func (c *CPUPowerCollector) Close() {
	if c == nil || !c.ready || c.closeQuery == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closeQuery.Call(c.hQuery)
	c.ready = false
}

type lhmSensor struct {
	Name       string
	SensorType string
	Value      float32
}

// QueryCPUFanRPM attempts to query CPU fan RPM from WMI if a hardware monitor service is active.
func QueryCPUFanRPM() uint32 {
	namespaces := []string{`root\LibreHardwareMonitor`, `root\OpenHardwareMonitor`}
	for _, ns := range namespaces {
		var sensors []lhmSensor
		query := "SELECT Name, SensorType, Value FROM Sensor WHERE SensorType = 'Fan'"
		err := wmi.QueryNamespace(query, &sensors, ns)
		if err == nil && len(sensors) > 0 {
			for _, s := range sensors {
				name := strings.ToLower(s.Name)
				if strings.Contains(name, "cpu") || strings.Contains(name, "fan #1") {
					if s.Value > 0 {
						return uint32(s.Value)
					}
				}
			}
			// If no specific CPU named fan, return the first fan
			if sensors[0].Value > 0 {
				return uint32(sensors[0].Value)
			}
		}
	}
	return 0
}
