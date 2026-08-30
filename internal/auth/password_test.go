package auth

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("correct horse battery")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(hash, argon2idPrefix) {
		t.Fatalf("expected argon2id hash, got %q", hash)
	}
	if !CheckPassword(hash, "correct horse battery") {
		t.Fatal("expected password to match")
	}
	if CheckPassword(hash, "wrong") {
		t.Fatal("expected password mismatch")
	}
	if NeedsRehash(hash) {
		t.Fatal("fresh argon2id hash should not need rehash under test params")
	}
}

func TestCheckPasswordLegacyBcrypt(t *testing.T) {
	legacy, err := bcrypt.GenerateFromPassword([]byte("legacy-secret"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	if !CheckPassword(string(legacy), "legacy-secret") {
		t.Fatal("expected bcrypt hash to verify")
	}
	if CheckPassword(string(legacy), "nope") {
		t.Fatal("expected bcrypt mismatch")
	}
	if !NeedsRehash(string(legacy)) {
		t.Fatal("bcrypt hashes should need rehash to argon2id")
	}
}

func TestCheckPasswordRejectsGarbage(t *testing.T) {
	if CheckPassword("", "x") || CheckPassword("not-a-hash", "x") || CheckPassword("$argon2id$bad", "x") {
		t.Fatal("expected garbage hashes to fail")
	}
}

func TestScorePassword(t *testing.T) {
	SetPasswordPolicy(DefaultPasswordPolicy())
	t.Cleanup(func() { SetPasswordPolicy(DefaultPasswordPolicy()) })

	if st := ScorePassword("short"); st.Valid {
		t.Fatal("short password should be invalid")
	}
	if st := ScorePassword("longpassword"); !st.Valid {
		t.Fatalf("12+ char password should be valid: %+v", st)
	}
	if st := ScorePassword("Aa1!aaaa"); !st.Valid {
		t.Fatalf("mixed password should be valid: %+v", st)
	}
	if err := ValidatePassword("weak"); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestScorePasswordStrictPolicy(t *testing.T) {
	SetPasswordPolicy(PasswordPolicy{
		MinLength:     10,
		LongLength:    0,
		MinKinds:      0,
		RequireUpper:  true,
		RequireLower:  true,
		RequireDigit:  true,
		RequireSymbol: true,
	})
	t.Cleanup(func() { SetPasswordPolicy(DefaultPasswordPolicy()) })

	if st := ScorePassword("Aa1!aaaaaa"); !st.Valid {
		t.Fatalf("expected valid mixed password: %+v", st)
	}
	if st := ScorePassword("aaaaaaaaaa"); st.Valid {
		t.Fatalf("expected invalid without required classes: %+v", st)
	}
	found := false
	for _, r := range ScorePassword("aaaaaaaaaa").Requirements {
		if r.ID == "requireUpper" && !r.Met {
			found = true
		}
	}
	if !found {
		t.Fatal("expected unmet requireUpper requirement")
	}
}
