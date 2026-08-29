// Package altcha provides optional ALTCHA proof-of-work challenge creation
// and payload verification for Athenaeum auth forms.
package altcha

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	altchalib "github.com/altcha-org/altcha-lib-go/v2"

	"athenaeum/internal/config"
)

const (
	defaultCost         = 5000
	defaultExpiresSecs  = 300
	defaultChallengeURL = "/api/altcha/challenge"
	hmacSecretFile      = "altcha_hmac_secret"     // #nosec G101 -- filename under data dir, not a credential
	hmacKeySecretFile   = "altcha_hmac_key_secret" // #nosec G101 -- filename under data dir, not a credential
)

var (
	// ErrDisabled means ALTCHA is not enabled.
	ErrDisabled = errors.New("altcha disabled")
	// ErrMissingPayload means the client omitted the altcha field.
	ErrMissingPayload = errors.New("altcha verification required")
	// ErrInvalidPayload means the payload failed cryptographic checks.
	ErrInvalidPayload = errors.New("altcha verification failed")
	// ErrExpired means the challenge expired.
	ErrExpired = errors.New("altcha challenge expired")
)

// Mode selects challenge generation and verification strategy.
type Mode string

const (
	ModeBuiltin  Mode = "builtin"
	ModeSentinel Mode = "sentinel"
)

// WidgetOptions are public customization knobs for the browser widget.
type WidgetOptions struct {
	Auto       string `json:"auto,omitempty"`
	Display    string `json:"display,omitempty"`
	HideFooter bool   `json:"hideFooter,omitempty"`
	HideLogo   bool   `json:"hideLogo,omitempty"`
	Language   string `json:"language,omitempty"`
	Name       string `json:"name,omitempty"`
	Theme      string `json:"theme,omitempty"`
	Type       string `json:"type,omitempty"`
	Workers    int    `json:"workers,omitempty"`
}

// PublicConfig is safe to expose to the browser.
type PublicConfig struct {
	Enabled         bool          `json:"enabled"`
	ChallengeURL    string        `json:"challengeUrl,omitempty"`
	ProtectLogin    bool          `json:"protectLogin"`
	ProtectSetup    bool          `json:"protectSetup"`
	ProtectRegister bool          `json:"protectRegister"`
	Widget          WidgetOptions `json:"widget"`
}

// Service creates challenges and verifies payloads.
type Service struct {
	cfg       config.Config
	hmac      string
	hmacKey   string
	deriveKey altchalib.DeriveKeyFunc
}

// New builds a Service from config. When enabled without an HMAC secret,
// secrets are loaded from or written into the data directory.
func New(cfg config.Config) (*Service, error) {
	if !cfg.AltchaEnabled {
		return &Service{cfg: cfg}, nil
	}

	hmacSecret := strings.TrimSpace(cfg.AltchaHMACSecret)
	hmacKeySecret := strings.TrimSpace(cfg.AltchaHMACKeySecret)

	if cfg.AltchaMode == string(ModeBuiltin) || cfg.AltchaMode == "" {
		var err error
		hmacSecret, err = ensureSecret(cfg.DataDir, hmacSecretFile, hmacSecret)
		if err != nil {
			return nil, err
		}
		hmacKeySecret, err = ensureSecret(cfg.DataDir, hmacKeySecretFile, hmacKeySecret)
		if err != nil {
			return nil, err
		}
	}

	return &Service{
		cfg:       cfg,
		hmac:      hmacSecret,
		hmacKey:   hmacKeySecret,
		deriveKey: altchalib.DeriveKeyPBKDF2(),
	}, nil
}

func ensureSecret(dataDir, name, provided string) (string, error) {
	if provided != "" {
		return provided, nil
	}
	path := filepath.Join(dataDir, name)
	if b, err := os.ReadFile(path); err == nil { // #nosec G304 -- fixed filename under configured data dir
		if s := strings.TrimSpace(string(b)); s != "" {
			return s, nil
		}
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("altcha secret: %w", err)
	}
	secret := hex.EncodeToString(raw)
	if err := os.MkdirAll(dataDir, 0o750); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(secret+"\n"), 0o600); err != nil {
		return "", err
	}
	return secret, nil
}

// Enabled reports whether ALTCHA is turned on.
func (s *Service) Enabled() bool {
	return s != nil && s.cfg.AltchaEnabled
}

// Mode returns the configured mode.
func (s *Service) Mode() Mode {
	if s == nil {
		return ModeBuiltin
	}
	switch strings.ToLower(strings.TrimSpace(s.cfg.AltchaMode)) {
	case string(ModeSentinel):
		return ModeSentinel
	default:
		return ModeBuiltin
	}
}

// Public returns browser-safe configuration.
func (s *Service) Public() PublicConfig {
	if s == nil || !s.Enabled() {
		return PublicConfig{}
	}
	challengeURL := strings.TrimSpace(s.cfg.AltchaChallengeURL)
	if challengeURL == "" {
		if s.Mode() == ModeSentinel {
			challengeURL = strings.TrimSpace(s.cfg.AltchaSentinelURL)
		} else {
			challengeURL = defaultChallengeURL
		}
	}
	return PublicConfig{
		Enabled:         true,
		ChallengeURL:    challengeURL,
		ProtectLogin:    s.protects("login"),
		ProtectSetup:    s.protects("setup"),
		ProtectRegister: s.protects("register-public"),
		Widget: WidgetOptions{
			Auto:       firstNonEmpty(s.cfg.AltchaWidgetAuto, "onsubmit"),
			Display:    firstNonEmpty(s.cfg.AltchaWidgetDisplay, "standard"),
			HideFooter: s.cfg.AltchaWidgetHideFooter,
			HideLogo:   s.cfg.AltchaWidgetHideLogo,
			Language:   s.cfg.AltchaWidgetLanguage,
			Name:       firstNonEmpty(s.cfg.AltchaWidgetName, "altcha"),
			Theme:      firstNonEmpty(s.cfg.AltchaWidgetTheme, "auto"),
			Type:       firstNonEmpty(s.cfg.AltchaWidgetType, "checkbox"),
			Workers:    s.cfg.AltchaWidgetWorkers,
		},
	}
}

func (s *Service) protects(action string) bool {
	raw := strings.TrimSpace(s.cfg.AltchaProtect)
	if raw == "" {
		raw = "login,setup"
	}
	for part := range strings.SplitSeq(raw, ",") {
		if strings.EqualFold(strings.TrimSpace(part), action) {
			return true
		}
	}
	return false
}

// CreateChallenge returns a fresh builtin PoW challenge.
func (s *Service) CreateChallenge() (altchalib.Challenge, error) {
	if !s.Enabled() || s.Mode() != ModeBuiltin {
		return altchalib.Challenge{}, ErrDisabled
	}
	cost := s.cfg.AltchaCost
	if cost <= 0 {
		cost = defaultCost
	}
	expiresSecs := s.cfg.AltchaExpiresSecs
	if expiresSecs <= 0 {
		expiresSecs = defaultExpiresSecs
	}
	expires := time.Now().Add(time.Duration(expiresSecs) * time.Second)

	n, err := rand.Int(rand.Reader, big.NewInt(5000))
	if err != nil {
		return altchalib.Challenge{}, err
	}
	counter := 5000 + int(n.Int64())

	return altchalib.CreateChallenge(altchalib.CreateChallengeOptions{
		Algorithm:              "PBKDF2/SHA-256",
		DeriveKey:              s.deriveKey,
		HMACSignatureSecret:    s.hmac,
		HMACKeySignatureSecret: s.hmacKey,
		Cost:                   cost,
		KeyLength:              32,
		Counter:                &counter,
		ExpiresAt:              &expires,
	})
}

// VerifyPayload checks a base64-encoded ALTCHA payload for the given action.
func (s *Service) VerifyPayload(ctx context.Context, action, payload string) error {
	if !s.Enabled() || !s.protects(action) {
		return nil
	}
	payload = strings.TrimSpace(payload)
	if payload == "" {
		return ErrMissingPayload
	}
	switch s.Mode() {
	case ModeSentinel:
		return s.verifySentinel(ctx, payload)
	default:
		return s.verifyBuiltin(payload)
	}
}

func (s *Service) verifyBuiltin(payload string) error {
	decoded, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return ErrInvalidPayload
	}
	var body altchalib.Payload
	if err := json.Unmarshal(decoded, &body); err != nil {
		return ErrInvalidPayload
	}
	result, err := altchalib.VerifySolution(altchalib.VerifySolutionOptions{
		Challenge:              body.Challenge,
		Solution:               body.Solution,
		DeriveKey:              s.deriveKey,
		HMACSignatureSecret:    s.hmac,
		HMACKeySignatureSecret: s.hmacKey,
	})
	if err != nil {
		return fmt.Errorf("altcha verify: %w", err)
	}
	if result.Expired {
		return ErrExpired
	}
	if !result.Verified {
		return ErrInvalidPayload
	}
	return nil
}

func (s *Service) verifySentinel(ctx context.Context, payload string) error {
	verifyURL := strings.TrimSpace(s.cfg.AltchaVerifyURL)
	if verifyURL == "" {
		base := strings.TrimRight(strings.TrimSpace(s.cfg.AltchaSentinelURL), "/")
		if base == "" {
			return fmt.Errorf("altcha sentinel verify URL not configured")
		}
		verifyURL = base + "/v1/verify/signature"
	}
	secret := strings.TrimSpace(s.cfg.AltchaAPIKeySecret)
	if secret == "" {
		secret = s.hmac
	}
	result, err := altchalib.VerifyServer(ctx, altchalib.VerifyServerOptions{
		URL:          verifyURL,
		Payload:      payload,
		Secret:       secret,
		Timeout:      5 * time.Second,
		Retries:      1,
		RetryBackoff: altchalib.RetryBackoffExponential,
	})
	if err != nil {
		return fmt.Errorf("altcha sentinel: %w", err)
	}
	if !result.Verified {
		return ErrInvalidPayload
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
