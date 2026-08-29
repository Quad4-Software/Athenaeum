// Package config holds runtime configuration for the Athenaeum server.
package config

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"athenaeum/internal/brand"
	"athenaeum/internal/term"
)

// Config is the resolved runtime configuration.
type Config struct {
	Addr                   string
	LibraryDir             string
	DataDir                string
	WebDir                 string
	AdminUser              string
	AdminPass              string
	UploadMaxBytes         int64
	ScanWorkers            int
	SentryDSN              string
	SentryDSNPublic        string
	SentryEnvironment      string
	SentryRelease          string
	SentryTracesSampleRate float64
	LogLevel               string
	LogFile                string
	ColorMode              string
	NoColor                bool
	Sandbox                string
	SandboxLandlock        bool
	SandboxSeccomp         bool
	Demo                   bool
	PprofAddr              string
	DatabaseDriver         string
	DatabaseURL            string

	AltchaEnabled          bool
	AltchaMode             string
	AltchaHMACSecret       string
	AltchaHMACKeySecret    string
	AltchaChallengeURL     string
	AltchaSentinelURL      string
	AltchaVerifyURL        string
	AltchaAPIKeySecret     string
	AltchaCost             int
	AltchaExpiresSecs      int
	AltchaProtect          string
	AltchaWidgetAuto       string
	AltchaWidgetDisplay    string
	AltchaWidgetType       string
	AltchaWidgetTheme      string
	AltchaWidgetLanguage   string
	AltchaWidgetName       string
	AltchaWidgetWorkers    int
	AltchaWidgetHideLogo   bool
	AltchaWidgetHideFooter bool

	PasswordMinLength     int
	PasswordLongLength    int
	PasswordMinKinds      int
	PasswordRequireLower  bool
	PasswordRequireUpper  bool
	PasswordRequireDigit  bool
	PasswordRequireSymbol bool
}

// SentryPublicDSN returns the DSN exposed to the browser, defaulting to SentryDSN.
func (c Config) SentryPublicDSN() string {
	if c.SentryDSNPublic != "" {
		return c.SentryDSNPublic
	}
	return c.SentryDSN
}

// UploadDir returns the directory for in-progress upload parts.
func (c Config) UploadDir() string { return filepath.Join(c.DataDir, "uploads") }

// DBPath returns the sqlite database path inside the data directory.
func (c Config) DBPath() string { return ResolveDBPath(c.DataDir) }

// UsesPostgres reports whether the configured driver is PostgreSQL.
func (c Config) UsesPostgres() bool {
	d := strings.ToLower(strings.TrimSpace(c.DatabaseDriver))
	return d == "postgres" || d == "postgresql" || d == "pg"
}

// ResolveDBPath picks the active database file, preferring athenaeum.db
// and falling back to the legacy reader.db when upgrading an existing install.
func ResolveDBPath(dataDir string) string {
	current := filepath.Join(dataDir, brand.DBFilename)
	if _, err := os.Stat(current); err == nil {
		return current
	}
	legacy := filepath.Join(dataDir, "reader.db")
	if _, err := os.Stat(legacy); err == nil {
		return legacy
	}
	return current
}

// CoverDir returns the directory used to cache extracted cover images.
func (c Config) CoverDir() string { return filepath.Join(c.DataDir, "covers") }

// TempDir returns the directory for short-lived scan downloads.
func (c Config) TempDir() string { return filepath.Join(c.DataDir, "tmp") }

// I18nDir returns the directory for custom locale JSON files.
func (c Config) I18nDir() string { return filepath.Join(c.DataDir, "i18n") }

// LogFileDir returns the parent directory of LogFile when set, else empty.
func (c Config) LogFileDir() string {
	if c.LogFile == "" {
		return ""
	}
	return filepath.Dir(c.LogFile)
}

// Parse reads configuration from command-line flags, falling back to
// environment variables (including optional .env) and sane defaults.
//
// Precedence (lowest to highest): defaults, .env, environment, CLI flags.
func Parse(args []string) (Config, error) {
	if err := LoadEnv(); err != nil {
		return Config{}, err
	}

	fs := flag.NewFlagSet(brand.Name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	addr := fs.String("addr", Env("ATHENAEUM_ADDR", ":8080"), "HTTP listen address")
	library := fs.String("library", Env("ATHENAEUM_LIBRARY", "./library"), "root directory of the book library")
	data := fs.String("data", Env("ATHENAEUM_DATA", "./data"), "directory for the database and caches")
	webDir := fs.String("web-dir", Env("ATHENAEUM_WEB_DIR", ""), "serve frontend from this directory instead of embedded assets")
	adminUser := fs.String("admin-user", Env("ATHENAEUM_ADMIN_USER", ""), "create admin user with this username on startup if no users exist")
	adminPass := fs.String("admin-pass", Env("ATHENAEUM_ADMIN_PASS", ""), "password for the admin user (requires --admin-user)")
	uploadMaxStr := fs.String("upload-max-bytes", Env("ATHENAEUM_UPLOAD_MAX_BYTES", "2147483648"), "maximum upload size in bytes (default 2GB)")
	scanWorkers := fs.String("scan-workers", Env("ATHENAEUM_SCAN_WORKERS", "2"), "parallel library index workers (default 2)")
	sentryDSN := fs.String("sentry-dsn", Env("ATHENAEUM_SENTRY_DSN", ""), "Sentry or GlitchTip DSN for server error reporting")
	sentryDSNPublic := fs.String("sentry-dsn-public", Env("ATHENAEUM_SENTRY_DSN_PUBLIC", ""), "browser DSN for client error reporting (defaults to sentry-dsn)")
	sentryEnv := fs.String("sentry-environment", Env("ATHENAEUM_SENTRY_ENVIRONMENT", ""), "Sentry environment name (e.g. production)")
	sentryRelease := fs.String("sentry-release", Env("ATHENAEUM_SENTRY_RELEASE", ""), "Sentry release name (defaults to app version)")
	sentryTraces := fs.String("sentry-traces-sample-rate", Env("ATHENAEUM_SENTRY_TRACES_SAMPLE_RATE", "0"), "Sentry performance traces sample rate 0.0-1.0")
	logLevel := fs.String("log-level", Env("ATHENAEUM_LOG_LEVEL", "info"), "log level: debug, info, warn, error")
	logFile := fs.String("log-file", Env("ATHENAEUM_LOG_FILE", ""), "append logs to this file in addition to stderr")
	debug := fs.Bool("debug", EnvBool("ATHENAEUM_DEBUG", false), "enable debug logging (shortcut for --log-level=debug)")
	colorMode := fs.String("color", Env("ATHENAEUM_COLOR", "auto"), "CLI color: auto, always, never")
	noColor := fs.Bool("no-color", EnvBool("ATHENAEUM_NO_COLOR", false), "disable CLI color (same as --color=never)")
	sandbox := fs.String("sandbox", Env("ATHENAEUM_SANDBOX", "off"), "Linux sandbox mode: off, try, or strict")
	sandboxLandlock := fs.Bool("sandbox-landlock", EnvBool("ATHENAEUM_SANDBOX_LANDLOCK", true), "enable Landlock filesystem restrictions when sandbox is on")
	sandboxSeccomp := fs.Bool("sandbox-seccomp", EnvBool("ATHENAEUM_SANDBOX_SECCOMP", true), "enable seccomp-bpf syscall denylist when sandbox is on")
	demo := fs.Bool("demo", EnvBool("ATHENAEUM_DEMO", false), "seed a generated demo library with books, audiobooks, and covers")
	pprofAddr := fs.String("pprof", Env("ATHENAEUM_PPROF", ""), "loopback pprof listen address (e.g. 127.0.0.1:6060, empty disables)")
	databaseDriver := fs.String("database-driver", Env("ATHENAEUM_DATABASE_DRIVER", "sqlite"), "database driver: sqlite (default) or postgres")
	databaseURL := fs.String("database-url", Env("ATHENAEUM_DATABASE_URL", ""), "PostgreSQL connection URL (required when --database-driver=postgres)")
	altchaEnabled := fs.Bool("altcha", EnvBool("ATHENAEUM_ALTCHA_ENABLED", false), "require ALTCHA proof-of-work on protected auth forms")
	altchaMode := fs.String("altcha-mode", Env("ATHENAEUM_ALTCHA_MODE", "builtin"), "ALTCHA mode: builtin or sentinel")
	altchaHMAC := fs.String("altcha-hmac-secret", Env("ATHENAEUM_ALTCHA_HMAC_SECRET", ""), "HMAC secret for builtin challenges (auto-persisted when empty)")
	altchaHMACKey := fs.String("altcha-hmac-key-secret", Env("ATHENAEUM_ALTCHA_HMAC_KEY_SECRET", ""), "optional HMAC key signature secret")
	altchaChallengeURL := fs.String("altcha-challenge-url", Env("ATHENAEUM_ALTCHA_CHALLENGE_URL", ""), "widget challenge URL (defaults to /api/altcha/challenge or sentinel URL)")
	altchaSentinelURL := fs.String("altcha-sentinel-url", Env("ATHENAEUM_ALTCHA_SENTINEL_URL", ""), "ALTCHA Sentinel base URL or challenge endpoint")
	altchaVerifyURL := fs.String("altcha-verify-url", Env("ATHENAEUM_ALTCHA_VERIFY_URL", ""), "Sentinel verify endpoint (defaults to {sentinel}/v1/verify/signature)")
	altchaAPIKeySecret := fs.String("altcha-api-key-secret", Env("ATHENAEUM_ALTCHA_API_KEY_SECRET", ""), "Sentinel API key secret for remote verification")
	altchaCost := fs.String("altcha-cost", Env("ATHENAEUM_ALTCHA_COST", "5000"), "builtin PoW cost / PBKDF2 iterations")
	altchaExpires := fs.String("altcha-expires", Env("ATHENAEUM_ALTCHA_EXPIRES", "300"), "builtin challenge expiry in seconds")
	altchaProtect := fs.String("altcha-protect", Env("ATHENAEUM_ALTCHA_PROTECT", "login,setup"), "comma-separated forms to protect: login, setup")
	altchaAuto := fs.String("altcha-widget-auto", Env("ATHENAEUM_ALTCHA_WIDGET_AUTO", "onsubmit"), "widget auto: off, onfocus, onload, onsubmit")
	altchaDisplay := fs.String("altcha-widget-display", Env("ATHENAEUM_ALTCHA_WIDGET_DISPLAY", "standard"), "widget display: standard, bar, floating, overlay, invisible")
	altchaType := fs.String("altcha-widget-type", Env("ATHENAEUM_ALTCHA_WIDGET_TYPE", "checkbox"), "widget type: checkbox, switch, native")
	altchaTheme := fs.String("altcha-widget-theme", Env("ATHENAEUM_ALTCHA_WIDGET_THEME", "auto"), "widget theme: auto, light, dark, or a named ALTCHA theme")
	altchaLang := fs.String("altcha-widget-language", Env("ATHENAEUM_ALTCHA_WIDGET_LANGUAGE", ""), "widget language ISO code (empty follows UI locale)")
	altchaName := fs.String("altcha-widget-name", Env("ATHENAEUM_ALTCHA_WIDGET_NAME", "altcha"), "hidden input name for the ALTCHA payload")
	altchaWorkers := fs.String("altcha-widget-workers", Env("ATHENAEUM_ALTCHA_WIDGET_WORKERS", "0"), "PoW web workers (0 = library default)")
	altchaHideLogo := fs.Bool("altcha-widget-hide-logo", EnvBool("ATHENAEUM_ALTCHA_WIDGET_HIDE_LOGO", false), "hide ALTCHA logo in the widget")
	altchaHideFooter := fs.Bool("altcha-widget-hide-footer", EnvBool("ATHENAEUM_ALTCHA_WIDGET_HIDE_FOOTER", false), "hide ALTCHA footer attribution")
	passwordMinLength := fs.String("password-min-length", Env("ATHENAEUM_PASSWORD_MIN_LENGTH", "8"), "minimum password length")
	passwordLongLength := fs.String("password-long-length", Env("ATHENAEUM_PASSWORD_LONG_LENGTH", "12"), "length that satisfies diversity without min-kinds (0 disables)")
	passwordMinKinds := fs.String("password-min-kinds", Env("ATHENAEUM_PASSWORD_MIN_KINDS", "3"), "minimum character classes (0 disables diversity rule)")
	passwordRequireLower := fs.Bool("password-require-lower", EnvBool("ATHENAEUM_PASSWORD_REQUIRE_LOWER", false), "require a lowercase letter")
	passwordRequireUpper := fs.Bool("password-require-upper", EnvBool("ATHENAEUM_PASSWORD_REQUIRE_UPPER", false), "require an uppercase letter")
	passwordRequireDigit := fs.Bool("password-require-digit", EnvBool("ATHENAEUM_PASSWORD_REQUIRE_DIGIT", false), "require a digit")
	passwordRequireSymbol := fs.Bool("password-require-symbol", EnvBool("ATHENAEUM_PASSWORD_REQUIRE_SYMBOL", false), "require a symbol")

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			PrintHelp(os.Stdout)
			os.Exit(0)
		}
		return Config{}, err
	}

	level := strings.TrimSpace(*logLevel)
	if *debug {
		level = "debug"
	}

	mode := strings.TrimSpace(*colorMode)
	if *noColor {
		mode = "never"
	}

	libAbs, err := filepath.Abs(*library)
	if err != nil {
		return Config{}, err
	}
	dataAbs, err := filepath.Abs(*data)
	if err != nil {
		return Config{}, err
	}
	webAbs := ""
	if strings.TrimSpace(*webDir) != "" {
		webAbs, err = filepath.Abs(strings.TrimSpace(*webDir))
		if err != nil {
			return Config{}, err
		}
	}

	return Config{
		Addr:                   *addr,
		LibraryDir:             libAbs,
		DataDir:                dataAbs,
		WebDir:                 webAbs,
		AdminUser:              *adminUser,
		AdminPass:              *adminPass,
		UploadMaxBytes:         envInt64(*uploadMaxStr, 2<<30),
		ScanWorkers:            int(envInt64(*scanWorkers, 2)),
		SentryDSN:              strings.TrimSpace(*sentryDSN),
		SentryDSNPublic:        strings.TrimSpace(*sentryDSNPublic),
		SentryEnvironment:      strings.TrimSpace(*sentryEnv),
		SentryRelease:          strings.TrimSpace(*sentryRelease),
		SentryTracesSampleRate: envFloat64(*sentryTraces, 0),
		LogLevel:               level,
		LogFile:                strings.TrimSpace(*logFile),
		ColorMode:              mode,
		NoColor:                *noColor,
		Sandbox:                strings.TrimSpace(*sandbox),
		SandboxLandlock:        *sandboxLandlock,
		SandboxSeccomp:         *sandboxSeccomp,
		Demo:                   *demo,
		PprofAddr:              strings.TrimSpace(*pprofAddr),
		DatabaseDriver:         strings.TrimSpace(*databaseDriver),
		DatabaseURL:            strings.TrimSpace(*databaseURL),
		AltchaEnabled:          *altchaEnabled,
		AltchaMode:             strings.TrimSpace(*altchaMode),
		AltchaHMACSecret:       strings.TrimSpace(*altchaHMAC),
		AltchaHMACKeySecret:    strings.TrimSpace(*altchaHMACKey),
		AltchaChallengeURL:     strings.TrimSpace(*altchaChallengeURL),
		AltchaSentinelURL:      strings.TrimSpace(*altchaSentinelURL),
		AltchaVerifyURL:        strings.TrimSpace(*altchaVerifyURL),
		AltchaAPIKeySecret:     strings.TrimSpace(*altchaAPIKeySecret),
		AltchaCost:             int(envInt64(*altchaCost, defaultAltchaCost)),
		AltchaExpiresSecs:      int(envInt64(*altchaExpires, defaultAltchaExpires)),
		AltchaProtect:          strings.TrimSpace(*altchaProtect),
		AltchaWidgetAuto:       strings.TrimSpace(*altchaAuto),
		AltchaWidgetDisplay:    strings.TrimSpace(*altchaDisplay),
		AltchaWidgetType:       strings.TrimSpace(*altchaType),
		AltchaWidgetTheme:      strings.TrimSpace(*altchaTheme),
		AltchaWidgetLanguage:   strings.TrimSpace(*altchaLang),
		AltchaWidgetName:       strings.TrimSpace(*altchaName),
		AltchaWidgetWorkers:    int(envInt64AllowZero(*altchaWorkers, 0)),
		AltchaWidgetHideLogo:   *altchaHideLogo,
		AltchaWidgetHideFooter: *altchaHideFooter,
		PasswordMinLength:      int(envInt64(*passwordMinLength, 8)),
		PasswordLongLength:     int(envInt64AllowZero(*passwordLongLength, 12)),
		PasswordMinKinds:       int(envInt64AllowZero(*passwordMinKinds, 3)),
		PasswordRequireLower:   *passwordRequireLower,
		PasswordRequireUpper:   *passwordRequireUpper,
		PasswordRequireDigit:   *passwordRequireDigit,
		PasswordRequireSymbol:  *passwordRequireSymbol,
	}, nil
}

const (
	defaultAltchaCost    = 5000
	defaultAltchaExpires = 300
)

// PrintHelp writes colored usage for the server binary.
func PrintHelp(w io.Writer) {
	name := strings.ToLower(brand.Name)
	term.Fprintln(w, term.Bold(w, brand.Name)+" - self-hosted EPUB, PDF, and audiobook library server")
	term.Fprintln(w)
	term.Fprintln(w, term.Header(w, "Usage:"))
	term.Fprintf(w, "  %s [flags]\n", name)
	term.Fprintf(w, "  %s users <command> [flags] [args]\n", name)
	term.Fprintln(w)
	term.Fprintln(w, term.Header(w, "Flags:"))
	printFlag(w, "--addr", "HTTP listen address (default :8080)")
	printFlag(w, "--library", "library root directory (default ./library)")
	printFlag(w, "--data", "database and cache directory (default ./data)")
	printFlag(w, "--web-dir", "serve frontend from directory (default: embedded assets)")
	printFlag(w, "--admin-user", "bootstrap admin username when no users exist")
	printFlag(w, "--admin-pass", "bootstrap admin password (requires --admin-user)")
	printFlag(w, "--upload-max-bytes", "maximum upload size in bytes (default 2GB)")
	printFlag(w, "--scan-workers", "parallel library index workers (default 2)")
	printFlag(w, "--log-level", "debug, info, warn, or error (default info)")
	printFlag(w, "--log-file", "also append logs to this file")
	printFlag(w, "--debug", "enable debug logging")
	printFlag(w, "--color", "auto, always, or never (default auto)")
	printFlag(w, "--no-color", "disable color (honors NO_COLOR)")
	printFlag(w, "--sandbox", "off, try, or strict Linux Landlock/seccomp (default off)")
	printFlag(w, "--sandbox-landlock", "toggle Landlock when sandbox is enabled (default true)")
	printFlag(w, "--sandbox-seccomp", "toggle seccomp-bpf when sandbox is enabled (default true)")
	printFlag(w, "--demo", "seed generated demo books, audiobooks, and covers")
	printFlag(w, "--pprof", "loopback pprof address (e.g. 127.0.0.1:6060)")
	printFlag(w, "--database-driver", "sqlite (default) or postgres")
	printFlag(w, "--database-url", "PostgreSQL URL when using postgres")
	printFlag(w, "--sentry-dsn", "Sentry or GlitchTip DSN")
	printFlag(w, "--altcha", "enable ALTCHA on login/setup (optional)")
	printFlag(w, "--altcha-mode", "builtin or sentinel (default builtin)")
	printFlag(w, "--password-min-length", "minimum password length (default 8)")
	printFlag(w, "--password-long-length", "diversity escape length (default 12, 0 off)")
	printFlag(w, "--password-min-kinds", "minimum character classes (default 3, 0 off)")
	printFlag(w, "--password-require-lower", "require lowercase letter")
	printFlag(w, "--password-require-upper", "require uppercase letter")
	printFlag(w, "--password-require-digit", "require digit")
	printFlag(w, "--password-require-symbol", "require symbol")
	printFlag(w, "--help", "show this help")
	term.Fprintln(w)
	term.Fprintln(w, term.Header(w, "Environment:"))
	term.Fprintln(w, "  Each flag has an ATHENAEUM_* equivalent. Flags override env. See .env.example.")
	term.Fprintln(w)
	term.Fprintln(w, term.Dim(w, "Also: NO_COLOR / FORCE_COLOR for terminal color conventions."))
	term.Fprintln(w)
	term.Fprintf(w, "Run %s for user management help.\n", term.Cyan(w, name+" users help"))
}

func printFlag(w io.Writer, name, desc string) {
	term.Fprintf(w, "  %-22s %s\n", term.Flag(w, name), desc)
}

func envInt64(v string, def int64) int64 {
	if v == "" {
		return def
	}
	var n int64
	if _, err := fmt.Sscan(v, &n); err == nil && n > 0 {
		return n
	}
	return def
}

func envInt64AllowZero(v string, def int64) int64 {
	if v == "" {
		return def
	}
	var n int64
	if _, err := fmt.Sscan(v, &n); err == nil && n >= 0 {
		return n
	}
	return def
}

func envFloat64(v string, def float64) float64 {
	if v == "" {
		return def
	}
	var f float64
	if _, err := fmt.Sscan(v, &f); err == nil && f >= 0 && f <= 1 {
		return f
	}
	return def
}
