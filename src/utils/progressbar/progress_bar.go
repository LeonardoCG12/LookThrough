package progressbar

import (
	"fmt"
	"strings"
)

func PrintProgressBar(current, total int64) {
	if total <= 0 {
		return
	}

	if current < 0 {
		current = 0
	}

	if current > total {
		current = total
	}

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
