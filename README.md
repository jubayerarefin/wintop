# wintop ⬢

A modern, fast, terminal-based resource and process monitor for Windows written entirely in **Go**, inspired by **btop**.

---

## ✨ Features

* **100% Pure Go**: No CGO, no Cygwin/WSL dependencies. Native Windows binary.
* **Braille Canvas Graphs**: High-density 2×4 dot Braille curves (`U+2800`–`U+28FF`) for real-time CPU, GPU, and Network bandwidth monitoring.
* **Live System Telemetry**:
  * **CPU**: Overall usage %, per-core real-time mini load bars, and model information.
  * **GPU**: GPU Processing / Utilization %, VRAM usage (Used / Total & %), Fan Speed %, Temperature (°C), and real-time Braille workload graph.
  * **Memory & Swap**: Physical RAM usage, Pagefile usage, and Drive-by-drive capacity meters.
  * **Network**: Real-time Download (RX) and Upload (TX) speed calculation with bandwidth history graphs.
  * **Processes**: Full process table with PID, Process Name, CPU%, Memory%, RSS memory size, and Thread count.
* **Interactive Controls**:
  * Process searching and real-time filtering (`/`).
  * Dynamic column sorting (`s` to cycle CPU / MEM / PID / NAME, `r` to reverse).
  * Smooth process scrolling (`↑`, `↓`, `PageUp`, `PageDown`).
  * Process termination (`Shift + K`).
* **Sleek Aesthetic**: btop/Dracula neon palette with rounded box borders and gradient progress meters.

---

## ⌨️ Keybindings

| Key | Action |
| :--- | :--- |
| `q` / `Ctrl+C` | Quit `wintop` |
| `↑` / `k` | Move selection up in process table |
| `↓` / `j` | Move selection down in process table |
| `PgUp` / `PgDn` | Scroll process table by 10 rows |
| `/` | Enter search / filter mode |
| `Esc` / `Enter` | Exit search / filter mode |
| `c` | Clear current search filter |
| `s` | Cycle sort column (`CPU%` → `MEM%` → `PID` → `NAME`) |
| `r` | Reverse sort order (Ascending / Descending) |
| `Shift + K` | Terminate the highlighted process |

---

## 🚀 Building & Running

### Prerequisites
* Go 1.21 or higher installed on Windows.

### Run Directly
```powershell
go run .
```

### Build Executable
```powershell
go build -ldflags="-s -w" -o wintop.exe .
.\wintop.exe
```

---

## 🏗️ Architecture

```
wintop/
├── main.go               # Application entrypoint & Bubble Tea lifecycle
├── go.mod / go.sum
└── internal/
    ├── metrics/          # Windows system, GPU & process collection engine
    │   ├── collector.go      # Multi-threaded polling loop & history buffers
    │   ├── collector_test.go # Windows telemetry verification tests
    │   ├── gpu.go            # GPU detection, utilization, VRAM & fan speed
    │   └── gpu_test.go       # GPU metrics unit tests
    └── ui/               # Bubble Tea TUI presentation layer
        ├── model.go          # State machine, key handlers & layout manager
        ├── theme.go          # Dracula/Tokyo-Night color palette & progress bars
        ├── braille.go        # 2x4 dot Braille canvas plotting engine
        ├── cpu_view.go       # CPU box & per-core mini meters
        ├── gpu_view.go       # GPU box, VRAM, Fan speed & Braille workload graph
        ├── mem_view.go       # RAM, Pagefile & Disk volume meters
        ├── net_view.go       # Network RX/TX rate meters & Braille graph
        └── proc_view.go      # Interactive process table
```
