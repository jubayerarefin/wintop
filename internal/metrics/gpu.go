package metrics

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

// GPUInfo holds telemetry for a graphic processing unit.
type GPUInfo struct {
	Name           string
	UtilizationGPU float64 // %
	MemUsedMB      uint64  // MB
	MemTotalMB     uint64  // MB
	MemUsedPercent float64 // %
	FanSpeedPct    float64 // %
	FanSpeedRPM    uint32  // RPM
	Temperature    float64 // Celsius
	PowerWatts     float64 // Watts
	PowerLimitW    float64 // Watts
}

type nvmlClient struct {
	dll                     *syscall.DLL
	getDeviceCount          *syscall.Proc
	getHandleByIndex        *syscall.Proc
	getName                 *syscall.Proc
	getUtilRates            *syscall.Proc
	getMemInfo              *syscall.Proc
	getFanSpeed             *syscall.Proc
	getFanSpeedRPM          *syscall.Proc
	getTemp                 *syscall.Proc
	getPowerUsage           *syscall.Proc
	getEnforcedPowerLimit   *syscall.Proc
	initialized             bool
}

type nvmlUtil struct {
	GPU    uint32
	Memory uint32
}

type nvmlMem struct {
	Total uint64
	Free  uint64
	Used  uint64
}

type nvmlFanSpeed struct {
	Version uint32
	Fan     uint32
	Speed   uint32
}

var (
	globalNVML     *nvmlClient
	globalNVMLOnce sync.Once
)

func initNVML() *nvmlClient {
	dllPath := filepath.Join(os.Getenv("SystemRoot"), "System32", "nvml.dll")
	dll, err := syscall.LoadDLL(dllPath)
	if err != nil {
		return nil
	}

	initProc, err := dll.FindProc("nvmlInit_v2")
	if err != nil {
		initProc, err = dll.FindProc("nvmlInit")
		if err != nil {
			dll.Release()
			return nil
		}
	}
	ret, _, _ := initProc.Call()
	if ret != 0 {
		dll.Release()
		return nil
	}

	getHandle, err := dll.FindProc("nvmlDeviceGetHandleByIndex_v2")
	if err != nil {
		getHandle, _ = dll.FindProc("nvmlDeviceGetHandleByIndex")
	}

	client := &nvmlClient{
		dll:                   dll,
		getDeviceCount:        findProcOptional(dll, "nvmlDeviceGetCount_v2", "nvmlDeviceGetCount"),
		getHandleByIndex:      getHandle,
		getName:               findProcOptional(dll, "nvmlDeviceGetName"),
		getUtilRates:          findProcOptional(dll, "nvmlDeviceGetUtilizationRates"),
		getMemInfo:            findProcOptional(dll, "nvmlDeviceGetMemoryInfo"),
		getFanSpeed:           findProcOptional(dll, "nvmlDeviceGetFanSpeed"),
		getFanSpeedRPM:        findProcOptional(dll, "nvmlDeviceGetFanSpeedRPM"),
		getTemp:               findProcOptional(dll, "nvmlDeviceGetTemperature"),
		getPowerUsage:         findProcOptional(dll, "nvmlDeviceGetPowerUsage"),
		getEnforcedPowerLimit: findProcOptional(dll, "nvmlDeviceGetEnforcedPowerLimit"),
		initialized:           true,
	}

	return client
}

func findProcOptional(dll *syscall.DLL, names ...string) *syscall.Proc {
	for _, n := range names {
		if p, err := dll.FindProc(n); err == nil {
			return p
		}
	}
	return nil
}

func collectGPUsDirectNVML() []GPUInfo {
	globalNVMLOnce.Do(func() {
		globalNVML = initNVML()
	})
	if globalNVML == nil || !globalNVML.initialized || globalNVML.getHandleByIndex == nil {
		return nil
	}

	var count uint32 = 1
	if globalNVML.getDeviceCount != nil {
		globalNVML.getDeviceCount.Call(uintptr(unsafe.Pointer(&count)))
	}
	if count == 0 {
		count = 1
	}

	var gpus []GPUInfo
	for i := uint32(0); i < count; i++ {
		var dev uintptr
		ret, _, _ := globalNVML.getHandleByIndex.Call(uintptr(i), uintptr(unsafe.Pointer(&dev)))
		if ret != 0 {
			continue
		}

		var info GPUInfo

		// Name
		if globalNVML.getName != nil {
			var nameBuf [96]byte
			globalNVML.getName.Call(dev, uintptr(unsafe.Pointer(&nameBuf[0])), 96)
			info.Name = strings.TrimRight(string(nameBuf[:]), "\x00 ")
		}

		// Utilization
		if globalNVML.getUtilRates != nil {
			var u nvmlUtil
			ret, _, _ := globalNVML.getUtilRates.Call(dev, uintptr(unsafe.Pointer(&u)))
			if ret == 0 {
				info.UtilizationGPU = float64(u.GPU)
			}
		}

		// Memory
		if globalNVML.getMemInfo != nil {
			var m nvmlMem
			ret, _, _ := globalNVML.getMemInfo.Call(dev, uintptr(unsafe.Pointer(&m)))
			if ret == 0 && m.Total > 0 {
				info.MemTotalMB = m.Total / (1024 * 1024)
				info.MemUsedMB = m.Used / (1024 * 1024)
				info.MemUsedPercent = (float64(m.Used) / float64(m.Total)) * 100.0
			}
		}

		// Fan Speed %
		if globalNVML.getFanSpeed != nil {
			var speed uint32
			ret, _, _ := globalNVML.getFanSpeed.Call(dev, uintptr(unsafe.Pointer(&speed)))
			if ret == 0 {
				info.FanSpeedPct = float64(speed)
			}
		}

		// Fan Speed RPM
		if globalNVML.getFanSpeedRPM != nil {
			var fan nvmlFanSpeed
			fan.Version = uint32(unsafe.Sizeof(nvmlFanSpeed{})) | (1 << 24)
			fan.Fan = 0
			ret, _, _ := globalNVML.getFanSpeedRPM.Call(dev, uintptr(unsafe.Pointer(&fan)))
			if ret == 0 {
				info.FanSpeedRPM = fan.Speed
			}
		}

		// Temperature
		if globalNVML.getTemp != nil {
			var temp uint32
			ret, _, _ := globalNVML.getTemp.Call(dev, 0, uintptr(unsafe.Pointer(&temp)))
			if ret == 0 {
				info.Temperature = float64(temp)
			}
		}

		// Power Usage (milliwatts)
		if globalNVML.getPowerUsage != nil {
			var pwr uint32
			ret, _, _ := globalNVML.getPowerUsage.Call(dev, uintptr(unsafe.Pointer(&pwr)))
			if ret == 0 {
				info.PowerWatts = float64(pwr) / 1000.0
			}
		}

		// Power Limit (milliwatts)
		if globalNVML.getEnforcedPowerLimit != nil {
			var lim uint32
			ret, _, _ := globalNVML.getEnforcedPowerLimit.Call(dev, uintptr(unsafe.Pointer(&lim)))
			if ret == 0 {
				info.PowerLimitW = float64(lim) / 1000.0
			}
		}

		gpus = append(gpus, info)
	}

	return gpus
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

func collectGPUsFallback() []GPUInfo {
	smiPath := findNvidiaSmi()
	if smiPath == "" {
		return nil
	}

	cmd := exec.Command(smiPath,
		"--query-gpu=name,utilization.gpu,memory.used,memory.total,fan.speed,temperature.gpu,power.draw,enforced.power.limit",
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

		var pwrWatts, pwrLimit float64
		if len(parts) >= 7 {
			pwrWatts, _ = strconv.ParseFloat(strings.TrimSpace(parts[6]), 64)
		}
		if len(parts) >= 8 {
			pwrLimit, _ = strconv.ParseFloat(strings.TrimSpace(parts[7]), 64)
		}

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
			PowerWatts:     pwrWatts,
			PowerLimitW:    pwrLimit,
		})
	}
	return gpus
}

func collectGPUs() []GPUInfo {
	if gpus := collectGPUsDirectNVML(); len(gpus) > 0 {
		return gpus
	}
	return collectGPUsFallback()
}
