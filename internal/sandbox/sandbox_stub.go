//go:build !linux

package sandbox

import (
	"fmt"
	"log/slog"
	"runtime"
)

func applyPlatform(cfg Config, log *slog.Logger) (Status, error) {
	detail := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
	msg := fmt.Sprintf("sandbox requested but unsupported on %s", detail)
	st := Status{
		Mode: cfg.Mode,
		Landlock: Component{
			State:  StateUnsupported,
			Detail: detail,
		},
		Seccomp: Component{
			State:  StateUnsupported,
			Detail: detail,
		},
	}
	if cfg.Mode == ModeStrict {
		return st, fmt.Errorf("%s", msg)
	}
	log.Info(msg)
	return st, nil
}
