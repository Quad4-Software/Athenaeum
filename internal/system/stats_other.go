//go:build !linux && !darwin && !windows

package system

import (
	"math"
	"runtime"
)

func memoryUsage() (total, used int64) {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	return uint64ToInt64(ms.Sys), uint64ToInt64(ms.Alloc)
}

func cpuUsage() float64 {
	return 0
}

func diskUsage(path string) (DiskUsage, bool) {
	_ = path
	return DiskUsage{}, false
}

func uint64ToInt64(v uint64) int64 {
	if v > math.MaxInt64 {
		return math.MaxInt64
	}
	return int64(v)
}
