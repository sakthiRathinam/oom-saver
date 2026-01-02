package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "oom-saver",
	Short: "A beautiful OOM killer and process monitor for Linux",
	Long: `oom-saver is a powerful process monitoring and management tool.
It helps you monitor system processes, detect zombies, and prevent out-of-memory situations.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
