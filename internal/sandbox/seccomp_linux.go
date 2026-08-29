//go:build linux

package sandbox

import (
	"fmt"
	"runtime"
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	seccompSetModeFilter   = 1
	seccompFilterFlagTsync = 1

	auditArchX86_64  = 0xc000003e
	auditArchAARCH64 = 0xc00000b7
	auditArchI386    = 0x40000003
	auditArchARM     = 0x40000028

	bpfLd  = 0x00
	bpfW   = 0x00
	bpfAbs = 0x20
	bpfJmp = 0x05
	bpfJeq = 0x10
	bpfK   = 0x00
	bpfRet = 0x06

	seccompRetKillProcess = 0x80000000
	seccompRetAllow       = 0x7fff0000
	seccompRetErrno       = 0x00050000

	seccompDataNr   = 0
	seccompDataArch = 4
)

type sockFilter struct {
	Code uint16
	Jt   uint8
	Jf   uint8
	K    uint32
}

type sockFprog struct {
	Len    uint16
	Filter *sockFilter
}

func applySeccomp() error {
	denied := deniedSyscalls()
	arch := nativeAuditArch()
	if arch == 0 {
		return fmt.Errorf("seccomp: unsupported architecture %s", runtime.GOARCH)
	}

	prog := make([]sockFilter, 0, 16+len(denied)*2)
	// Validate architecture.
	prog = append(prog,
		bpfLoad(seccompDataArch),
		bpfJumpIf(arch, 1, 0),
		bpfReturn(seccompRetKillProcess),
	)
	// Load syscall number.
	prog = append(prog, bpfLoad(seccompDataNr))

	for _, nr := range denied {
		// if nr == K: return EPERM; else continue
		prog = append(prog,
			bpfJumpIf(uint32(nr), 0, 1), // #nosec G115 -- syscall numbers from fixed unix constants
			bpfReturn(seccompRetErrno|uint32(unix.EPERM)),
		)
	}
	prog = append(prog, bpfReturn(seccompRetAllow))

	fprog := sockFprog{
		Len:    uint16(len(prog)), // #nosec G115 -- filter length fits BPF program size
		Filter: &prog[0],
	}

	_, _, errno := unix.Syscall(
		unix.SYS_SECCOMP,
		seccompSetModeFilter,
		seccompFilterFlagTsync,
		uintptr(unsafe.Pointer(&fprog)), // #nosec G103 -- required for seccomp filter install
	)
	if errno != 0 {
		// Fallback without TSYNC for older kernels.
		_, _, errno = unix.Syscall(
			unix.SYS_SECCOMP,
			seccompSetModeFilter,
			0,
			uintptr(unsafe.Pointer(&fprog)), // #nosec G103 -- required for seccomp filter install
		)
		if errno != 0 {
			return fmt.Errorf("seccomp: %w", errno)
		}
	}
	return nil
}

func bpfLoad(offset uint32) sockFilter {
	return sockFilter{Code: bpfLd | bpfW | bpfAbs, K: offset}
}

func bpfJumpIf(k uint32, jt, jf uint8) sockFilter {
	return sockFilter{Code: bpfJmp | bpfJeq | bpfK, Jt: jt, Jf: jf, K: k}
}

func bpfReturn(v uint32) sockFilter {
	return sockFilter{Code: bpfRet | bpfK, K: v}
}

func nativeAuditArch() uint32 {
	switch runtime.GOARCH {
	case "amd64":
		return auditArchX86_64
	case "arm64":
		return auditArchAARCH64
	case "386":
		return auditArchI386
	case "arm":
		return auditArchARM
	default:
		return 0
	}
}

func deniedSyscalls() []int {
	// Denylist of high-risk calls. Keep short so a Go HTTP server keeps working.
	base := []string{
		"mount", "umount2", "pivot_root", "swapon", "swapoff",
		"reboot", "sethostname", "setdomainname",
		"init_module", "finit_module", "delete_module",
		"kexec_load", "kexec_file_load",
		"ptrace", "process_vm_readv", "process_vm_writev",
		"userfaultfd", "perf_event_open",
		"bpf", "open_by_handle_at",
	}
	out := make([]int, 0, len(base))
	for _, name := range base {
		if nr, ok := syscallNumber(name); ok {
			out = append(out, nr)
		}
	}
	return out
}

func syscallNumber(name string) (int, bool) {
	// Map by GOARCH using generated numbers from x/sys/unix where available.
	switch runtime.GOARCH {
	case "amd64":
		return amd64Syscall(name)
	case "arm64":
		return arm64Syscall(name)
	default:
		return 0, false
	}
}

func amd64Syscall(name string) (int, bool) {
	m := map[string]int{
		"mount":             165,
		"umount2":           166,
		"pivot_root":        155,
		"swapon":            167,
		"swapoff":           168,
		"reboot":            169,
		"sethostname":       170,
		"setdomainname":     171,
		"init_module":       175,
		"finit_module":      313,
		"delete_module":     176,
		"kexec_load":        246,
		"kexec_file_load":   320,
		"ptrace":            101,
		"process_vm_readv":  310,
		"process_vm_writev": 311,
		"userfaultfd":       323,
		"perf_event_open":   298,
		"bpf":               321,
		"open_by_handle_at": 304,
	}
	nr, ok := m[name]
	return nr, ok
}

func arm64Syscall(name string) (int, bool) {
	m := map[string]int{
		"mount":             40,
		"umount2":           39,
		"pivot_root":        41,
		"swapon":            224,
		"swapoff":           225,
		"reboot":            142,
		"sethostname":       161,
		"setdomainname":     162,
		"init_module":       105,
		"finit_module":      273,
		"delete_module":     106,
		"kexec_load":        104,
		"kexec_file_load":   294,
		"ptrace":            117,
		"process_vm_readv":  270,
		"process_vm_writev": 271,
		"userfaultfd":       282,
		"perf_event_open":   241,
		"bpf":               280,
		"open_by_handle_at": 265,
	}
	nr, ok := m[name]
	return nr, ok
}
