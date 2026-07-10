package progressbar

import (
	"fmt"
	"strings"
)

func PrintProgressBar(current, total int32) {
	if total == 0 {
		return
	}

	updateInterval := int32(1)
	if total > 1000 {
		updateInterval = 10
	} else if total > 100 {
		updateInterval = 5
	}

	if current%updateInterval == 0 || current == total || current == 1 {
		percent := float64(current) / float64(total) * 100
		barLength := 40
		completed := int(float64(barLength) * float64(current) / float64(total))

		bar := strings.Repeat("=", completed)
		if completed < barLength {
			bar += ">"
			bar += strings.Repeat("-", barLength-completed-1)
		}

		fmt.Printf("\r\033[K[+] PROGRESS: [%s] %.1f%% (%d/%d)", bar, percent, current, total)
	}
}
