// Package auth provides password hashing and session token helpers.
package auth

import (
	"errors"
	"fmt"
	"sync"
)

// PasswordPolicy is the configurable password strength policy.
type PasswordPolicy struct {
	MinLength     int  `json:"minLength"`
	LongLength    int  `json:"longLength"`
	MinKinds      int  `json:"minKinds"`
	RequireLower  bool `json:"requireLower"`
	RequireUpper  bool `json:"requireUpper"`
	RequireDigit  bool `json:"requireDigit"`
	RequireSymbol bool `json:"requireSymbol"`
}

// PasswordRequirement is one checklist item for UI feedback.
type PasswordRequirement struct {
	ID  string `json:"id"`
	Met bool   `json:"met"`
}

// PasswordStrength describes how strong a password is (0–4).
type PasswordStrength struct {
	Score        int                   `json:"score"`
	Label        string                `json:"label"`
	Valid        bool                  `json:"valid"`
	Issues       []string              `json:"issues,omitempty"`
	Requirements []PasswordRequirement `json:"requirements,omitempty"`
}

var (
	policyMu sync.RWMutex
	policy   = DefaultPasswordPolicy()
)

// DefaultPasswordPolicy matches the historical built-in rules.
func DefaultPasswordPolicy() PasswordPolicy {
	return PasswordPolicy{
		MinLength:  8,
		LongLength: 12,
		MinKinds:   3,
	}
}

// NormalizePasswordPolicy clamps policy fields to safe ranges.
func NormalizePasswordPolicy(p PasswordPolicy) PasswordPolicy {
	if p.MinLength < 4 {
		p.MinLength = 4
	}
	if p.MinLength > 128 {
		p.MinLength = 128
	}
	if p.LongLength < 0 {
		p.LongLength = 0
	}
	if p.LongLength > 256 {
		p.LongLength = 256
	}
	if p.LongLength > 0 && p.LongLength < p.MinLength {
		p.LongLength = p.MinLength
	}
	if p.MinKinds < 0 {
		p.MinKinds = 0
	}
	if p.MinKinds > 4 {
		p.MinKinds = 4
	}
	return p
}

// SetPasswordPolicy installs the runtime password policy.
func SetPasswordPolicy(p PasswordPolicy) {
	policyMu.Lock()
	defer policyMu.Unlock()
	policy = NormalizePasswordPolicy(p)
}

// GetPasswordPolicy returns the active password policy.
func GetPasswordPolicy() PasswordPolicy {
	policyMu.RLock()
	defer policyMu.RUnlock()
	return policy
}

// ValidatePassword reports whether password meets minimum policy.
func ValidatePassword(password string) error {
	st := ScorePassword(password)
	if st.Valid {
		return nil
	}
	if len(st.Issues) > 0 {
		return errors.New(st.Issues[0])
	}
	return errors.New("password is too weak")
}

// ScorePassword rates password strength for UI feedback using the active policy.
func ScorePassword(password string) PasswordStrength {
	return ScorePasswordWithPolicy(password, GetPasswordPolicy())
}

// ScorePasswordWithPolicy rates password strength against an explicit policy.
func ScorePasswordWithPolicy(password string, p PasswordPolicy) PasswordStrength {
	p = NormalizePasswordPolicy(p)
	var st PasswordStrength

	var hasLower, hasUpper, hasDigit, hasSymbol bool
	for _, r := range password {
		switch {
		case r >= 'a' && r <= 'z':
			hasLower = true
		case r >= 'A' && r <= 'Z':
			hasUpper = true
		case r >= '0' && r <= '9':
			hasDigit = true
		default:
			hasSymbol = true
		}
	}
	kinds := 0
	if hasLower {
		kinds++
	}
	if hasUpper {
		kinds++
	}
	if hasDigit {
		kinds++
	}
	if hasSymbol {
		kinds++
	}

	lengthOK := len(password) >= p.MinLength
	diversityOK := true
	if p.MinKinds > 0 {
		longOK := p.LongLength > 0 && len(password) >= p.LongLength
		diversityOK = longOK || kinds >= p.MinKinds
	}

	reqs := []PasswordRequirement{{
		ID:  "minLength",
		Met: lengthOK,
	}}
	if p.RequireLower {
		reqs = append(reqs, PasswordRequirement{ID: "requireLower", Met: hasLower})
	}
	if p.RequireUpper {
		reqs = append(reqs, PasswordRequirement{ID: "requireUpper", Met: hasUpper})
	}
	if p.RequireDigit {
		reqs = append(reqs, PasswordRequirement{ID: "requireDigit", Met: hasDigit})
	}
	if p.RequireSymbol {
		reqs = append(reqs, PasswordRequirement{ID: "requireSymbol", Met: hasSymbol})
	}
	if p.MinKinds > 0 {
		reqs = append(reqs, PasswordRequirement{ID: "diversity", Met: diversityOK})
	}
	st.Requirements = reqs

	if !lengthOK {
		st.Issues = append(st.Issues, fmt.Sprintf("at least %d characters", p.MinLength))
	}
	if p.RequireLower && !hasLower {
		st.Issues = append(st.Issues, "include a lowercase letter")
	}
	if p.RequireUpper && !hasUpper {
		st.Issues = append(st.Issues, "include an uppercase letter")
	}
	if p.RequireDigit && !hasDigit {
		st.Issues = append(st.Issues, "include a digit")
	}
	if p.RequireSymbol && !hasSymbol {
		st.Issues = append(st.Issues, "include a symbol")
	}
	if p.MinKinds > 0 && !diversityOK {
		if p.LongLength > 0 {
			st.Issues = append(st.Issues, fmt.Sprintf("use %d+ characters or mix upper, lower, digits, and symbols", p.LongLength))
		} else {
			st.Issues = append(st.Issues, fmt.Sprintf("use at least %d character types", p.MinKinds))
		}
	}

	score := 0
	if lengthOK {
		score++
	}
	if p.LongLength > 0 && len(password) >= p.LongLength {
		score++
	} else if p.LongLength == 0 && lengthOK && len(password) >= p.MinLength+4 {
		score++
	}
	if kinds >= 3 {
		score++
	}
	if kinds >= 4 && len(password) >= max(10, p.MinLength) {
		score++
	}
	st.Score = score
	switch {
	case score <= 1:
		st.Label = "weak"
	case score == 2:
		st.Label = "fair"
	case score == 3:
		st.Label = "good"
	default:
		st.Label = "strong"
	}

	st.Valid = lengthOK && diversityOK &&
		(!p.RequireLower || hasLower) &&
		(!p.RequireUpper || hasUpper) &&
		(!p.RequireDigit || hasDigit) &&
		(!p.RequireSymbol || hasSymbol)
	return st
}
