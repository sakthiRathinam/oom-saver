# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

OOM-Killer is a beautiful Cobra-based CLI tool for Linux system process monitoring and management. It monitors processes, detects and kills zombies, and can be installed as a systemd service.

## Development Commands

### Build
```bash
go build -o oom-killer
```

### Run Commands
```bash
# List all processes
./oom-killer list

# Monitor processes continuously
./oom-killer monitor

# Show statistics
./oom-killer stats

# Kill a specific process
./oom-killer kill <PID>

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
│   ├── list.go               # List processes once
│   ├── monitor.go            # Continuous monitoring
│   ├── stats.go              # Process statistics
│   ├── kill.go               # Kill specific process
│   └── install.go            # Install as systemd service
├── pkg/
│   ├── process/              # Process management logic
│   │   └── process.go        # Process detection, parsing, killing
│   └── ui/                   # UI formatting and display
│       └── ui.go             # Colors, progress bars, tables
├── go.mod
└── CLAUDE.md
```

### Core Components

**pkg/process/process.go**
- `Process` struct: Represents a system process (Name, PID, Status)
- `GetAllRunningProcesses()`: Main entry point for getting all processes
- `getAllRunningProcessesFromLinux()`: Reads from /proc directory
- `parseProcessState()`: Maps Linux state codes to readable names
- `KillProcessIfZombie()`: Automatically kills zombie processes
- `KillProcess()`: Kills a specific process by PID and signal

**pkg/ui/ui.go**
- Color scheme: Green (running), Yellow (sleeping/idle), Red (zombie), Cyan (headers)
- `PrintHeader()`: Prints formatted headers with borders
- `PrintProcessTable()`: Displays process table with colors
- `CreateProgressBar()`: Creates animated progress bars
- `PrintStats()`: Displays process statistics by status

**cmd/ Commands**
- Each command is a separate file that uses Cobra framework
- Commands use pkg/process for logic and pkg/ui for display
- Flags are defined per-command for customization

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
