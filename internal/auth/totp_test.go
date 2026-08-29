package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
)

func TestGenerateAndValidateTOTP(t *testing.T) {
	secret, url, err := GenerateTOTPSecret("alice")
	if err != nil {
		t.Fatal(err)
	}
	if secret == "" || !strings.Contains(url, "otpauth://") {
		t.Fatalf("secret=%q url=%q", secret, url)
	}
	code, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if !ValidateTOTPCode(secret, code) {
		t.Fatal("expected valid code")
	}
	if !ValidateTOTPCode(secret, "  "+code+"  ") {
		t.Fatal("expected trimmed code to validate")
	}
	if ValidateTOTPCode(secret, "000000") {
		t.Fatal("expected invalid code")
	}
	if ValidateTOTPCode("", code) || ValidateTOTPCode(secret, "") {
		t.Fatal("empty secret/code should fail")
	}
}
