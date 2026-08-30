package system

// DiskUsage describes free space for a filesystem path.
type DiskUsage struct {
	Path      string  `json:"path"`
	Total     int64   `json:"total"`
	Used      int64   `json:"used"`
	Available int64   `json:"available"`
	Percent   float64 `json:"percent"`
}

// Stats is the host resource snapshot returned to the admin UI.
type Stats struct {
	Version    string         `json:"version"`
	WebVersion string         `json:"webVersion"`
	CPUPercent float64        `json:"cpuPercent"`
	MemUsed    int64          `json:"memUsed"`
	MemTotal   int64          `json:"memTotal"`
	MemPercent float64        `json:"memPercent"`
	Disks      []DiskUsage    `json:"disks"`
	Sandbox    map[string]any `json:"sandbox,omitempty"`
}

// ReadStats collects CPU, memory, and disk usage for the given paths.
func ReadStats(paths []string) Stats {
	st := Stats{}
	st.MemTotal, st.MemUsed = memoryUsage()
	if st.MemTotal > 0 {
		st.MemPercent = float64(st.MemUsed) / float64(st.MemTotal) * 100
	}
	st.CPUPercent = cpuUsage()
	for _, path := range paths {
		if path == "" {
			continue
		}
		if du, ok := diskUsage(path); ok {
			st.Disks = append(st.Disks, du)
		}
	}
	if st.Disks == nil {
		st.Disks = []DiskUsage{}
	}
	return st
}

// DiskFree returns available bytes on the filesystem at path, or -1 on error.
func DiskFree(path string) int64 {
	if du, ok := diskUsage(path); ok {
		return du.Available
	}
	return -1
}
