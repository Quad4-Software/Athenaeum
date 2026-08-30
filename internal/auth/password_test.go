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
	if !CheckPassword(hash, "correct horse battery") {
		t.Fatal("expected password to match")
	}
	if CheckPassword(hash, "wrong") {
		t.Fatal("expected password mismatch")
	}
	// Under go test, hashes use MinCost so -race CI stays under the package timeout.
	if cost, err := bcrypt.Cost([]byte(hash)); err != nil || cost != bcrypt.MinCost {
		t.Fatalf("test hash cost=%d err=%v, want MinCost=%d", cost, err, bcrypt.MinCost)
	}
	if !strings.Contains(hash, "$2a$04$") && !strings.Contains(hash, "$2b$04$") {
		t.Fatalf("expected min-cost bcrypt prefix, got %q", hash)
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
