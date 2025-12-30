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

var killSignal string

var killCmd = &cobra.Command{
	Use:   "kill <PID>",
	Short: "Kill a specific process by PID",
	Long:  `Send a signal to kill a specific process. Default signal is SIGTERM.`,
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

		fmt.Printf("%s About to send %s to PID %d\n", ui.Yellow("⚠️"), ui.Bold(killSignal), pid)
		fmt.Print("Continue? (y/N): ")

		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "y" && response != "yes" {
			fmt.Println(ui.Yellow("✗ Cancelled"))
			return nil
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
}
