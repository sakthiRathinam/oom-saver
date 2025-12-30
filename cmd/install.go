package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"sakthiRathinam/oom-saver/pkg/ui"
)

const systemdService = `[Unit]
Description=OOM Killer - Process Monitor and Zombie Killer
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/oom-killer monitor
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
`

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install oom-killer as a systemd service",
	Long:  `Install the oom-killer binary to /usr/local/bin and create a systemd service.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if os.Geteuid() != 0 {
			return fmt.Errorf("%s this command requires root privileges. Run with sudo", ui.Red("✗"))
		}

		ui.PrintHeader("📦 INSTALLING OOM-KILLER")

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
		err = os.WriteFile("/etc/systemd/system/oom-killer.service", []byte(systemdService), 0644)
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
