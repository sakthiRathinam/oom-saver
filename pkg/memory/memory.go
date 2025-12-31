package memory

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type MemoryStats struct {
	TotalMB     int
	FreeMB      int
	AvailableMB int
	UsedMB      int
	UsedPercent float64
}

type MemoryAlert struct {
	LastAlertTime   time.Time
	AlertCooldown   time.Duration
	ThresholdGB     int
	NotificationSent bool
}

func NewMemoryAlert(thresholdGB int, cooldownMinutes int) *MemoryAlert {
	return &MemoryAlert{
		ThresholdGB:   thresholdGB,
		AlertCooldown: time.Duration(cooldownMinutes) * time.Minute,
	}
}

// GetMemoryStats reads memory information from /proc/meminfo
func GetMemoryStats() (*MemoryStats, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return nil, fmt.Errorf("failed to open /proc/meminfo: %w", err)
	}
	defer file.Close()

	stats := &MemoryStats{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		if len(fields) < 2 {
			continue
		}

		key := strings.TrimSuffix(fields[0], ":")
		valueKB, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}

		valueMB := valueKB / 1024

		switch key {
		case "MemTotal":
			stats.TotalMB = valueMB
		case "MemFree":
			stats.FreeMB = valueMB
		case "MemAvailable":
			stats.AvailableMB = valueMB
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading /proc/meminfo: %w", err)
	}

	stats.UsedMB = stats.TotalMB - stats.AvailableMB
	if stats.TotalMB > 0 {
		stats.UsedPercent = float64(stats.UsedMB) / float64(stats.TotalMB) * 100
	}

	return stats, nil
}

// CheckMemoryThreshold checks if available memory is below the threshold
func (ma *MemoryAlert) CheckMemoryThreshold(stats *MemoryStats) (bool, string) {
	thresholdMB := ma.ThresholdGB * 1024

	if stats.AvailableMB <= thresholdMB {
		return true, fmt.Sprintf("Low memory! Only %d MB (%.1f%%) available out of %d MB total",
			stats.AvailableMB, float64(stats.AvailableMB)/float64(stats.TotalMB)*100, stats.TotalMB)
	}

	return false, ""
}

// ShouldSendAlert checks if enough time has passed since the last alert
func (ma *MemoryAlert) ShouldSendAlert() bool {
	if ma.NotificationSent && time.Since(ma.LastAlertTime) < ma.AlertCooldown {
		return false
	}
	return true
}

// SendDesktopNotification sends a desktop notification using notify-send
func SendDesktopNotification(title string, message string, urgency string) error {
	// Check if notify-send is available
	_, err := exec.LookPath("notify-send")
	if err != nil {
		return fmt.Errorf("notify-send not found. Install libnotify-bin: %w", err)
	}

	// Send notification
	cmd := exec.Command("notify-send", "-u", urgency, "-i", "dialog-warning", title, message)

	// Try to use the display from environment
	if display := os.Getenv("DISPLAY"); display != "" {
		cmd.Env = append(os.Environ(), "DISPLAY="+display)
	}

	// Also try to detect user's DBUS session
	if dbusAddr := os.Getenv("DBUS_SESSION_BUS_ADDRESS"); dbusAddr != "" {
		cmd.Env = append(cmd.Env, "DBUS_SESSION_BUS_ADDRESS="+dbusAddr)
	}

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	return nil
}

// NotifyIfLowMemory checks memory and sends notification if below threshold
func (ma *MemoryAlert) NotifyIfLowMemory() error {
	stats, err := GetMemoryStats()
	if err != nil {
		return err
	}

	isLow, message := ma.CheckMemoryThreshold(stats)

	if isLow && ma.ShouldSendAlert() {
		err := SendDesktopNotification(
			"OOM-Saver",
			message,
			"critical",
		)

		if err != nil {
			// Notification failed, but don't stop monitoring
			fmt.Printf("Warning: Failed to send desktop notification: %v\n", err)
		} else {
			ma.LastAlertTime = time.Now()
			ma.NotificationSent = true
		}
	} else if !isLow {
		// Reset notification flag when memory is back to normal
		ma.NotificationSent = false
	}

	return nil
}

// GetMemoryStatusString returns a formatted string with current memory status
func GetMemoryStatusString(stats *MemoryStats) string {
	return fmt.Sprintf("Memory: %d/%d MB used (%.1f%%), %d MB available",
		stats.UsedMB, stats.TotalMB, stats.UsedPercent, stats.AvailableMB)
}
