package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"sakthiRathinam/oom-saver/pkg/process"
	"sakthiRathinam/oom-saver/pkg/ui"
)

var (
	listLimit  int
	listStatus string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all running processes",
	Long:  `Display a snapshot of all currently running processes with their status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.PrintHeader("🔍 PROCESS LIST")
		ui.PrintTimestamp()

		bar := ui.CreateProgressBar(100, "Scanning processes...")
		go func() {
			for i := 0; i < 100; i++ {
				bar.Add(1)
				time.Sleep(10 * time.Millisecond)
			}
		}()

		processes, err := process.GetAllRunningProcesses()
		bar.Finish()

		if err != nil {
			return fmt.Errorf("failed to get processes: %w", err)
		}

		if listStatus != "" {
			var filtered []process.Process
			for _, p := range processes {
				if p.Status == listStatus {
					filtered = append(filtered, p)
				}
			}
			processes = filtered
		}

		ui.PrintProcessTable(processes, listLimit)
		fmt.Println()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().IntVarP(&listLimit, "limit", "l", 200, "Maximum number of processes to display")
	listCmd.Flags().StringVarP(&listStatus, "status", "s", "", "Filter by status (e.g., zombie, running, sleeping)")
}
