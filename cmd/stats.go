package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"sakthiRathinam/oom-saver/pkg/process"
	"sakthiRathinam/oom-saver/pkg/ui"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show process statistics",
	Long:  `Display summary statistics about running processes grouped by status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		processes, err := process.GetAllRunningProcesses()
		if err != nil {
			return fmt.Errorf("failed to get processes: %w", err)
		}

		ui.PrintStats(processes)
		fmt.Println()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
