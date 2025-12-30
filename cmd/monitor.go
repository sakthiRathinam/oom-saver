package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"sakthiRathinam/oom-saver/pkg/process"
	"sakthiRathinam/oom-saver/pkg/ui"
)

var (
	monitorInterval time.Duration
	monitorLimit    int
)

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Monitor processes continuously",
	Long:  `Continuously monitor running processes and automatically kill zombie processes.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.PrintHeader("👁️  PROCESS MONITOR")
		fmt.Printf("\n%s Monitoring processes every %s. Press Ctrl+C to exit.\n", ui.Cyan("ℹ️"), ui.Bold(monitorInterval.String()))

		ticker := time.NewTicker(monitorInterval)
		defer ticker.Stop()

		updateProcessList()

		for range ticker.C {
			updateProcessList()
		}

		return nil
	},
}

func updateProcessList() {
	ui.PrintTimestamp()

	processes, err := process.GetAllRunningProcesses()
	if err != nil {
		fmt.Printf("%s Error fetching processes: %v\n", ui.Red("✗"), err)
		return
	}

	processes, err = process.KillProcessIfZombie(processes)
	if err != nil {
		fmt.Printf("%s Error killing zombies: %v\n", ui.Red("✗"), err)
		return
	}

	ui.PrintProcessTable(processes, monitorLimit)
	fmt.Println()
}

func init() {
	rootCmd.AddCommand(monitorCmd)
	monitorCmd.Flags().DurationVarP(&monitorInterval, "interval", "i", 5*time.Second, "Monitoring interval")
	monitorCmd.Flags().IntVarP(&monitorLimit, "limit", "l", 200, "Maximum number of processes to display")
}
