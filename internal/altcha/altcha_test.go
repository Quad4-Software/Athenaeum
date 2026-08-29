package altcha_test

import (
	"encoding/base64"
	"encoding/json"
	"path/filepath"
	"testing"

	altchalib "github.com/altcha-org/altcha-lib-go/v2"

	"athenaeum/internal/altcha"
	"athenaeum/internal/config"
)

func TestBuiltinChallengeRoundTrip(t *testing.T) {
	dir := t.TempDir()
	svc, err := altcha.New(config.Config{
		DataDir:             dir,
		AltchaEnabled:       true,
		AltchaMode:          "builtin",
		AltchaHMACSecret:    "test-hmac-secret",
		AltchaHMACKeySecret: "test-hmac-key-secret",
		AltchaCost:          100,
		AltchaExpiresSecs:   60,
		AltchaProtect:       "login,setup",
		AltchaWidgetAuto:    "onsubmit",
		AltchaWidgetDisplay: "standard",
		AltchaWidgetType:    "checkbox",
		AltchaWidgetTheme:   "auto",
		AltchaWidgetName:    "altcha",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !svc.Enabled() {
		t.Fatal("expected enabled")
	}
	pub := svc.Public()
	if !pub.Enabled || pub.ChallengeURL != "/api/altcha/challenge" {
		t.Fatalf("public=%+v", pub)
	}
	if !pub.ProtectLogin || !pub.ProtectSetup {
		t.Fatalf("protect flags %+v", pub)
	}

	challenge, err := svc.CreateChallenge()
	if err != nil {
		t.Fatal(err)
	}
	solution, err := altchalib.SolveChallenge(altchalib.SolveChallengeOptions{
		Challenge: challenge,
		DeriveKey: altchalib.DeriveKeyPBKDF2(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if solution == nil {
		t.Fatal("nil solution")
	}
	raw, err := json.Marshal(altchalib.Payload{Challenge: challenge, Solution: *solution})
	if err != nil {
		t.Fatal(err)
	}
	payload := base64.StdEncoding.EncodeToString(raw)
	if err := svc.VerifyPayload(t.Context(), "login", payload); err != nil {
		t.Fatalf("verify: %v", err)
	}
	if err := svc.VerifyPayload(t.Context(), "login", ""); err != altcha.ErrMissingPayload {
		t.Fatalf("empty payload err=%v", err)
	}
}

func TestDisabledSkipsVerification(t *testing.T) {
	svc, err := altcha.New(config.Config{DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.VerifyPayload(t.Context(), "login", ""); err != nil {
		t.Fatal(err)
	}
}

func TestPersistsHMACSecret(t *testing.T) {
	dir := t.TempDir()
	svc, err := altcha.New(config.Config{
		DataDir:       dir,
		AltchaEnabled: true,
		AltchaMode:    "builtin",
		AltchaCost:    100,
	})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "altcha_hmac_secret")
	if _, err := filepath.Glob(path); err != nil {
		t.Fatal(err)
	}
	challenge, err := svc.CreateChallenge()
	if err != nil {
		t.Fatal(err)
	}
	if challenge.Signature == "" {
		t.Fatal("expected signature")
	}
	svc2, err := altcha.New(config.Config{
		DataDir:       dir,
		AltchaEnabled: true,
		AltchaMode:    "builtin",
		AltchaCost:    100,
	})
	if err != nil {
		t.Fatal(err)
	}
	solution, err := altchalib.SolveChallenge(altchalib.SolveChallengeOptions{
		Challenge: challenge,
		DeriveKey: altchalib.DeriveKeyPBKDF2(),
	})
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(altchalib.Payload{Challenge: challenge, Solution: *solution})
	payload := base64.StdEncoding.EncodeToString(raw)
	if err := svc2.VerifyPayload(t.Context(), "login", payload); err != nil {
		t.Fatalf("verify with persisted secret: %v", err)
	}
}
