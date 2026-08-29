package auth

import (
	"testing"
	"testing/quick"
)

func TestPropertyNormalizePasswordPolicyBounds(t *testing.T) {
	fn := func(minLen, longLen, minKinds int, up, lo, dig, sym bool) bool {
		p := NormalizePasswordPolicy(PasswordPolicy{
			MinLength:     minLen,
			LongLength:    longLen,
			MinKinds:      minKinds,
			RequireUpper:  up,
			RequireLower:  lo,
			RequireDigit:  dig,
			RequireSymbol: sym,
		})
		if p.MinLength < 4 || p.MinLength > 128 {
			return false
		}
		if p.LongLength < 0 || p.LongLength > 256 {
			return false
		}
		if p.LongLength > 0 && p.LongLength < p.MinLength {
			return false
		}
		if p.MinKinds < 0 || p.MinKinds > 4 {
			return false
		}
		return true
	}
	if err := quick.Check(fn, &quick.Config{MaxCount: 1000}); err != nil {
		t.Fatal(err)
	}
}

func TestPropertyScorePasswordDeterministic(t *testing.T) {
	SetPasswordPolicy(DefaultPasswordPolicy())
	t.Cleanup(func() { SetPasswordPolicy(DefaultPasswordPolicy()) })

	fn := func(pw string) bool {
		a := ScorePassword(pw)
		b := ScorePassword(pw)
		return a.Score == b.Score && a.Valid == b.Valid && a.Label == b.Label
	}
	if err := quick.Check(fn, &quick.Config{MaxCount: 400}); err != nil {
		t.Fatal(err)
	}
}

func TestPropertyValidatePasswordMatchesScore(t *testing.T) {
	SetPasswordPolicy(DefaultPasswordPolicy())
	t.Cleanup(func() { SetPasswordPolicy(DefaultPasswordPolicy()) })

	fn := func(pw string) bool {
		st := ScorePassword(pw)
		err := ValidatePassword(pw)
		if st.Valid {
			return err == nil
		}
		return err != nil
	}
	if err := quick.Check(fn, &quick.Config{MaxCount: 400}); err != nil {
		t.Fatal(err)
	}
}
