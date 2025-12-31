package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"sakthiRathinam/oom-saver/pkg/process"
	"sakthiRathinam/oom-saver/pkg/ui"
)

var (
	monitorInterval        time.Duration
	monitorLimit           int
	monitorAutoKillAll     bool
	monitorNoAutoKill      bool
	monitorKillUserProcs   bool
	monitorKillBrowsers    bool
	monitorKillSafe        bool
	monitorKillImportant   bool
	monitorMinOOMScore     int
	monitorZombiesOnly     bool
	monitorUseConfig       bool
)

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Monitor processes continuously",
	Long:  `Continuously monitor running processes and automatically kill safe zombie processes.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.PrintHeader("👁️  PROCESS MONITOR")
		fmt.Printf("\n%s Monitoring processes every %s. Press Ctrl+C to exit.\n", ui.Cyan("ℹ️"), ui.Bold(monitorInterval.String()))

		if monitorNoAutoKill {
			fmt.Printf("%s Auto-kill is DISABLED\n", ui.Yellow("⚠️"))
		} else if monitorUseConfig {
			fmt.Printf("%s Using custom cleanup configuration:\n", ui.Green("✓"))
			if monitorKillUserProcs {
				fmt.Printf("   • User processes (UID >= 1000)\n")
			}
			if monitorKillBrowsers {
				fmt.Printf("   • Browser processes\n")
			}
			if monitorKillSafe {
				fmt.Printf("   • Safe level processes\n")
			}
			if monitorKillImportant {
				fmt.Printf("   • Important level processes\n")
			}
			if monitorMinOOMScore > 0 {
				fmt.Printf("   • Processes with OOM score >= %d\n", monitorMinOOMScore)
			}
			if monitorZombiesOnly {
				fmt.Printf("   • Zombies only mode enabled\n")
			}
		} else if monitorAutoKillAll {
			fmt.Printf("%s Auto-killing ALL zombies (including critical/important)\n", ui.Red("⚠️"))
		} else {
			fmt.Printf("%s Auto-killing only SAFE zombies\n", ui.Green("✓"))
		}

		ticker := time.NewTicker(monitorInterval)
		defer ticker.Stop()

		killProcessToCleanUPMEM()

		for range ticker.C {
			killProcessToCleanUPMEM()
		}

		return nil
	},
}

func killProcessToCleanUPMEM() {
	ui.PrintTimestamp()

	processes, err := process.GetAllRunningProcesses()
	if err != nil {
		fmt.Printf("%s Error fetching processes: %v\n", ui.Red("✗"), err)
		return
	}

	if !monitorNoAutoKill {
		if monitorUseConfig {
			// Use custom cleanup configuration
			config := process.CleanupConfig{
				KillUserProcesses:  monitorKillUserProcs,
				KillBrowsers:       monitorKillBrowsers,
				KillSafeLevel:      monitorKillSafe,
				KillImportantLevel: monitorKillImportant,
				MinOOMScore:        monitorMinOOMScore,
				KillZombiesOnly:    monitorZombiesOnly,
			}
			processes, err = process.KillProcessWithConfig(processes, config)
			if err != nil {
				fmt.Printf("%s Error killing processes: %v\n", ui.Red("✗"), err)
				return
			}
		} else {
			// Use legacy zombie killing
			processes, err = process.KillProcessIfZombie(processes, monitorAutoKillAll)
			if err != nil {
				fmt.Printf("%s Error killing zombies: %v\n", ui.Red("✗"), err)
				return
			}
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

	// Custom cleanup configuration flags
	monitorCmd.Flags().BoolVar(&monitorUseConfig, "use-config", false, "Enable custom cleanup configuration")
	monitorCmd.Flags().BoolVar(&monitorKillUserProcs, "kill-user-processes", false, "Auto-kill user processes (UID >= 1000)")
	monitorCmd.Flags().BoolVar(&monitorKillBrowsers, "kill-browsers", false, "Auto-kill browser processes")
	monitorCmd.Flags().BoolVar(&monitorKillSafe, "kill-safe", false, "Auto-kill safe level processes")
	monitorCmd.Flags().BoolVar(&monitorKillImportant, "kill-important", false, "Auto-kill important level processes")
	monitorCmd.Flags().IntVar(&monitorMinOOMScore, "min-oom-score", 0, "Minimum OOM score to kill (0 = disabled)")
	monitorCmd.Flags().BoolVar(&monitorZombiesOnly, "zombies-only", false, "Only kill zombie processes (ignore running processes)")
}
