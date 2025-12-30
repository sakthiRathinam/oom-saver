package ui

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
	"sakthiRathinam/oom-saver/pkg/process"
)

var (
	Green  = color.New(color.FgGreen).SprintFunc()
	Yellow = color.New(color.FgYellow).SprintFunc()
	Red    = color.New(color.FgRed).SprintFunc()
	Cyan   = color.New(color.FgCyan).SprintFunc()
	Bold   = color.New(color.Bold).SprintFunc()
)

func GetStatusColor(status string) func(a ...interface{}) string {
	switch status {
	case "running":
		return Green
	case "zombie", "dead":
		return Red
	case "sleeping", "idle", "disk-sleep":
		return Yellow
	default:
		return fmt.Sprint
	}
}

func PrintHeader(title string) {
	fmt.Println()
	fmt.Println(Cyan("═══════════════════════════════════════════════════════════════════"))
	fmt.Printf("  %s\n", Bold(title))
	fmt.Println(Cyan("═══════════════════════════════════════════════════════════════════"))
}

func PrintTimestamp() {
	fmt.Printf("\n%s %s\n", Cyan("⏰ Timestamp:"), time.Now().Format("2006-01-02 15:04:05"))
}

func PrintProcessTable(processes []process.Process, limit int) {
	if len(processes) == 0 {
		fmt.Println(Yellow("No processes found"))
		return
	}

	if limit > len(processes) || limit <= 0 {
		limit = len(processes)
	}

	fmt.Printf("\n%s %s\n", Cyan("📊 Total processes:"), Bold(fmt.Sprintf("%d", len(processes))))
	fmt.Println(Cyan("───────────────────────────────────────────────────────────────────"))
	fmt.Printf("%-8s %-30s %-15s\n", "PID", "NAME", "STATUS")
	fmt.Println(Cyan("───────────────────────────────────────────────────────────────────"))

	for i := 0; i < limit; i++ {
		p := processes[i]
		colorFunc := GetStatusColor(p.Status)
		fmt.Printf("%-8d %-30s %s\n", p.PID, p.Name, colorFunc(p.Status))
	}

	if len(processes) > limit {
		fmt.Printf("\n%s %d more processes...\n", Yellow("⋯"), len(processes)-limit)
	}
}

func CreateProgressBar(max int, description string) *progressbar.ProgressBar {
	return progressbar.NewOptions(max,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWidth(40),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerPadding: "░",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)
}

func PrintStats(processes []process.Process) {
	stats := make(map[string]int)
	for _, p := range processes {
		stats[p.Status]++
	}

	PrintHeader("📈 PROCESS STATISTICS")
	PrintTimestamp()

	fmt.Printf("\n%s %s\n", Bold("Total Processes:"), Cyan(fmt.Sprintf("%d", len(processes))))
	fmt.Println(Cyan("───────────────────────────────────────────────────────────────────"))

	for status, count := range stats {
		colorFunc := GetStatusColor(status)
		fmt.Printf("  %-20s %s\n", colorFunc(status+":"), Bold(fmt.Sprintf("%d", count)))
	}
}
