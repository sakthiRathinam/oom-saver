# OOM-Killer

A beautiful, intelligent Linux process monitoring and management tool with smart safety classification and automated zombie cleanup.

## Features

- **4-Level Safety Classification System** - Automatically classifies processes as Critical, Important, Safe, or Unknown
- **Smart Zombie Process Detection** - Finds and safely kills zombie processes
- **Intelligent Auto-Cleanup** - Configurable automatic cleanup based on safety levels and OOM scores
- **Memory Monitoring & Alerts** - Desktop notifications when system memory is low (before OOM killer kicks in)
- **Beautiful CLI Output** - Color-coded tables, progress bars, and intuitive icons
- **Systemd Integration** - Install as a background service with interactive configuration
- **Flexible Filtering** - Filter processes by status, safety level, or custom criteria
- **Safety Guards** - Prevents accidental killing of critical system processes

## Installation

### Prerequisites

- Linux operating system
- Go 1.16 or higher (for building from source)
- Root access (for systemd service installation)
- **notify-send** (required for memory alerts - usually pre-installed)

### Build from Source

```bash
git clone <repository-url>
cd oom-saver
go build -o oom-killer
```

### Install as Systemd Service

```bash
sudo ./oom-killer install
```

The install command will:
1. **Check dependencies** - Verifies `notify-send` is installed (required for memory alerts)
2. **Interactive configuration** - Asks you:
   - What types of processes to auto-kill (user processes, browsers, etc.)
   - Which safety levels to target (Safe, Important, etc.)
   - Memory alert settings (threshold and cooldown)
   - OOM score thresholds for aggressive cleanup
   - How often to scan for problematic processes
3. **Show summary** - Displays all settings before proceeding
4. **Install** - Copies binary and creates systemd service

**If notify-send is missing**, the installer will stop and show you how to install it:
```bash
# Debian/Ubuntu
sudo apt install libnotify-bin

# Fedora/RHEL
sudo dnf install libnotify

# Arch Linux
sudo pacman -S libnotify

# openSUSE
sudo zypper install libnotify-tools
```

**Default behavior**: Only cleanup user-owned processes (UID >= 1000) and browser processes, preserving all system services.

## Usage

### List Processes

```bash
# List all processes with safety classification
./oom-killer list

# Filter by safety level
./oom-killer list --safety critical
./oom-killer list --safety safe

# Filter by status
./oom-killer list --status zombie

# Limit output
./oom-killer list --limit 50
```

### Monitor Processes

```bash
# Monitor with safe auto-kill (default)
./oom-killer monitor

# Monitor with aggressive zombie cleanup
./oom-killer monitor --auto-kill-all-zombies

# Monitor without auto-kill
./oom-killer monitor --no-auto-kill

# Monitor with memory alerts (desktop notifications)
./oom-killer monitor --memory-alert
./oom-killer monitor --memory-alert --memory-threshold=2 --memory-cooldown=10
```

### Show Statistics

```bash
# Display process statistics by status and safety level
./oom-killer stats
```

### Classify a Process

```bash
# Show detailed classification for a specific PID
./oom-killer classify <PID>
```

### Kill a Process

```bash
# Kill with safety checks
./oom-killer kill <PID>

# Force kill critical process (requires confirmation)
./oom-killer kill <PID> --force
```

## Safety Classification

### 🔴 Critical (Never auto-kill)

- PID 1 (init/systemd)
- Kernel threads (`[kworker]`, `[ksoftirqd]`, etc.)
- Essential system services: systemd-*, sshd, dbus-daemon, NetworkManager
- Processes with OOM score < -500
- Requires `--force` flag to kill

### 🟡 Important (Warn before kill)

- System daemons: cron, rsyslog, journald
- Network services: nginx, apache
- Databases: postgres, mysql, mongodb
- Root-owned processes parented by systemd
- Requires confirmation before killing

### 🟢 Safe (Can kill freely)

- User processes (UID >= 1000)
- High OOM score (> 300)
- Zombie processes (already dead)
- Browser processes (Chrome, Firefox, etc.)
- Standard confirmation only

### ⚪ Unknown (Requires investigation)

- Processes that don't fit other categories
- Manual judgment recommended

## Configuration

### Systemd Service

After installation, manage the service with:

```bash
# Check status
sudo systemctl status oom-killer

# Start/stop/restart
sudo systemctl start oom-killer
sudo systemctl stop oom-killer
sudo systemctl restart oom-killer

# View logs
sudo journalctl -u oom-killer -f

# Disable auto-start
sudo systemctl disable oom-killer
```

The service runs `oom-killer monitor` with your configured options continuously in the background.

### Reconfigure Service

To change cleanup rules after installation:

```bash
# Uninstall and reinstall with new settings
sudo systemctl stop oom-killer
sudo systemctl disable oom-killer
sudo rm /etc/systemd/system/oom-killer.service
sudo ./oom-killer install
```

Or manually edit `/etc/systemd/system/oom-killer.service` and reload:

```bash
sudo systemctl daemon-reload
sudo systemctl restart oom-killer
```

## How It Works

### Memory Monitoring

OOM-Killer can monitor system memory and send desktop notifications **before** the Linux OOM killer starts killing processes.

When memory alerts are enabled:
1. Monitors available system memory every scan interval
2. Sends a desktop notification when available memory drops below the threshold (default: 3 GB)
3. Includes a cooldown period to avoid notification spam (default: 15 minutes)
4. Automatically stops alerting when memory is back to normal levels

**Requirements:** Desktop notifications require `notify-send` (usually pre-installed):

```bash
# Install if missing
sudo apt install libnotify-bin  # Debian/Ubuntu
sudo dnf install libnotify      # Fedora
sudo pacman -S libnotify        # Arch
```

The installer automatically configures environment variables for desktop notifications when running as a systemd service.

### Process Detection

OOM-Killer reads from the Linux `/proc` filesystem to gather:
- Process ID (PID)
- Process name
- Status (running, sleeping, zombie, etc.)
- Owner (UID)
- Parent process (PPID)
- Linux OOM score

### Classification Algorithm

Each process is classified based on:

1. **PID 1 check** - Always critical
2. **Kernel thread detection** - Names in brackets `[...]` are critical
3. **Name matching** - Against hardcoded critical/important process lists
4. **OOM score** - Scores < -500 are critical, > 300 are safe
5. **Ownership** - User processes (UID >= 1000) are generally safe
6. **Parent process** - Root processes with systemd parent are important
7. **Status** - All zombies are safe (already dead)

### Smart Zombie Killing

By default, OOM-Killer only kills zombie processes that are classified as "Safe":
- User-owned zombies
- Zombies from crashed browsers
- High OOM score zombies

Critical and important zombies are reported but not killed automatically, allowing system administrators to investigate.

## Architecture

```
oom-saver/
├── main.go                 # Entry point
├── cmd/                    # Cobra CLI commands
│   ├── root.go            # Root command
│   ├── list.go            # List processes
│   ├── monitor.go         # Continuous monitoring
│   ├── stats.go           # Statistics
│   ├── kill.go            # Kill process
│   ├── classify.go        # Classify process
│   └── install.go         # Install systemd service
├── pkg/
│   ├── process/           # Core process logic
│   │   ├── process.go     # Detection, parsing, killing
│   │   └── classifier.go  # Safety classification
│   └── ui/                # CLI interface
│       └── ui.go          # Colors, tables, progress bars
└── README.md
```

## Dependencies

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Color](https://github.com/fatih/color) - Terminal colors
- [ProgressBar](https://github.com/schollz/progressbar) - Progress bars

## Development

```bash
# Run tests
go test ./...

# Format code
go fmt ./...

# Build
go build -o oom-killer

# Run locally
./oom-killer list
```

## Platform Support

- **Linux**: Fully supported ✅
- **macOS**: Not implemented ❌
- **Windows**: Not implemented ❌

## Safety & Warnings

- **Always review** what processes will be killed before enabling auto-kill
- **Never force-kill** critical processes unless you know what you're doing
- **Test in a safe environment** before deploying to production servers
- **Monitor logs** regularly when running as a service
- The tool is designed to be safe by default, but system administration requires care

## License

[Add your license here]

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## Troubleshooting

### Permission denied errors

Most commands require root access to read all process information:
```bash
sudo ./oom-killer list
```

### Service won't start

Check logs for details:
```bash
sudo journalctl -u oom-killer -n 50
```

### Too aggressive killing

Reconfigure with more conservative settings:
```bash
sudo ./oom-killer install
# Choose more restrictive options during setup
```

## FAQ

**Q: Will this kill my important applications?**
A: By default, no. The tool only auto-kills user processes and browsers that are already zombies or problematic. System services are protected.

**Q: What happens if I accidentally kill a critical process?**
A: The tool requires `--force` flag and double confirmation to kill critical processes. If you do kill one, systemd typically restarts essential services automatically.

**Q: How often does the monitor check for processes?**
A: By default, every 10 seconds. This is configurable during installation.

**Q: Can I run this on a production server?**
A: Yes, but start with conservative settings and monitor carefully. The default "user processes & browsers only" mode is safe for most servers.
