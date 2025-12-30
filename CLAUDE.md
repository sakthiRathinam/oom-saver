# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

OOM-Killer is a beautiful Cobra-based CLI tool for Linux system process monitoring and management. It features intelligent process safety classification, monitors processes, detects and kills zombies safely, and can be installed as a systemd service.

### Key Features

- **4-Level Safety Classification**: Critical, Important, Safe, Unknown
- **Smart Zombie Killing**: Only kills safe zombies by default
- **Beautiful CLI Output**: Colors, icons, progress bars
- **Safety Guards**: Prevents accidental killing of critical processes
- **Flexible Filtering**: Filter by status or safety level
- **Systemd Integration**: Install as a system service

## Development Commands

### Build
```bash
go build -o oom-killer
```

### Run Commands
```bash
# List all processes
./oom-killer list

# List with filters
./oom-killer list --safety critical  # Show only critical processes
./oom-killer list --safety safe      # Show only safe processes
./oom-killer list --status zombie    # Show only zombies
./oom-killer list --limit 50         # Limit output

# Classify a specific process
./oom-killer classify <PID>

# Monitor processes continuously
./oom-killer monitor
./oom-killer monitor --auto-kill-all-zombies  # Kill all zombies (unsafe)
./oom-killer monitor --no-auto-kill           # Disable auto-kill

# Show statistics (includes safety breakdown)
./oom-killer stats

# Kill a specific process (with safety checks)
./oom-killer kill <PID>
./oom-killer kill <PID> --force  # Force kill critical process (dangerous)

# Install as systemd service (requires sudo)
sudo ./oom-killer install
```

### Testing
```bash
go test ./...
```

### Format Code
```bash
go fmt ./...
```

## Code Architecture

### Project Structure

```
oom-saver/
├── main.go                    # Entry point - executes Cobra root command
├── cmd/                       # Cobra commands
│   ├── root.go               # Root command configuration
│   ├── list.go               # List processes once (with safety filtering)
│   ├── monitor.go            # Continuous monitoring (smart zombie killing)
│   ├── stats.go              # Process statistics (with safety breakdown)
│   ├── kill.go               # Kill specific process (with safety checks)
│   ├── classify.go           # Show detailed process classification
│   └── install.go            # Install as systemd service
├── pkg/
│   ├── process/              # Process management logic
│   │   ├── process.go        # Process detection, parsing, killing
│   │   └── classifier.go     # Safety classification logic
│   └── ui/                   # UI formatting and display
│       └── ui.go             # Colors, progress bars, tables, safety icons
├── go.mod
└── CLAUDE.md
```

### Core Components

**pkg/process/process.go**
- `Process` struct: Represents a system process (Name, PID, Status, SafetyLevel, UID, PPID, OOMScore)
- `GetAllRunningProcesses()`: Main entry point - gets all processes and classifies them
- `getAllRunningProcessesFromLinux()`: Reads from /proc directory
- `parseProcessState()`: Maps Linux state codes to readable names
- `KillProcessIfZombie(processes, killAll)`: Smart zombie killing (safe only by default)
- `KillProcessWithSafety()`: Kills with safety checks
- `GetProcessByPID()`: Gets detailed info for specific PID

**pkg/process/classifier.go** (NEW)
- `ClassifyProcess()`: Main classification function - determines safety level
- `readProcessUID()`: Reads process owner from /proc/[pid]/status
- `readProcessPPID()`: Reads parent process ID
- `readProcessOOMScore()`: Reads Linux OOM score
- `isKernelThread()`: Detects kernel threads by name pattern
- `isCriticalProcessName()`: Checks against critical process list
- Hardcoded lists of critical and important process names

**pkg/ui/ui.go**
- Color scheme:
  - Status: Green (running), Yellow (sleeping/idle), Red (zombie), Cyan (headers)
  - Safety: RedBold (critical), Yellow (important), Green (safe), White (unknown)
- `GetSafetyColor()`: Returns color function for safety level
- `GetSafetyIcon()`: Returns emoji icon (🔴 🟡 🟢 ⚪)
- `PrintHeader()`: Prints formatted headers with borders
- `PrintProcessTable()`: Displays process table with colors and safety column
- `CreateProgressBar()`: Creates animated progress bars
- `PrintStats()`: Displays statistics by status AND safety level

**cmd/ Commands**
- Each command is a separate file that uses Cobra framework
- Commands use pkg/process for logic and pkg/ui for display
- `list.go`: Added `--safety` filter
- `monitor.go`: Added `--auto-kill-all-zombies` and `--no-auto-kill` flags
- `kill.go`: Added safety checks and `--force` flag
- `classify.go`: NEW - Shows detailed classification for a PID

### Safety Classification System

The process classifier assigns one of four safety levels to each process:

**🔴 Critical (Never auto-kill)**
- PID 1 (systemd/init)
- Kernel threads (names in brackets like `[kworker]`)
- Essential services: systemd-*, sshd, dbus-daemon, NetworkManager
- Processes with OOM score < -500
- Killing requires `--force` flag and double confirmation

**🟡 Important (Warn before kill)**
- System daemons: cron, rsyslog, journald, udev
- Network services: nginx, apache
- Databases: postgres, mysql, mongodb
- Root-owned processes with systemd as parent (PPID=1)
- Requires confirmation before killing

**🟢 Safe (Can kill freely)**
- User processes (UID >= 1000)
- High OOM score (> 300)
- **All zombie processes** (already dead)
- Standard confirmation only

**⚪ Unknown (Requires investigation)**
- Processes that don't fit other categories
- Manual judgment required

### Classification Logic

The classifier reads from `/proc/[pid]/`:
1. **UID** from `/proc/[pid]/status` - identifies process owner
2. **PPID** from `/proc/[pid]/status` - identifies parent process
3. **OOM Score** from `/proc/[pid]/oom_score` - Linux kernel's kill priority
4. **Process name** - matches against hardcoded critical/important lists
5. **Status** - zombies are always safe

### Platform Support

Currently supports:
- **Linux**: Fully implemented using /proc filesystem
- **Windows**: Not implemented
- **macOS**: Not implemented

### Dependencies

- `github.com/spf13/cobra` - CLI framework
- `github.com/fatih/color` - Terminal colors
- `github.com/schollz/progressbar/v3` - Progress bars

### systemd Service

The `install` command creates a systemd service at `/etc/systemd/system/oom-killer.service` that runs `oom-killer monitor` continuously. The binary is installed to `/usr/local/bin/oom-killer`.

Manage the service with:
```bash
sudo systemctl status oom-killer
sudo systemctl stop oom-killer
sudo systemctl start oom-killer
sudo journalctl -u oom-killer -f
```
