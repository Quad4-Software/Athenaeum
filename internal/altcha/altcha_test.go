package altcha_test

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestVerifySentinelHTTPtest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/verify/signature" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"verified": true})
	}))
	defer srv.Close()

	svc, err := altcha.New(config.Config{
		DataDir:            t.TempDir(),
		AltchaEnabled:      true,
		AltchaMode:         "sentinel",
		AltchaSentinelURL:  srv.URL,
		AltchaAPIKeySecret: "sentinel-secret",
		AltchaProtect:      "login",
	})
	if err != nil {
		t.Fatal(err)
	}
	if svc.Mode() != altcha.ModeSentinel {
		t.Fatalf("mode=%v", svc.Mode())
	}
	if err := svc.VerifyPayload(t.Context(), "login", "payload-token"); err != nil {
		t.Fatalf("verify: %v", err)
	}

	fail := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"verified": false, "reason": "bad"})
	}))
	defer fail.Close()
	svcFail, err := altcha.New(config.Config{
		DataDir:          t.TempDir(),
		AltchaEnabled:    true,
		AltchaMode:       "sentinel",
		AltchaVerifyURL:  fail.URL,
		AltchaProtect:    "login",
		AltchaHMACSecret: "fallback-secret",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := svcFail.VerifyPayload(t.Context(), "login", "payload-token"); err != altcha.ErrInvalidPayload {
		t.Fatalf("err=%v", err)
	}

	svcMissing, err := altcha.New(config.Config{
		DataDir:       t.TempDir(),
		AltchaEnabled: true,
		AltchaMode:    "sentinel",
		AltchaProtect: "login",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := svcMissing.VerifyPayload(t.Context(), "login", "x"); err == nil {
		t.Fatal("expected missing verify URL error")
	}
}
