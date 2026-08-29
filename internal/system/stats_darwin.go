//go:build darwin

package system

import (
	"math"
	"math/bits"
	"runtime"
	"syscall"
)

func memoryUsage() (total, used int64) {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	total = uint64ToInt64(ms.Sys)
	used = uint64ToInt64(ms.Alloc)
	return total, used
}

func cpuUsage() float64 {
	return 0
}

func diskUsage(path string) (DiskUsage, bool) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return DiskUsage{}, false
	}
	bsize := int64(st.Bsize)
	total := statBlocksBytes(st.Blocks, bsize)
	avail := statBlocksBytes(st.Bavail, bsize)
	used := total - avail
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

func statBlocksBytes(blocks uint64, bsize int64) int64 {
	if bsize <= 0 || blocks == 0 {
		return 0
	}
	bsizeU := uint64(bsize)
	hi, lo := bits.Mul64(blocks, bsizeU)
	if hi > 0 || lo > math.MaxInt64 {
		return math.MaxInt64
	}
	return int64(lo)
}
