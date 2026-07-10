package gethardware

import (
	"runtime"
	"time"
)

func CalculateDynamicLimit() int {
	cores := runtime.NumCPU()
	optimalLimit := cores * 8

	if optimalLimit < 20 {
		return 20
	}
	if optimalLimit > 150 {
		return 150
	}

	return optimalLimit
}

func ThrottleIfMemoryHigh() {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	if mem.Alloc > 500*1024*1024 {
		time.Sleep(50 * time.Millisecond)
	}
}
