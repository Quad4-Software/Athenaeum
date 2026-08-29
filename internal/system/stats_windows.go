//go:build windows

package system

import (
	"math"
	"runtime"

	"golang.org/x/sys/windows"
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
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return DiskUsage{}, false
	}
	var freeBytesAvailable, totalNumberOfBytes, totalNumberOfFreeBytes uint64
	err = windows.GetDiskFreeSpaceEx(p, &freeBytesAvailable, &totalNumberOfBytes, &totalNumberOfFreeBytes)
	if err != nil {
		return DiskUsage{}, false
	}
	total := uint64ToInt64(totalNumberOfBytes)
	avail := uint64ToInt64(freeBytesAvailable)
	used := total - avail
	if used < 0 {
		used = 0
	}
	pct := 0.0
	if total > 0 {
		pct = float64(used) / float64(total) * 100
	}
	return DiskUsage{
		Path:      path,
		Total:     total,
		Used:      used,
		Available: avail,
		Percent:   pct,
	}, true
}

func uint64ToInt64(v uint64) int64 {
	if v > math.MaxInt64 {
		return math.MaxInt64
	}
	return int64(v)
}
