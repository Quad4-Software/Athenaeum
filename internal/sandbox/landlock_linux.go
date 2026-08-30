//go:build linux

package sandbox

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/landlock-lsm/go-landlock/landlock"
	llsys "github.com/landlock-lsm/go-landlock/landlock/syscall"
	"golang.org/x/sys/unix"
)

const landlockRequestABI = 9

func applyPlatform(cfg Config, log *slog.Logger) (Status, error) {
	st := Status{Mode: cfg.Mode}

	if !cfg.EnableLandlock && !cfg.EnableSeccomp {
		st.Landlock = Component{State: StateDisabled, Detail: "toggle off"}
		st.Seccomp = Component{State: StateDisabled, Detail: "toggle off"}
		log.Info("sandbox enabled but landlock and seccomp both disabled")
		return st, nil
	}

	if cfg.EnableLandlock {
		ll, err := applyLandlock(cfg)
		if err != nil {
			ll.Err = err.Error()
			st.Landlock = ll
			if herr := handle(cfg, log, "landlock", err); herr != nil {
				return st, herr
			}
		} else {
			st.Landlock = ll
		}
	} else {
		st.Landlock = Component{State: StateDisabled, Detail: "toggle off"}
	}

	if cfg.EnableSeccomp {
		sc, err := applySeccompWithNoNewPrivs()
		if err != nil {
			sc.Err = err.Error()
			st.Seccomp = sc
			if herr := handle(cfg, log, "seccomp", err); herr != nil {
				return st, herr
			}
		} else {
			st.Seccomp = sc
		}
	} else {
		st.Seccomp = Component{State: StateDisabled, Detail: "toggle off"}
	}
	return st, nil
}

func applySeccompWithNoNewPrivs() (Component, error) {
	denied := deniedSyscalls()
	detail := fmt.Sprintf("denylist=%d no_new_privs arch=%s", len(denied), runtime.GOARCH)
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err := unix.Prctl(unix.PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0); err != nil {
		return Component{State: StateSkipped, Detail: detail},
			fmt.Errorf("prctl(NO_NEW_PRIVS): %w", err)
	}
	if err := applySeccomp(); err != nil {
		hint := err.Error()
		if isLikelyContainerSeccompBlock(err) {
			hint += " (common inside Docker/Podman when an outer seccomp profile is already active)"
		}
		return Component{State: StateSkipped, Detail: detail}, fmt.Errorf("%s", hint)
	}
	return Component{State: StateApplied, Detail: detail}, nil
}

func isLikelyContainerSeccompBlock(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "operation not permitted") ||
		strings.Contains(msg, "permission denied") ||
		strings.Contains(msg, "eperm") ||
		strings.Contains(msg, "eacces")
}

func handle(cfg Config, log *slog.Logger, what string, err error) error {
	if cfg.Mode == ModeStrict {
		return fmt.Errorf("%s: %w", what, err)
	}
	log.Warn("sandbox "+what+" skipped", "err", err)
	return nil
}

func applyLandlock(cfg Config) (Component, error) {
	kernelABI, abiErr := llsys.LandlockGetABIVersion()
	abiDetail := landlockABIDetail(kernelABI, abiErr, cfg.Mode)

	roDirs, roFiles := classifyPaths(append([]string{
		"/usr", "/bin", "/lib", "/lib64",
		"/etc/ssl", "/etc/ca-certificates",
		"/proc/self", "/proc/sys/kernel/random",
		"/dev",
		"/etc/resolv.conf", "/etc/hosts", "/etc/nsswitch.conf", "/etc/localtime",
	}, cfg.ReadOnlyPaths...))

	rwDirs, rwFiles := classifyPaths(cfg.ReadWritePaths)
	if tmp := os.TempDir(); tmp != "" {
		d, f := classifyPaths([]string{tmp})
		rwDirs = append(rwDirs, d...)
		rwFiles = append(rwFiles, f...)
	}

	roDirs = uniqueAbs(roDirs)
	roFiles = uniqueAbs(roFiles)
	rwDirs = uniqueAbs(rwDirs)
	rwFiles = uniqueAbs(rwFiles)

	rules := make([]landlock.Rule, 0, 4)
	if len(roDirs) > 0 {
		rules = append(rules, landlock.RODirs(roDirs...).IgnoreIfMissing())
	}
	if len(roFiles) > 0 {
		rules = append(rules, landlock.ROFiles(roFiles...).IgnoreIfMissing())
	}
	if len(rwDirs) > 0 {
		rules = append(rules, landlock.RWDirs(rwDirs...).IgnoreIfMissing())
	}
	if len(rwFiles) > 0 {
		rules = append(rules, landlock.RWFiles(rwFiles...).IgnoreIfMissing())
	}
	if len(rules) == 0 {
		return Component{State: StateSkipped, Detail: abiDetail}, fmt.Errorf("no landlock paths to restrict")
	}

	pathDetail := fmt.Sprintf("ro=%d rw=%d", len(roDirs)+len(roFiles), len(rwDirs)+len(rwFiles))
	detail := abiDetail + " " + pathDetail

	llCfg := landlock.V9.BestEffort()
	if cfg.Mode == ModeStrict {
		llCfg = landlock.V9
	}
	if err := llCfg.RestrictPaths(rules...); err != nil {
		return Component{State: StateSkipped, Detail: detail}, err
	}
	return Component{State: StateApplied, Detail: detail}, nil
}

func landlockABIDetail(kernelABI int, abiErr error, mode Mode) string {
	effort := "best-effort"
	if mode == ModeStrict {
		effort = "strict"
	}
	if abiErr != nil {
		return fmt.Sprintf("request=V%d %s kernel=unavailable", landlockRequestABI, effort)
	}
	effective := min(kernelABI, landlockRequestABI)
	if effective < 1 {
		return fmt.Sprintf("request=V%d %s kernel=none", landlockRequestABI, effort)
	}
	return fmt.Sprintf("request=V%d effective=V%d %s kernel_abi=%d",
		landlockRequestABI, effective, effort, kernelABI)
}

func classifyPaths(paths []string) (dirs, files []string) {
	for _, p := range paths {
		if p == "" {
			continue
		}
		abs, err := filepath.Abs(p)
		if err != nil {
			continue
		}
		abs = filepath.Clean(abs)
		st, err := os.Stat(abs)
		if err != nil {
			dirs = append(dirs, abs)
			continue
		}
		if st.IsDir() {
			dirs = append(dirs, abs)
		} else {
			files = append(files, abs)
		}
	}
	return dirs, files
}

func uniqueAbs(paths []string) []string {
	seen := make(map[string]struct{}, len(paths))
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		if p == "" {
			continue
		}
		abs, err := filepath.Abs(p)
		if err != nil {
			continue
		}
		abs = filepath.Clean(abs)
		if _, ok := seen[abs]; ok {
			continue
		}
		seen[abs] = struct{}{}
		out = append(out, abs)
	}
	return out
}
