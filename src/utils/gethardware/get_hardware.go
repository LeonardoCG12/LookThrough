package gethardware

import (
	"fmt"
	"runtime"
)

const (
	minimumAutomaticWorkers = 4
	maximumAutomaticWorkers = 32
	MaximumManualWorkers    = 128
)

func CalculateDynamicWorkerLimit() int {
	workers := runtime.GOMAXPROCS(0) * 2

	if workers < minimumAutomaticWorkers {
		return minimumAutomaticWorkers
	}

	if workers > maximumAutomaticWorkers {
		return maximumAutomaticWorkers
	}

	return workers
}

func ResolveWorkerLimit(requested int) (int, error) {
	if requested == 0 {
		return CalculateDynamicWorkerLimit(), nil
	}

	if requested < 0 {
		return 0, fmt.Errorf("worker count must be zero for automatic mode or a positive number")
	}

	if requested > MaximumManualWorkers {
		return 0, fmt.Errorf("worker count cannot exceed %d", MaximumManualWorkers)
	}

	return requested, nil
}
