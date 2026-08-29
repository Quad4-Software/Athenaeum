package auth

import (
	"errors"
	"strings"

	"github.com/pquerna/otp/totp"

	"athenaeum/internal/brand"
)

// ErrInvalidTOTPCode is returned when a submitted TOTP code fails validation.
var ErrInvalidTOTPCode = errors.New("invalid authenticator code")

// GenerateTOTPSecret creates a new random TOTP secret and its otpauth URL
// for the given account name (typically the username).
func GenerateTOTPSecret(accountName string) (secret string, otpauthURL string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      brand.Name,
		AccountName: accountName,
	})
	if err != nil {
		return "", "", err
	}
	return key.Secret(), key.URL(), nil
}

// ValidateTOTPCode reports whether code is a valid current TOTP for secret.
func ValidateTOTPCode(secret, code string) bool {
	code = strings.TrimSpace(code)
	if code == "" || secret == "" {
		return false
	}
	return totp.Validate(code, secret)
}
