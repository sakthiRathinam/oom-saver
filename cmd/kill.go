package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"sakthiRathinam/oom-saver/pkg/process"
	"sakthiRathinam/oom-saver/pkg/ui"
)

var (
	killSignal string
	killForce  bool
)

var killCmd = &cobra.Command{
	Use:   "kill <PID>",
	Short: "Kill a specific process by PID",
	Long:  `Send a signal to kill a specific process. Safety checks prevent killing critical processes without --force flag.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pid, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid PID: %s", args[0])
		}

		var sig syscall.Signal
		switch strings.ToUpper(killSignal) {
		case "SIGTERM", "TERM":
			sig = syscall.SIGTERM
		case "SIGKILL", "KILL":
			sig = syscall.SIGKILL
		default:
			return fmt.Errorf("unsupported signal: %s (use SIGTERM or SIGKILL)", killSignal)
		}

		proc, err := process.GetProcessByPID(pid)
		if err != nil {
			return fmt.Errorf("%s %w", ui.Red("✗"), err)
		}

		safetyColor := ui.GetSafetyColor(proc.SafetyLevel)
		safetyIcon := ui.GetSafetyIcon(proc.SafetyLevel)

		fmt.Printf("\n%s Process Information:\n", ui.Cyan("ℹ️"))
		fmt.Printf("  PID:    %d\n", proc.PID)
		fmt.Printf("  Name:   %s\n", proc.Name)
		fmt.Printf("  Status: %s\n", proc.Status)
		fmt.Printf("  Safety: %s %s\n", safetyIcon, safetyColor(proc.SafetyLevel))
		fmt.Println()

		if proc.SafetyLevel == "critical" && !killForce {
			fmt.Printf("%s Cannot kill CRITICAL process without --force flag!\n", ui.RedBold("⛔"))
			fmt.Printf("%s This is a system-critical process. Killing it may crash your system.\n", ui.Red("⚠️"))
			fmt.Printf("%s Use --force flag only if you know what you're doing.\n\n", ui.Yellow("💡"))
			return fmt.Errorf("safety check failed")
		}

		if proc.SafetyLevel == "critical" && killForce {
			fmt.Printf("%s %s KILLING CRITICAL PROCESS!\n", ui.RedBold("⛔"), ui.RedBold("WARNING:"))
			fmt.Printf("%s This may CRASH your system or cause data loss!\n", ui.Red("⚠️"))
			fmt.Print(ui.RedBold("Type 'I UNDERSTAND THE RISK' to continue: "))

			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(response)

			if response != "I UNDERSTAND THE RISK" {
				fmt.Println(ui.Yellow("✗ Cancelled"))
				return nil
			}
		} else if proc.SafetyLevel == "important" {
			fmt.Printf("%s About to send %s to IMPORTANT process (PID %d)\n", ui.Yellow("⚠️"), ui.Bold(killSignal), pid)
			fmt.Printf("%s This may affect system services or running applications.\n", ui.Yellow("⚠️"))
			fmt.Print("Continue? (y/N): ")

			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))

			if response != "y" && response != "yes" {
				fmt.Println(ui.Yellow("✗ Cancelled"))
				return nil
			}
		} else {
			fmt.Printf("%s About to send %s to PID %d\n", ui.Yellow("⚠️"), ui.Bold(killSignal), pid)
			fmt.Print("Continue? (y/N): ")

			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))

			if response != "y" && response != "yes" {
				fmt.Println(ui.Yellow("✗ Cancelled"))
				return nil
			}
		}

		err = process.KillProcess(pid, sig)
		if err != nil {
			return fmt.Errorf("%s failed to kill process %d: %w", ui.Red("✗"), pid, err)
		}

		fmt.Printf("%s Successfully sent %s to PID %d\n", ui.Green("✓"), killSignal, pid)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(killCmd)
	killCmd.Flags().StringVarP(&killSignal, "signal", "s", "SIGTERM", "Signal to send (SIGTERM or SIGKILL)")
	killCmd.Flags().BoolVarP(&killForce, "force", "f", false, "Force kill even for critical processes")
}
