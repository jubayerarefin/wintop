package ui

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"wintop/internal/metrics"
)

type SortBy int

const (
	SortByCPU SortBy = iota
	SortByMem
	SortByPID
	SortByName
)

type TickMsg time.Time

type Model struct {
	collector *metrics.Collector
	theme     Theme

	width  int
	height int

	// Current snapshot & histories
	snapshot SystemSnapshotWrapper

	// Process navigation
	selectedProcIdx int
	procScrollOff   int
	sortBy          SortBy
	sortAsc         bool

	// Search / Filtering
	filterMode bool
	filterText string

	// Status / Notification
	statusMsg    string
	statusExpiry time.Time
}

type SystemSnapshotWrapper struct {
	snap        metrics.SystemSnapshot
	cpuHistory  []float64
	netRecvHist []float64
	netSentHist []float64
}

func NewModel(collector *metrics.Collector) Model {
	return Model{
		collector: collector,
		theme:     DefaultTheme(),
		sortBy:    SortByCPU,
		sortAsc:   false,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		tickCmd(),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case TickMsg:
		snap, cpuH, netR, netS := m.collector.GetSnapshot()
		m.snapshot = SystemSnapshotWrapper{
			snap:        snap,
			cpuHistory:  cpuH,
			netRecvHist: netR,
			netSentHist: netS,
		}
		if time.Now().After(m.statusExpiry) {
			m.statusMsg = ""
		}
		return m, tickCmd()

	case tea.KeyMsg:
		if m.filterMode {
			switch msg.Type {
			case tea.KeyEnter, tea.KeyEsc:
				m.filterMode = false
			case tea.KeyBackspace:
				if len(m.filterText) > 0 {
					m.filterText = m.filterText[:len(m.filterText)-1]
					m.selectedProcIdx = 0
				}
			case tea.KeyRunes:
				m.filterText += string(msg.Runes)
				m.selectedProcIdx = 0
			}
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			m.collector.Stop()
			return m, tea.Quit

		case "/":
			m.filterMode = true
			return m, nil

		case "c":
			// Clear filter
			m.filterText = ""
			m.selectedProcIdx = 0
			m.statusMsg = "Filter cleared"
			m.statusExpiry = time.Now().Add(2 * time.Second)

		case "s":
			// Cycle sort column
			m.sortBy = (m.sortBy + 1) % 4
			m.selectedProcIdx = 0

		case "r":
			// Toggle sort ascending/descending
			m.sortAsc = !m.sortAsc
			m.selectedProcIdx = 0

		case "up", "k":
			if m.selectedProcIdx > 0 {
				m.selectedProcIdx--
			}

		case "down", "j":
			procs := m.getFilteredProcesses()
			if m.selectedProcIdx < len(procs)-1 {
				m.selectedProcIdx++
			}

		case "pgup":
			m.selectedProcIdx -= 10
			if m.selectedProcIdx < 0 {
				m.selectedProcIdx = 0
			}

		case "pgdown":
			procs := m.getFilteredProcesses()
			m.selectedProcIdx += 10
			if m.selectedProcIdx >= len(procs) {
				m.selectedProcIdx = len(procs) - 1
			}

		case "K": // Shift+K: Kill selected process
			procs := m.getFilteredProcesses()
			if m.selectedProcIdx >= 0 && m.selectedProcIdx < len(procs) {
				target := procs[m.selectedProcIdx]
				err := killProcess(target.PID)
				if err != nil {
					m.statusMsg = fmt.Sprintf("Failed to kill PID %d: %v", target.PID, err)
				} else {
					m.statusMsg = fmt.Sprintf("Terminated PID %d (%s)", target.PID, target.Name)
				}
				m.statusExpiry = time.Now().Add(3 * time.Second)
			}
		}
	}

	return m, nil
}

func killProcess(pid int32) error {
	p, err := os.FindProcess(int(pid))
	if err != nil {
		return err
	}
	return p.Kill()
}

func (m Model) getFilteredProcesses() []metrics.ProcessInfo {
	var list []metrics.ProcessInfo
	filter := strings.ToLower(m.filterText)

	for _, p := range m.snapshot.snap.Processes {
		if filter == "" || strings.Contains(strings.ToLower(p.Name), filter) || strings.Contains(fmt.Sprintf("%d", p.PID), filter) {
			list = append(list, p)
		}
	}

	sort.Slice(list, func(i, j int) bool {
		var less bool
		switch m.sortBy {
		case SortByCPU:
			less = list[i].CPUPercent < list[j].CPUPercent
		case SortByMem:
			less = list[i].MemPercent < list[j].MemPercent
		case SortByPID:
			less = list[i].PID < list[j].PID
		case SortByName:
			less = strings.ToLower(list[i].Name) < strings.ToLower(list[j].Name)
		}
		if m.sortAsc {
			return less
		}
		return !less
	})

	return list
}

func (m Model) View() string {
	if m.width < 50 || m.height < 15 {
		return "Terminal window is too small for wintop. Please enlarge the window."
	}

	// Calculate sub-dimensions
	topBarH := 1
	bottomBarH := 1
	mainH := m.height - topBarH - bottomBarH - 2
	if mainH < 10 {
		mainH = 10
	}

	leftW := m.width / 2
	if leftW < 42 {
		leftW = 42
	}
	rightW := m.width - leftW - 2
	if rightW < 30 {
		rightW = 30
	}

	// Height breakdown for left column: CPU (40%), Mem & Disks (35%), Net (25%)
	cpuH := (mainH * 38) / 100
	memH := (mainH * 35) / 100
	netH := mainH - cpuH - memH
	if netH < 5 {
		netH = 5
	}

	// Render Header
	header := m.renderHeader()

	// Render Left Column Boxes
	cpuBox := m.renderCPUBox(leftW, cpuH)
	memBox := m.renderMemBox(leftW, memH)
	netBox := m.renderNetBox(leftW, netH)
	leftCol := lipgloss.JoinVertical(lipgloss.Left, cpuBox, memBox, netBox)

	// Render Right Column (Processes)
	procBox := m.renderProcBox(rightW, mainH)

	// Join Main View
	mainView := lipgloss.JoinHorizontal(lipgloss.Top, leftCol, procBox)

	// Render Footer
	footer := m.renderFooter()

	return lipgloss.JoinVertical(lipgloss.Left, header, mainView, footer)
}

func (m Model) renderHeader() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#bd93f9")).
		Render("⬢ wintop")

	infoStyle := lipgloss.NewStyle().Foreground(m.theme.TextMuted)
	valStyle := lipgloss.NewStyle().Foreground(m.theme.TextSecondary).Bold(true)

	osInfo := fmt.Sprintf("%s %s", infoStyle.Render("OS:"), valStyle.Render(m.snapshot.snap.OSName))
	hostInfo := fmt.Sprintf("%s %s", infoStyle.Render("Host:"), valStyle.Render(m.snapshot.snap.Hostname))
	uptimeInfo := fmt.Sprintf("%s %s", infoStyle.Render("Uptime:"), valStyle.Render(FormatDuration(m.snapshot.snap.Uptime)))

	content := fmt.Sprintf("%s   │   %s   │   %s   │   %s", title, osInfo, hostInfo, uptimeInfo)
	return lipgloss.NewStyle().
		Width(m.width).
		Background(lipgloss.Color("#1e1f29")).
		Padding(0, 1).
		Render(content)
}

func (m Model) renderFooter() string {
	sortName := "CPU%"
	switch m.sortBy {
	case SortByMem:
		sortName = "MEM%"
	case SortByPID:
		sortName = "PID"
	case SortByName:
		sortName = "NAME"
	}
	dir := "▼"
	if m.sortAsc {
		dir = "▲"
	}

	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#50fa7b")).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#f8f8f2"))

	var items []string
	items = append(items, fmt.Sprintf("%s %s", keyStyle.Render("[q]"), descStyle.Render("Quit")))
	items = append(items, fmt.Sprintf("%s %s", keyStyle.Render("[/]"), descStyle.Render("Filter")))
	items = append(items, fmt.Sprintf("%s %s", keyStyle.Render("[s]"), descStyle.Render(fmt.Sprintf("Sort: %s%s", sortName, dir))))
	items = append(items, fmt.Sprintf("%s %s", keyStyle.Render("[r]"), descStyle.Render("Reverse")))
	items = append(items, fmt.Sprintf("%s %s", keyStyle.Render("[Shift+K]"), descStyle.Render("Kill Proc")))

	if m.filterMode {
		items = []string{
			lipgloss.NewStyle().Foreground(lipgloss.Color("#ffb86c")).Bold(true).Render(fmt.Sprintf("FILTER: %s █ (Enter/Esc to finish)", m.filterText)),
		}
	} else if m.filterText != "" {
		items = append(items, fmt.Sprintf("%s %s", keyStyle.Render("[c]"), descStyle.Render(fmt.Sprintf("Clear '%s'", m.filterText))))
	}

	if m.statusMsg != "" {
		items = append(items, lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5555")).Bold(true).Render(m.statusMsg))
	}

	bar := strings.Join(items, "   ")
	return lipgloss.NewStyle().
		Width(m.width).
		Background(lipgloss.Color("#1e1f29")).
		Padding(0, 1).
		Render(bar)
}
