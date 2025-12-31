package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"sakthiRathinam/oom-saver/pkg/ui"
)

type CleanupSettings struct {
	KillUserProcesses bool
	KillBrowsers      bool
	KillSafe          bool
	KillImportant     bool
	MinOOMScore       int
	ZombiesOnly       bool
	Interval          int
	MemoryAlert       bool
	MemoryThreshold   int
	MemoryCooldown    int
}

func askYesNo(question string, defaultYes bool) bool {
	reader := bufio.NewReader(os.Stdin)
	defaultStr := "y/N"
	if defaultYes {
		defaultStr = "Y/n"
	}

	fmt.Printf("%s %s [%s]: ", ui.Cyan("?"), question, defaultStr)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response == "" {
		return defaultYes
	}

	return response == "y" || response == "yes"
}

func askNumber(question string, defaultValue int, min int, max int) int {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("%s %s [%d]: ", ui.Cyan("?"), question, defaultValue)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(response)

	if response == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(response)
	if err != nil || value < min || value > max {
		fmt.Printf("%s Invalid input, using default: %d\n", ui.Yellow("⚠️"), defaultValue)
		return defaultValue
	}

	return value
}

func getCleanupSettings() CleanupSettings {
	settings := CleanupSettings{}

	fmt.Printf("\n%s\n\n", ui.Bold("Configure Cleanup Rules"))
	fmt.Printf("The systemd service will continuously monitor and clean up processes.\n")
	fmt.Printf("Configure what types of processes should be automatically killed:\n\n")

	// Default: user processes and browsers only
	settings.KillUserProcesses = askYesNo("Auto-kill user processes (UID >= 1000)?", true)
	settings.KillBrowsers = askYesNo("Auto-kill browser processes (Chrome, Firefox, etc.)?", true)
	settings.ZombiesOnly = askYesNo("Only target zombie/problematic processes (safer)?", true)

	fmt.Printf("\n%s\n", ui.Bold("Advanced Options"))
	fmt.Printf("Configure additional cleanup rules based on safety classification:\n\n")

	if askYesNo("Enable advanced safety-based cleanup?", false) {
		settings.KillSafe = askYesNo("  Auto-kill 'Safe' classified processes?", false)
		settings.KillImportant = askYesNo("  Auto-kill 'Important' classified processes?", false)
	}

	if askYesNo("Enable OOM score-based cleanup?", false) {
		settings.MinOOMScore = askNumber("  Minimum OOM score to kill (recommended: 500-800)", 0, 0, 1000)
	}

	fmt.Printf("\n%s\n", ui.Bold("Memory Monitoring"))
	fmt.Printf("Enable desktop notifications when system memory is low:\n\n")

	settings.MemoryAlert = askYesNo("Enable memory alerts?", true)
	if settings.MemoryAlert {
		settings.MemoryThreshold = askNumber("  Alert when available memory is below (GB)", 3, 1, 32)
		settings.MemoryCooldown = askNumber("  Cooldown between alerts (minutes)", 15, 1, 120)
	}

	fmt.Printf("\n%s\n", ui.Bold("Monitoring Interval"))
	settings.Interval = askNumber("How often should the monitor scan (seconds)?", 10, 1, 3600)

	return settings
}

func generateSystemdService(settings CleanupSettings) string {
	cmdFlags := []string{"monitor", "--use-config"}

	if settings.KillUserProcesses {
		cmdFlags = append(cmdFlags, "--kill-user-processes")
	}
	if settings.KillBrowsers {
		cmdFlags = append(cmdFlags, "--kill-browsers")
	}
	if settings.KillSafe {
		cmdFlags = append(cmdFlags, "--kill-safe")
	}
	if settings.KillImportant {
		cmdFlags = append(cmdFlags, "--kill-important")
	}
	if settings.MinOOMScore > 0 {
		cmdFlags = append(cmdFlags, fmt.Sprintf("--min-oom-score=%d", settings.MinOOMScore))
	}
	if settings.ZombiesOnly {
		cmdFlags = append(cmdFlags, "--zombies-only")
	}
	if settings.MemoryAlert {
		cmdFlags = append(cmdFlags, "--memory-alert")
		if settings.MemoryThreshold != 3 {
			cmdFlags = append(cmdFlags, fmt.Sprintf("--memory-threshold=%d", settings.MemoryThreshold))
		}
		if settings.MemoryCooldown != 15 {
			cmdFlags = append(cmdFlags, fmt.Sprintf("--memory-cooldown=%d", settings.MemoryCooldown))
		}
	}
	if settings.Interval != 5 {
		cmdFlags = append(cmdFlags, fmt.Sprintf("--interval=%ds", settings.Interval))
	}

	execStart := "/usr/local/bin/oom-killer " + strings.Join(cmdFlags, " ")

	return fmt.Sprintf(`[Unit]
Description=OOM Killer - Process Monitor and Zombie Killer
After=network.target

[Service]
Type=simple
ExecStart=%s
Restart=always
RestartSec=10
Environment="DISPLAY=:0"
Environment="DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/1000/bus"

[Install]
WantedBy=multi-user.target
`, execStart)
}

func formatBool(value bool) string {
	if value {
		return ui.Green("enabled")
	}
	return ui.Yellow("disabled")
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install oom-killer as a systemd service",
	Long:  `Install the oom-killer binary to /usr/local/bin and create a systemd service.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if os.Geteuid() != 0 {
			return fmt.Errorf("%s this command requires root privileges. Run with sudo", ui.Red("✗"))
		}

		ui.PrintHeader("📦 INSTALLING OOM-KILLER")

		// Check if notify-send is installed
		fmt.Printf("\n%s Checking dependencies...\n", ui.Cyan("ℹ️"))
		_, err := exec.LookPath("notify-send")
		if err != nil {
			fmt.Printf("\n%s notify-send is not installed!\n", ui.Red("✗"))
			fmt.Printf("\nMemory alerts require notify-send for desktop notifications.\n")
			fmt.Printf("Please install it first:\n\n")
			fmt.Printf("  %s\n", ui.Cyan("sudo apt install libnotify-bin"))
			fmt.Printf("  %s (Debian/Ubuntu)\n\n", ui.Yellow("# For Debian/Ubuntu"))
			fmt.Printf("  %s\n", ui.Cyan("sudo dnf install libnotify"))
			fmt.Printf("  %s (Fedora/RHEL)\n\n", ui.Yellow("# For Fedora/RHEL"))
			fmt.Printf("  %s\n", ui.Cyan("sudo pacman -S libnotify"))
			fmt.Printf("  %s (Arch Linux)\n\n", ui.Yellow("# For Arch Linux"))
			fmt.Printf("  %s\n", ui.Cyan("sudo zypper install libnotify-tools"))
			fmt.Printf("  %s (openSUSE)\n\n", ui.Yellow("# For openSUSE"))
			return fmt.Errorf("dependency check failed: notify-send not found")
		}
		fmt.Printf("   %s notify-send found\n", ui.Green("✓"))

		// Interactive configuration
		settings := getCleanupSettings()

		// Generate systemd service with custom settings
		serviceContent := generateSystemdService(settings)

		fmt.Printf("\n%s\n", ui.Bold("Configuration Summary:"))
		fmt.Printf("  • User processes:     %s\n", formatBool(settings.KillUserProcesses))
		fmt.Printf("  • Browser processes:  %s\n", formatBool(settings.KillBrowsers))
		fmt.Printf("  • Zombies only:       %s\n", formatBool(settings.ZombiesOnly))
		fmt.Printf("  • Safe level:         %s\n", formatBool(settings.KillSafe))
		fmt.Printf("  • Important level:    %s\n", formatBool(settings.KillImportant))
		if settings.MinOOMScore > 0 {
			fmt.Printf("  • Min OOM score:      %d\n", settings.MinOOMScore)
		}
		fmt.Printf("  • Memory alerts:      %s\n", formatBool(settings.MemoryAlert))
		if settings.MemoryAlert {
			fmt.Printf("    - Threshold:        %d GB\n", settings.MemoryThreshold)
			fmt.Printf("    - Cooldown:         %d min\n", settings.MemoryCooldown)
		}
		fmt.Printf("  • Scan interval:      %ds\n", settings.Interval)

		if !askYesNo("\nProceed with installation?", true) {
			fmt.Printf("\n%s Installation cancelled\n", ui.Yellow("⚠️"))
			return nil
		}

		execPath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("failed to get executable path: %w", err)
		}

		fmt.Printf("\n%s Copying binary to /usr/local/bin/oom-killer...\n", ui.Cyan("1."))
		input, err := os.ReadFile(execPath)
		if err != nil {
			return fmt.Errorf("failed to read executable: %w", err)
		}

		err = os.WriteFile("/usr/local/bin/oom-killer", input, 0755)
		if err != nil {
			return fmt.Errorf("failed to copy binary: %w", err)
		}
		fmt.Printf("   %s Binary installed\n", ui.Green("✓"))

		fmt.Printf("\n%s Creating systemd service file...\n", ui.Cyan("2."))
		err = os.WriteFile("/etc/systemd/system/oom-killer.service", []byte(serviceContent), 0644)
		if err != nil {
			return fmt.Errorf("failed to create service file: %w", err)
		}
		fmt.Printf("   %s Service file created\n", ui.Green("✓"))

		fmt.Printf("\n%s Reloading systemd daemon...\n", ui.Cyan("3."))
		err = exec.Command("systemctl", "daemon-reload").Run()
		if err != nil {
			return fmt.Errorf("failed to reload systemd: %w", err)
		}
		fmt.Printf("   %s Daemon reloaded\n", ui.Green("✓"))

		fmt.Printf("\n%s Enabling service...\n", ui.Cyan("4."))
		err = exec.Command("systemctl", "enable", "oom-killer.service").Run()
		if err != nil {
			return fmt.Errorf("failed to enable service: %w", err)
		}
		fmt.Printf("   %s Service enabled\n", ui.Green("✓"))

		fmt.Printf("\n%s Starting service...\n", ui.Cyan("5."))
		err = exec.Command("systemctl", "start", "oom-killer.service").Run()
		if err != nil {
			return fmt.Errorf("failed to start service: %w", err)
		}
		fmt.Printf("   %s Service started\n", ui.Green("✓"))

		fmt.Printf("\n%s Installation complete!\n", ui.Green("✓"))
		fmt.Printf("\n%s\n", ui.Bold("Useful commands:"))
		fmt.Printf("  • Check status:  %s\n", ui.Cyan("sudo systemctl status oom-killer"))
		fmt.Printf("  • View logs:     %s\n", ui.Cyan("sudo journalctl -u oom-killer -f"))
		fmt.Printf("  • Stop service:  %s\n", ui.Cyan("sudo systemctl stop oom-killer"))
		fmt.Printf("  • Disable:       %s\n", ui.Cyan("sudo systemctl disable oom-killer"))
		fmt.Println()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
