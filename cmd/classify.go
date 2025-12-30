package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"sakthiRathinam/oom-saver/pkg/process"
	"sakthiRathinam/oom-saver/pkg/ui"
)

var classifyCmd = &cobra.Command{
	Use:   "classify <PID>",
	Short: "Show detailed classification info for a process",
	Long:  `Display detailed safety classification and process information for a specific PID.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pid, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid PID: %s", args[0])
		}

		proc, err := process.GetProcessByPID(pid)
		if err != nil {
			return fmt.Errorf("%s %w", ui.Red("✗"), err)
		}

		safetyColor := ui.GetSafetyColor(proc.SafetyLevel)
		safetyIcon := ui.GetSafetyIcon(proc.SafetyLevel)
		statusColor := ui.GetStatusColor(proc.Status)

		ui.PrintHeader("🔍 PROCESS CLASSIFICATION")

		fmt.Printf("\n%s\n", ui.Bold("Basic Information"))
		fmt.Println(ui.Cyan("───────────────────────────────────────────────────────────────────"))
		fmt.Printf("  PID:             %s\n", ui.Bold(fmt.Sprintf("%d", proc.PID)))
		fmt.Printf("  Name:            %s\n", ui.Bold(proc.Name))
		fmt.Printf("  Status:          %s\n", statusColor(proc.Status))
		fmt.Printf("  Owner (UID):     %d\n", proc.UID)
		fmt.Printf("  Parent PID:      %d\n", proc.PPID)
		fmt.Printf("  OOM Score:       %d\n", proc.OOMScore)

		fmt.Printf("\n%s\n", ui.Bold("Safety Classification"))
		fmt.Println(ui.Cyan("───────────────────────────────────────────────────────────────────"))
		fmt.Printf("  Safety Level:    %s %s\n", safetyIcon, safetyColor(proc.SafetyLevel))

		fmt.Printf("\n%s\n", ui.Bold("Classification Details"))
		fmt.Println(ui.Cyan("───────────────────────────────────────────────────────────────────"))

		switch proc.SafetyLevel {
		case "critical":
			fmt.Printf("  %s %s\n", ui.RedBold("🔴 CRITICAL PROCESS"), ui.RedBold("- DO NOT KILL"))
			fmt.Println("  This is a system-critical process. Killing it may:")
			fmt.Println("    • Crash the entire system")
			fmt.Println("    • Cause data loss or corruption")
			fmt.Println("    • Require a system reboot")
			fmt.Println()
			fmt.Println("  Reasons for classification:")
			if proc.PID == 1 {
				fmt.Println("    • PID 1 (init/systemd) - system manager")
			}
			if proc.OOMScore < -500 {
				fmt.Printf("    • Very negative OOM score (%d) - kernel protected\n", proc.OOMScore)
			}
			fmt.Println("    • Essential system service")

		case "important":
			fmt.Printf("  %s %s\n", ui.Yellow("🟡 IMPORTANT PROCESS"), ui.Yellow("- Kill with caution"))
			fmt.Println("  This is an important system process. Killing it may:")
			fmt.Println("    • Disrupt system services")
			fmt.Println("    • Affect running applications")
			fmt.Println("    • Require service restart")
			fmt.Println()
			fmt.Println("  Reasons for classification:")
			if proc.UID == 0 {
				fmt.Println("    • Owned by root")
			}
			if proc.PPID == 1 {
				fmt.Println("    • Child of systemd")
			}
			fmt.Println("    • System daemon or important service")

		case "safe":
			fmt.Printf("  %s %s\n", ui.Green("🟢 SAFE TO KILL"), ui.Green("- Can be terminated"))
			fmt.Println("  This process can be safely killed. It is likely:")
			fmt.Println("    • A user application")
			fmt.Println("    • Non-critical to system operation")
			fmt.Println("    • Safe to restart if needed")
			fmt.Println()
			fmt.Println("  Reasons for classification:")
			if proc.Status == "zombie" {
				fmt.Println("    • Zombie process (already dead)")
			}
			if proc.UID >= 1000 {
				fmt.Println("    • Owned by regular user (non-root)")
			}
			if proc.OOMScore > 300 {
				fmt.Printf("    • High OOM score (%d) - kernel considers killable\n", proc.OOMScore)
			}

		case "unknown":
			fmt.Printf("  %s %s\n", ui.White("⚪ UNKNOWN"), ui.White("- Requires investigation"))
			fmt.Println("  This process doesn't clearly fit other categories.")
			fmt.Println("  Manual investigation recommended before killing.")
			fmt.Println()
			fmt.Printf("  UID:       %d\n", proc.UID)
			fmt.Printf("  PPID:      %d\n", proc.PPID)
			fmt.Printf("  OOM Score: %d\n", proc.OOMScore)
		}

		fmt.Println()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(classifyCmd)
}
