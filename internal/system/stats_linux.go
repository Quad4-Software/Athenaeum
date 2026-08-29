//go:build linux

package system

import (
	"bufio"
	"math"
	"math/bits"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func memoryUsage() (total, used int64) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return runtimeMemFallback()
	}
	defer f.Close()
	var memTotal, memAvailable int64
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "MemTotal:") {
			memTotal = parseKB(line)
		} else if strings.HasPrefix(line, "MemAvailable:") {
			memAvailable = parseKB(line)
		}
	}
	if memTotal <= 0 {
		return 0, 0
	}
	if memAvailable <= 0 {
		return memTotal, memTotal / 2
	}
	return memTotal, memTotal - memAvailable
}

func parseKB(line string) int64 {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return 0
	}
	kb, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return 0
	}
	return kb * 1024
}

func cpuUsage() float64 {
	idle1, total1 := readCPUStat()
	if total1 == 0 {
		return 0
	}
	time.Sleep(120 * time.Millisecond)
	idle2, total2 := readCPUStat()
	if total2 <= total1 {
		return 0
	}
	idleDelta := float64(idle2 - idle1)
	totalDelta := float64(total2 - total1)
	return (1 - idleDelta/totalDelta) * 100
}

func readCPUStat() (idle, total uint64) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0, 0
	}
	line := strings.SplitN(string(data), "\n", 2)[0]
	fields := strings.Fields(line)
	if len(fields) < 5 || fields[0] != "cpu" {
		return 0, 0
	}
	for i := 1; i < len(fields); i++ {
		v, err := strconv.ParseUint(fields[i], 10, 64)
		if err != nil {
			return 0, 0
		}
		total += v
		if i == 4 {
			idle = v
		}
	}
	return idle, total
}

func diskUsage(path string) (DiskUsage, bool) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return DiskUsage{}, false
	}
	total := statBlocksBytes(st.Blocks, int64(st.Bsize))
	avail := statBlocksBytes(st.Bavail, int64(st.Bsize))
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

func runtimeMemFallback() (total, used int64) {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	return uint64ToInt64(ms.Sys), uint64ToInt64(ms.Alloc)
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
	if bsize > math.MaxInt64 {
		return math.MaxInt64
	}
	bsizeU := uint64(bsize)
	hi, lo := bits.Mul64(blocks, bsizeU)
	if hi > 0 || lo > math.MaxInt64 {
		return math.MaxInt64
	}
	return int64(lo)
}
