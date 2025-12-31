package process

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var criticalProcessNames = []string{
	"systemd", "init", "kthreadd", "kworker", "sshd", "dbus-daemon", "NetworkManager",
	"systemd-journald", "systemd-logind", "systemd-udevd", "systemd-networkd",
}

var importantProcessNames = []string{
	"cron", "crond", "rsyslog", "rsyslogd", "journald", "udev", "udevd",
	"nginx", "apache2", "httpd", "postgres", "postgresql", "mysqld", "mysql",
	"mongod", "mongodb", "redis-server", "dockerd", "containerd",
}

var browserProcessNames = []string{
	"chrome", "chromium", "firefox", "brave", "opera", "vivaldi",
	"safari", "edge", "msedge", "epiphany", "qutebrowser", "falkon",
	"palemoon", "waterfox", "seamonkey", "google-chrome",
}

func ClassifyProcess(p *Process) string {
	if p.Status == "zombie" {
		return "safe"
	}

	if p.PID == 1 {
		return "critical"
	}

	if isKernelThread(p.Name) {
		return "critical"
	}

	if p.OOMScore < -500 {
		return "critical"
	}

	if isCriticalProcessName(p.Name) {
		return "critical"
	}

	if isImportantProcessName(p.Name) {
		return "important"
	}

	if p.UID >= 1000 {
		return "safe"
	}

	if p.OOMScore > 300 {
		return "safe"
	}

	if p.UID == 0 && p.PPID == 1 {
		return "important"
	}

	return "unknown"
}

func isKernelThread(name string) bool {
	return strings.HasPrefix(name, "[") && strings.HasSuffix(name, "]")
}

func isCriticalProcessName(name string) bool {
	for _, critical := range criticalProcessNames {
		if name == critical || strings.HasPrefix(name, critical) {
			return true
		}
	}
	return false
}

func isImportantProcessName(name string) bool {
	for _, important := range importantProcessNames {
		if name == important || strings.HasPrefix(name, important) {
			return true
		}
	}
	return false
}

func IsBrowserProcess(name string) bool {
	lowerName := strings.ToLower(name)
	for _, browser := range browserProcessNames {
		if lowerName == browser || strings.Contains(lowerName, browser) {
			return true
		}
	}
	return false
}

func readProcessUID(pid int) (int, error) {
	statusPath := filepath.Join("/proc", strconv.Itoa(pid), "status")
	data, err := os.ReadFile(statusPath)
	if err != nil {
		return -1, err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Uid:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				uid, err := strconv.Atoi(fields[1])
				if err != nil {
					return -1, err
				}
				return uid, nil
			}
		}
	}

	return -1, nil
}

func readProcessPPID(pid int) (int, error) {
	statusPath := filepath.Join("/proc", strconv.Itoa(pid), "status")
	data, err := os.ReadFile(statusPath)
	if err != nil {
		return -1, err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "PPid:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				ppid, err := strconv.Atoi(fields[1])
				if err != nil {
					return -1, err
				}
				return ppid, nil
			}
		}
	}

	return -1, nil
}

func readProcessOOMScore(pid int) (int, error) {
	oomScorePath := filepath.Join("/proc", strconv.Itoa(pid), "oom_score")
	data, err := os.ReadFile(oomScorePath)
	if err != nil {
		return 0, err
	}

	score, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, err
	}

	return score, nil
}

func readProcessOOMScoreAdj(pid int) (int, error) {
	oomScoreAdjPath := filepath.Join("/proc", strconv.Itoa(pid), "oom_score_adj")
	data, err := os.ReadFile(oomScoreAdjPath)
	if err != nil {
		return 0, err
	}

	scoreAdj, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, err
	}

	return scoreAdj, nil
}
