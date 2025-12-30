package process

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

type Process struct {
	Name        string
	PID         int
	Status      string
	SafetyLevel string
	UID         int
	PPID        int
	OOMScore    int
}

func GetAllRunningProcesses() ([]Process, error) {
	currentOS := runtime.GOOS
	var processes []Process
	if currentOS == "windows" {
		return processes, errors.New("not implemented")
	} else if currentOS == "linux" {
		return getAllRunningProcessesFromLinux()
	} else if currentOS == "darwin" {
		return processes, errors.New("not implemented")
	}
	return processes, errors.New("unsupported operating system")
}

func getAllRunningProcessesFromLinux() ([]Process, error) {
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
			continue
		}
		processName := strings.TrimSpace(string(commData))

		statusPath := filepath.Join("/proc", entry.Name(), "status")
		statusData, err := os.ReadFile(statusPath)
		if err != nil {
			continue
		}

		processState := parseProcessState(string(statusData))

		uid, _ := readProcessUID(pid)
		ppid, _ := readProcessPPID(pid)
		oomScore, _ := readProcessOOMScore(pid)

		process := Process{
			Name:     processName,
			PID:      pid,
			Status:   processState,
			UID:      uid,
			PPID:     ppid,
			OOMScore: oomScore,
		}

		process.SafetyLevel = ClassifyProcess(&process)

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

func KillProcessIfZombie(processes []Process, killAll bool) ([]Process, error) {
	var activeProcesses []Process

	for _, proc := range processes {
		if proc.Status == "zombie" {
			shouldKill := killAll || proc.SafetyLevel == "safe"

			if shouldKill {
				fmt.Printf("Found zombie process: PID %d (%s) [%s], sending SIGTERM...\n", proc.PID, proc.Name, proc.SafetyLevel)
				err := syscall.Kill(proc.PID, syscall.SIGTERM)
				if err != nil {
					fmt.Printf("  Warning: failed to send signal to PID %d: %v\n", proc.PID, err)
				}
			} else {
				fmt.Printf("Skipping %s zombie: PID %d (%s) - use --auto-kill-all-zombies to kill\n", proc.SafetyLevel, proc.PID, proc.Name)
				activeProcesses = append(activeProcesses, proc)
			}
		} else {
			activeProcesses = append(activeProcesses, proc)
		}
	}

	return activeProcesses, nil
}

func KillProcess(pid int, signal syscall.Signal) error {
	return syscall.Kill(pid, signal)
}

func KillProcessWithSafety(pid int, signal syscall.Signal, force bool) error {
	processes, err := GetAllRunningProcesses()
	if err != nil {
		return fmt.Errorf("failed to get process info: %w", err)
	}

	var targetProcess *Process
	for _, p := range processes {
		if p.PID == pid {
			targetProcess = &p
			break
		}
	}

	if targetProcess == nil {
		return fmt.Errorf("process %d not found", pid)
	}

	if targetProcess.SafetyLevel == "critical" && !force {
		return fmt.Errorf("cannot kill critical process (PID %d, %s) without --force flag", pid, targetProcess.Name)
	}

	return syscall.Kill(pid, signal)
}

func GetProcessByPID(pid int) (*Process, error) {
	processes, err := GetAllRunningProcesses()
	if err != nil {
		return nil, err
	}

	for _, p := range processes {
		if p.PID == pid {
			return &p, nil
		}
	}

	return nil, fmt.Errorf("process %d not found", pid)
}
