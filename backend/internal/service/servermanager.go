package service

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bladekkk/mcap/backend/internal/config"
)

type ServerStatus string

const (
	StatusStopped   ServerStatus = "stopped"
	StatusStarting  ServerStatus = "starting"
	StatusRunning   ServerStatus = "running"
	StatusStopping  ServerStatus = "stopping"
	StatusCrashed   ServerStatus = "crashed"
)

type ServerProcess struct {
	ID           string
	Name         string
	Cmd          *exec.Cmd
	StdoutReader *bufio.Reader
	StderrReader *bufio.Reader
	Status       ServerStatus
	StartTime    time.Time
	Lock         sync.Mutex
}

type ServerManager struct {
	cfg     *config.Config
	processes map[string]*ServerProcess
	mu      sync.RWMutex
}

func NewServerManager(cfg *config.Config) *ServerManager {
	return &ServerManager{
		cfg:       cfg,
		processes: make(map[string]*ServerProcess),
	}
}

func (sm *ServerManager) GetProcess(serverID string) (*ServerProcess, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	p, ok := sm.processes[serverID]
	return p, ok
}

func (sm *ServerManager) CreateProcess(serverID, name, jarPath, javaPath string, xmsMB, xmxMB int, affinity string) *ServerProcess {
	cmdArgs := sm.buildJavaArgs(javaPath, xmsMB, xmxMB, jarPath)
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Dir = config.GetServerDataDir(sm.cfg.DataDir, serverID)
	cmd.Stdout = nil
	cmd.Stderr = nil

	sm.mu.Lock()
	defer sm.mu.Unlock()

	sp := &ServerProcess{
		ID:        serverID,
		Name:      name,
		Cmd:       cmd,
		Status:    StatusStopped,
		StartTime: time.Time{},
	}
	sm.processes[serverID] = sp
	return sp
}

func (sm *ServerManager) buildJavaArgs(javaPath string, xmsMB, xmxMB int, jarPath string) []string {
	args := []string{
		javaPath,
		"-Xms" + fmt.Sprintf("%dM", xmsMB),
		"-Xmx" + fmt.Sprintf("%dM", xmxMB),
		"-XX:+UseG1GC",
		"-XX:+ParallelRefProcEnabled",
		"-XX:MaxGCPauseMillis=200",
		"-XX:+UnlockExperimentalVMOptions",
		"-XX:G1NewSizePercent=5",
		"-XX:G1MaxNewSizePercent=15",
		"-XX:G1HeapRegionSize=16M",
		"-XX:G1ReservePercent=20",
		"-XX:G1HeapWastePercent=5",
		"-XX:G1MixedGCCountTarget=4",
		"-XX:InitiatingHeapOccupancyPercent=15",
		"-XX:G1MixedGCLiveThresholdPercent=90",
		"-XX:G1RSetUpdatingPauseTimePercent=5",
		"-XX:SurvivorRatio=32",
		"-XX:+ParallelGCThreads=4",
		"-XX:ConcGCThreads=4",
		"-XX:G1HeapWastePercent=5",
		"-XX:MaxTenuringThreshold=1",
		"-Dfml.ignoreInvalidMinecraftCertificates=true",
		"-Dfml.ignorePatchDiscrepancies=true",
		"-jar",
		jarPath,
		"nogui",
	}
	return args
}

func (sm *ServerManager) Start(serverID string) error {
	sm.mu.Lock()
	sp, ok := sm.processes[serverID]
	if !ok {
		sm.mu.Unlock()
		return fmt.Errorf("server not found: %s", serverID)
	}
	if sp.Status == StatusRunning {
		sm.mu.Unlock()
		return fmt.Errorf("server already running")
	}
	sp.Status = StatusStarting
	sm.mu.Unlock()

	// Получаем настройки из БД
	// В реальном коде загружаем из БД, здесь упрощено

	err := sp.Cmd.Start()
	if err != nil {
		sm.mu.Lock()
		sp.Status = StatusCrashed
		sm.mu.Unlock()
		return fmt.Errorf("failed to start: %w", err)
	}

	sm.mu.Lock()
	sp.Status = StatusRunning
	sp.StartTime = time.Now()
	sm.mu.Unlock()

	return nil
}

func (sm *ServerManager) Stop(serverID string, timeout time.Duration) error {
	sm.mu.Lock()
	sp, ok := sm.processes[serverID]
	if !ok {
		sm.mu.Unlock()
		return fmt.Errorf("server not found: %s", serverID)
	}
	if sp.Status != StatusRunning {
		sm.mu.Unlock()
		return fmt.Errorf("server is not running")
	}
	sp.Status = StatusStopping
	sm.mu.Unlock()

	// Пытаемся остановить gracefully через RCON (или signal)
	// Для упрощения — посылаем SIGTERM
	if sp.Cmd.Process != nil {
		if err := sp.Cmd.Process.Signal(syscall.SIGTERM); err != nil {
			return fmt.Errorf("failed to send SIGTERM: %w", err)
		}
	}

	// Ждем завершения
	done := make(chan error, 1)
	go func() {
		done <- sp.Cmd.Wait()
	}()

	select {
	case <-time.After(timeout):
		sm.mu.Lock()
		if sp.Cmd.Process != nil {
			_ = sp.Cmd.Process.Kill()
		}
		sm.mu.Unlock()
		return fmt.Errorf("server did not stop in time")
	case err := <-done:
		sm.mu.Lock()
		sp.Status = StatusStopped
		sm.mu.Unlock()
		return err
	}
}

func (sm *ServerManager) Restart(serverID string) error {
	if err := sm.Stop(serverID, 30*time.Second); err != nil {
		return err
	}
	return sm.Start(serverID)
}

func (sm *ServerManager) GetMetrics(serverID string) (*ServerMetrics, error) {
	sm.mu.RLock()
	sp, ok := sm.processes[serverID]
	if !ok {
		sm.mu.RUnlock()
		return nil, fmt.Errorf("server not found: %s", serverID)
	}
	if sp.Status != StatusRunning {
		sm.mu.RUnlock()
		return &ServerMetrics{Status: string(sp.Status)}, nil
	}
	sm.mu.RUnlock()

	// Получаем метрики процесса через ps или procfs
	return sm.getProcessMetrics(sp.Cmd.Process.Pid)
}

type ServerMetrics struct {
	PID           int     `json:"pid"`
	Status        string  `json:"status"`
	StartTime     string  `json:"start_time"`
	UptimeSeconds int64   `json:"uptime_seconds"`
	RSSMB         int     `json:"rss_mb"`
	CPUPercent    float64 `json:"cpu_percent"`
	TPS           float64 `json:"tps"`
	MSPT          float64 `json:"mspt"`
	OnlinePlayers int     `json:"online_players"`
}

func (sm *ServerManager) getProcessMetrics(pid int) (*ServerMetrics, error) {
	// Для Linux — читаем из /proc
	if runtime.GOOS == "linux" {
		return getLinuxMetrics(pid)
	}
	// Для Windows/macOS — упрощенная реализация
	return &ServerMetrics{
		PID:          pid,
		Status:       "unknown",
		RSSMB:        0,
		CPUPercent:   0,
		TPS:          20,
		MSPT:         50,
		OnlinePlayers: 0,
	}, nil
}

func getLinuxMetrics(pid int) (*ServerMetrics, error) {
	// Читаем /proc/[pid]/statm и /proc/[pid]/stat
	statsPath := fmt.Sprintf("/proc/%d/stat", pid)
	statmPath := fmt.Sprintf("/proc/%d/statm", pid)

	stats, err := os.ReadFile(statsPath)
	if err != nil {
		return nil, err
	}
	statm, err := os.ReadFile(statmPath)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(string(stats), " ")
	if len(parts) < 22 {
		return nil, fmt.Errorf("invalid stat format")
	}

	utime := parseInt64(parts[13])
	stime := parseInt64(parts[14])
	uptimeSec := readUptime()

	totalTime := utime + stime
	uptimeMillis := uptimeSec * 1000

	var cpuPercent float64
	if uptimeMillis > 0 {
		cpuPercent = float64(totalTime*100) / float64(uptimeMillis)
	}

	// statm: size, resident, shared, text, lib, data, dt
	statmParts := strings.Split(string(statm), " ")
	var rssPages int64
	if len(statmParts) >= 2 {
		rssPages = parseInt64(statmParts[1])
	}
	rssMB := int(rssPages * 4 / 1024) // pages -> KB -> MB (typical page size 4KB)

	return &ServerMetrics{
		PID:          pid,
		Status:       "running",
		UptimeSeconds: uptimeSec,
		RSSMB:        rssMB,
		CPUPercent:   cpuPercent,
		TPS:          20,
		MSPT:         50,
		OnlinePlayers: 0,
	}, nil
}

func parseInt64(s string) int64 {
	var n int64
	fmt.Sscanf(s, "%d", &n)
	return n
}

func readUptime() int64 {
	content, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0
	}
	var sec float64
	fmt.Sscanf(string(content), "%f", &sec)
	return int64(sec)
}
