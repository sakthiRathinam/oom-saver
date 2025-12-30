package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func main() {
	fmt.Println("Starting OOM-saver process monitor...")
	fmt.Println("Monitoring processes every 5 seconds. Press Ctrl+C to exit.")
	fmt.Println()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	updateProcessList()

	for range ticker.C {
		updateProcessList()
	}
}

func updateProcessList() {
	processes, err := get_all_running_processes_from_os()
	if err != nil {
		fmt.Printf("Error fetching processes: %v\n", err)
		return
	}

	fmt.Printf("\n[%s] Found %d running processes:\n", time.Now().Format("15:04:05"), len(processes))

	limit := 10
	if len(processes) < limit {
		limit = len(processes)
	}

	for i := 0; i < limit; i++ {
		p := processes[i]
		fmt.Printf("  PID: %-6d Name: %-20s Status: %s\n", p.PID, p.Name, p.Status)
	}

	if len(processes) > limit {
		fmt.Printf("  ... and %d more processes\n", len(processes)-limit)
	}
}

type Process struct {
	Name   string
	PID    int
	Status string
}

func get_all_running_processes_from_os() ([]Process, error) {
	current_os := runtime.GOOS
	var processes []Process
	if current_os == "windows" {
		return processes, errors.New("not implemented")
	} else if current_os == "linux" {
		return get_all_running_processes_from_linux()
	} else if current_os == "darwin" {
		return processes, errors.New("not implemented")
	}
	return processes, errors.New("unsupported operating system")
}

func get_all_running_processes_from_linux() ([]Process, error) {
	var processes []Process

	entries, err := os.ReadDir("/proc")
	if err != nil {
		return processes, fmt.Errorf("failed to read /proc directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		commPath := filepath.Join("/proc", entry.Name(), "comm")
		commData, err := os.ReadFile(commPath)
		if err != nil {
			// Process might have terminated during read
			continue
		}
		processName := strings.TrimSpace(string(commData))

		statusPath := filepath.Join("/proc", entry.Name(), "status")
		statusData, err := os.ReadFile(statusPath)
		if err != nil {
			continue
		}

		processState := parseProcessState(string(statusData))

		process := Process{
			Name:   processName,
			PID:    pid,
			Status: processState,
		}
		processes = append(processes, process)
	}

	return processes, nil
}

func parseProcessState(statusContent string) string {
	lines := strings.Split(statusContent, "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "State:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				state := parts[1]
				switch state {
				case "R":
					return "running"
				case "S":
					return "sleeping"
				case "D":
					return "disk-sleep"
				case "Z":
					return "zombie"
				case "T":
					return "stopped"
				case "t":
					return "tracing-stop"
				case "W":
					return "paging"
				case "X", "x":
					return "dead"
				case "K":
					return "wakekill"
				case "P":
					return "parked"
				case "I":
					return "idle"
				default:
					return state
				}
			}
		}
	}

	return "unknown"
}
