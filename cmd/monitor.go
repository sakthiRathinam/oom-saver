package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"sakthiRathinam/oom-saver/pkg/process"
	"sakthiRathinam/oom-saver/pkg/ui"
)

var (
	monitorInterval       time.Duration
	monitorLimit          int
	monitorAutoKillAll    bool
	monitorNoAutoKill     bool
)

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Monitor processes continuously",
	Long:  `Continuously monitor running processes and automatically kill safe zombie processes.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.PrintHeader("👁️  PROCESS MONITOR")
		fmt.Printf("\n%s Monitoring processes every %s. Press Ctrl+C to exit.\n", ui.Cyan("ℹ️"), ui.Bold(monitorInterval.String()))

		if monitorNoAutoKill {
			fmt.Printf("%s Zombie auto-kill is DISABLED\n", ui.Yellow("⚠️"))
		} else if monitorAutoKillAll {
			fmt.Printf("%s Auto-killing ALL zombies (including critical/important)\n", ui.Red("⚠️"))
		} else {
			fmt.Printf("%s Auto-killing only SAFE zombies\n", ui.Green("✓"))
		}

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

	if !monitorNoAutoKill {
		processes, err = process.KillProcessIfZombie(processes, monitorAutoKillAll)
		if err != nil {
			fmt.Printf("%s Error killing zombies: %v\n", ui.Red("✗"), err)
			return
		}
	}

	ui.PrintProcessTable(processes, monitorLimit)
	fmt.Println()
}

func init() {
	rootCmd.AddCommand(monitorCmd)
	monitorCmd.Flags().DurationVarP(&monitorInterval, "interval", "i", 5*time.Second, "Monitoring interval")
	monitorCmd.Flags().IntVarP(&monitorLimit, "limit", "l", 200, "Maximum number of processes to display")
	monitorCmd.Flags().BoolVar(&monitorAutoKillAll, "auto-kill-all-zombies", false, "Auto-kill all zombies including critical/important")
	monitorCmd.Flags().BoolVar(&monitorNoAutoKill, "no-auto-kill", false, "Disable automatic zombie killing")
}
