// Package sandbox applies optional Linux Landlock and seccomp-bpf restrictions.
package sandbox

import (
	"fmt"
	"log/slog"
)

// Mode selects how strictly sandbox setup failures are treated.
type Mode string

const (
	// ModeOff disables sandboxing.
	ModeOff Mode = "off"
	// ModeTry applies sandboxing when supported and logs failures.
	ModeTry Mode = "try"
	// ModeStrict fails startup when sandboxing cannot be applied on Linux.
	ModeStrict Mode = "strict"
)

// Config controls filesystem and syscall restrictions.
type Config struct {
	Mode Mode
	// ReadWritePaths are directories granted read+write Landlock access.
	ReadWritePaths []string
	// ReadOnlyPaths are paths granted read-only Landlock access.
	ReadOnlyPaths []string
	// EnableLandlock applies filesystem Landlock rules when true.
	EnableLandlock bool
	// EnableSeccomp applies a dangerous-syscall denylist when true.
	EnableSeccomp bool
}

// ParseMode interprets --sandbox values.
func ParseMode(s string) (Mode, error) {
	switch s {
	case "", "off", "false", "0", "no":
		return ModeOff, nil
	case "try", "auto":
		return ModeTry, nil
	case "strict", "on", "true", "1", "yes":
		return ModeStrict, nil
	default:
		return ModeOff, fmt.Errorf("invalid sandbox mode %q (want off, try, or strict)", s)
	}
}

// Apply installs Landlock and/or seccomp according to cfg.
// The returned Status always describes the resulting configuration.
func Apply(cfg Config, log *slog.Logger) (Status, error) {
	if log == nil {
		log = slog.Default()
	}
	if cfg.Mode == ModeOff {
		st := Status{
			Mode:     ModeOff,
			Landlock: Component{State: StateOff},
			Seccomp:  Component{State: StateOff},
		}
		return st, nil
	}
	return applyPlatform(cfg, log)
}
