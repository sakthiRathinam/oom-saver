package ui

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
	"sakthiRathinam/oom-saver/pkg/process"
)

var (
	Green     = color.New(color.FgGreen).SprintFunc()
	Yellow    = color.New(color.FgYellow).SprintFunc()
	Red       = color.New(color.FgRed).SprintFunc()
	Cyan      = color.New(color.FgCyan).SprintFunc()
	Bold      = color.New(color.Bold).SprintFunc()
	White     = color.New(color.FgWhite).SprintFunc()
	RedBold   = color.New(color.FgRed, color.Bold).SprintFunc()
	GreenBold = color.New(color.FgGreen, color.Bold).SprintFunc()
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

func GetSafetyColor(safetyLevel string) func(a ...interface{}) string {
	switch safetyLevel {
	case "critical":
		return RedBold
	case "important":
		return Yellow
	case "safe":
		return Green
	case "unknown":
		return White
	default:
		return fmt.Sprint
	}
}

func GetSafetyIcon(safetyLevel string) string {
	switch safetyLevel {
	case "critical":
		return "🔴"
	case "important":
		return "🟡"
	case "safe":
		return "🟢"
	case "unknown":
		return "⚪"
	default:
		return "  "
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
	fmt.Println(Cyan("═══════════════════════════════════════════════════════════════════════════════════"))
	fmt.Printf("%-8s %-30s %-15s %-15s\n", "PID", "NAME", "STATUS", "SAFETY")
	fmt.Println(Cyan("═══════════════════════════════════════════════════════════════════════════════════"))

	for i := 0; i < limit; i++ {
		p := processes[i]
		statusColor := GetStatusColor(p.Status)
		safetyColor := GetSafetyColor(p.SafetyLevel)
		safetyIcon := GetSafetyIcon(p.SafetyLevel)

		fmt.Printf("%-8d %-30s %-15s %s %s\n",
			p.PID,
			p.Name,
			statusColor(p.Status),
			safetyIcon,
			safetyColor(p.SafetyLevel))
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
	statusStats := make(map[string]int)
	safetyStats := make(map[string]int)

	for _, p := range processes {
		statusStats[p.Status]++
		safetyStats[p.SafetyLevel]++
	}

	PrintHeader("📈 PROCESS STATISTICS")
	PrintTimestamp()

	fmt.Printf("\n%s %s\n", Bold("Total Processes:"), Cyan(fmt.Sprintf("%d", len(processes))))

	fmt.Println(Cyan("\n━━━ By Status ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	for status, count := range statusStats {
		colorFunc := GetStatusColor(status)
		fmt.Printf("  %-20s %s\n", colorFunc(status+":"), Bold(fmt.Sprintf("%d", count)))
	}

	fmt.Println(Cyan("\n━━━ By Safety Level ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	safetyOrder := []string{"critical", "important", "safe", "unknown"}
	for _, safety := range safetyOrder {
		if count, ok := safetyStats[safety]; ok {
			colorFunc := GetSafetyColor(safety)
			icon := GetSafetyIcon(safety)
			fmt.Printf("  %s %-15s %s\n", icon, colorFunc(safety+":"), Bold(fmt.Sprintf("%d", count)))
		}
	}
}
