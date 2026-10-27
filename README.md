# wintop ⬢

A fast, modern terminal resource and process monitor for **Windows** written entirely in **100% pure Go**, inspired by the visual design and responsive ergonomics of **btop**.

---

## 📖 Background & Design Decisions

### 1. Why a Native Go Rewrite Instead of Porting btop's C++ Code?
`btop` is widely loved on Linux, macOS, and FreeBSD for its sleek responsive boxes and high-density Braille charts. However, porting `btop` directly to Windows was historically blocked by two major architectural hurdles:
* **Operating System Telemetry**: `btop`'s data gathering engine relies strictly on POSIX interfaces: reading `/proc` filesystems on Linux, `sysctl`/`kvm` on FreeBSD, and Mach kernel APIs (`libproc`) on macOS. It contains virtually no native Windows collection logic.
* **C++20 to Go Transpilation**: Automated transpilers only support simple C89/C99; modern C++20 with template metaprogramming, STL containers, and `std::jthread` cannot be mechanically transpiled to idiomatic Go.

`wintop` solves this by reimplementing btop's UI philosophy, themes, and Braille mathematics from scratch in pure Go, backed by a native Windows telemetry engine using Windows APIs, PDH performance counters, and NVML.

---

### 2. Engineering Architecture & Telemetry Insights

#### ⬡ High-Density Braille Canvas (`internal/ui/braille.go`)
Terminal cells are traditionally 1×1 character blocks. `wintop` implements a 2×4 dot sub-cell canvas using Unicode Braille patterns (`U+2800` through `U+28FF`):
* Each terminal character cell yields **2 horizontal columns** and **4 vertical rows** of resolution.
* Bitmask coordinates are mapped according to the Braille dot standard:
  $$\text{Row 0: }(0x01, 0x08), \quad \text{Row 1: }(0x02, 0x10), \quad \text{Row 2: }(0x04, 0x20), \quad \text{Row 3: }(0x40, 0x80)$$
* Renders real-time filled area graphs for CPU load, GPU compute utilization, and Network throughput.

#### ⬡ GPU Telemetry: Direct In-Process NVML (`internal/metrics/gpu.go`)
* **Sub-millisecond latency (< 1ms)**: Rather than spawning expensive subprocesses on every tick, `wintop` dynamically links to `C:\Windows\System32\nvml.dll` directly in-process via `syscall.LoadDLL`.
* **Telemetry Gathered**:
  * **Compute Utilization**: Active core load percentage.
  * **Video RAM (VRAM)**: Dedicated frame buffer usage (`Used / Total` in MB/GB and percentage).
  * **Fan Speed (Both % and RPM)**: Retrieves intended fan duty percentage via `nvmlDeviceGetFanSpeed` and real hardware rotation speed via `nvmlDeviceGetFanSpeedRPM`.
  * **Power Consumption**: Live package wattage via `nvmlDeviceGetPowerUsage` and board ceiling limits via `nvmlDeviceGetEnforcedPowerLimit`.
  * **Temperature**: Core thermal sensor in Celsius.
* **Graceful Fallback**: If `nvml.dll` is unavailable or on non-standard setups, falls back automatically to `nvidia-smi` CLI parsing; if no dedicated GPU exists, the panel automatically collapses to allocate space to CPU and Memory.

#### ⬡ CPU Telemetry: RAPL Power & Topology (`internal/metrics/cpu_power.go`)
* **Accurate Physical vs. Logical Topology**: Distinguishes physical hardware cores (`cpu.Counts(false)`) from logical SMT/HyperThreading threads (`cpu.Counts(true)`), rendering clean topological badges like `(8C/16T)`.
* **4-Column Dynamic Core Grid**: Instead of truncating multi-threaded CPUs when vertical space is restricted, `wintop` packs threads across 4 columns (`C00` through `C15`), displaying all 16+ logical processors in just 4 terminal rows.
* **Package Power Consumption (RAPL)**: Interfaces with the Windows Performance Data Helper (PDH) subsystem (`\Energy Meter(*_pkg)\Power`) to read Intel and AMD Ryzen CPU Package Power in real-time Watts without requiring kernel-level driver installation.
* **CPU Fan Speed & Motherboard Sensors**: Desktop motherboard fan headers (`CPU_FAN`) are wired to LPC Super I/O chips (e.g. ITE, Nuvoton) which Windows restricts from unprivileged user-mode access. `wintop` dynamically queries WMI sensor namespaces (`root\LibreHardwareMonitor`, `root\OpenHardwareMonitor`) when a hardware monitor is running to surface tachometer RPM, falling back cleanly to `FAN: N/A` when sensors are idle.

#### ⬡ Process Inspection (`internal/ui/proc_view.go`)
* **RSS (Resident Set Size)**: Real physical RAM currently occupied by the process (Windows Working Set), filtering out uncommitted virtual memory.
* **THR (Thread Count)**: Active execution threads dispatched across CPU cores.
* **Interactive Navigation**: Live search/filter (`/`), sorting by CPU%, MEM%, PID, or NAME (`s`), and termination signals (`Shift + K`).

---

## ✨ Features Summary

| Component | Telemetry & Visuals |
| :--- | :--- |
| **CPU Panel** | Model, Physical/Logical topology `(8C/16T)`, Package Power (Watts), Fan RPM, Total Load %, Braille area graph, and 4-column multi-core load bars. |
| **GPU Panel** | Model, Core Temp (°C), Processing Load %, Braille history graph, VRAM bar (`Used / Total`), Fan Speed (`%` and `RPM`), and Power (`Watts / Limit`). |
| **Memory / Disks** | RAM usage meter (`Used / Total / %`), Pagefile/Swap meter, and Drive partitions (`C:`, `D:`) with capacity and usage bars. |
| **Network Panel** | Live Download (RX) and Upload (TX) bitrates, total transfer volume, and real-time bandwidth Braille graph. |
| **Process Table** | Interactive table with PID, Process Name, CPU%, MEM%, RSS memory, and Thread count. |

---

## ⌨️ Keybindings

| Key | Action |
| :--- | :--- |
| `q` / `Ctrl+C` | Quit `wintop` |
| `↑` / `k` | Move selection up in process table |
| `↓` / `j` | Move selection down in process table |
| `PgUp` / `PgDn` | Scroll process table by 10 rows |
| `/` | Enter search / filter mode (type query) |
| `Esc` / `Enter` | Exit search / filter mode |
| `c` | Clear current search filter |
| `s` | Cycle sort column (`CPU%` → `MEM%` → `PID` → `NAME`) |
| `r` | Reverse sort order (Ascending / Descending) |
| `Shift + K` | Terminate highlighted process |

---

## 🚀 Building & Running

### Prerequisites
* Windows 10/11 (64-bit)
* Go 1.21 or higher

### Run from Source
```powershell
go run .
```

### Compile Release Binary
```powershell
go build -ldflags="-s -w" -o wintop.exe .
.\wintop.exe
```

---

## 🏗️ Project Architecture

```
wintop/
├── main.go                       # Application entrypoint & Bubble Tea lifecycle
├── go.mod / go.sum
└── internal/
    ├── metrics/                  # Windows system, CPU, GPU & process collection
    │   ├── collector.go          # Multi-threaded polling loop & history buffers
    │   ├── collector_test.go     # Telemetry & core count verification tests
    │   ├── cpu_power.go          # Windows PDH Energy Meter (RAPL) & WMI fan query
    │   ├── cpu_power_test.go     # CPU power collector tests
    │   ├── gpu.go                # Native NVML in-process telemetry (RPM, Watts, VRAM)
    │   └── gpu_test.go           # GPU metrics unit tests
    └── ui/                       # Bubble Tea TUI presentation layer
        ├── model.go              # State machine, key handlers & layout manager
        ├── theme.go              # Dracula/btop color palette & progress bars
        ├── helpers.go            # Formatting & string truncation utilities
        ├── braille.go            # 2×4 dot Braille canvas plotting engine
        ├── braille_test.go       # Braille dot mapping unit tests
        ├── cpu_view.go           # CPU box: 4-column core grid, Power (Watts), Fan RPM
        ├── cpu_test.go           # CPU topology and power/fan rendering tests
        ├── gpu_view.go           # GPU box: VRAM, Fan (RPM + %), Power (W) & Braille graph
        ├── gpu_test.go           # GPU rendering & RPM/Watts unit tests
        ├── mem_view.go           # RAM, Pagefile & Disk volume meters
        ├── net_view.go           # Network RX/TX rate meters & Braille graph
        └── proc_view.go          # Filterable, sortable process table
```

---

## 📄 License
Licensed under the [GNU Affero General Public License v3.0](LICENSE).
