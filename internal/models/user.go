package models

import "time"

// User is an authenticated library account.
type User struct {
	ID          int64      `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email,omitempty"`
	IsAdmin     bool       `json:"isAdmin"`
	IsGuest     bool       `json:"isGuest"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
	Permissions int64      `json:"-"`
	LocalAuth   bool       `json:"localAuth"`
	TOTPEnabled bool       `json:"-"`
	CreatedAt   time.Time  `json:"createdAt"`
}

// PermissionNames is the list of granted permissions for API responses.
func (u User) PermissionNames() []string {
	return PermissionList(EffectivePermissions(u))
}

// Public is the user record exposed to clients.
type UserPublic struct {
	ID          int64      `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email,omitempty"`
	IsAdmin     bool       `json:"isAdmin"`
	IsGuest     bool       `json:"isGuest,omitempty"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
	LocalAuth   bool       `json:"localAuth,omitempty"`
	TOTPEnabled bool       `json:"totpEnabled"`
	Permissions []string   `json:"permissions"`
	CreatedAt   time.Time  `json:"createdAt"`
}

func (u User) Public() UserPublic {
	return UserPublic{
		ID:          u.ID,
		Username:    u.Username,
		Email:       u.Email,
		IsAdmin:     u.IsAdmin,
		IsGuest:     u.IsGuest,
		ExpiresAt:   u.ExpiresAt,
		LocalAuth:   u.LocalAuth,
		TOTPEnabled: u.TOTPEnabled,
		Permissions: u.PermissionNames(),
		CreatedAt:   u.CreatedAt,
	}
}

// LoginChallenge is returned instead of a session when a password check
// succeeds but a second authentication factor is still required.
type LoginChallenge struct {
	NeedsTOTP bool   `json:"needsTotp"`
	TOTPToken string `json:"totpToken"`
}

// AuthSettings holds server-wide authentication policy toggles.
type AuthSettings struct {
	AllowRegistration bool `json:"allowRegistration"`
	RequireTOTP       bool `json:"requireTotp"`
}

// GuestCredentials is returned once when an admin creates a temporary account.
type GuestCredentials struct {
	User     UserPublic `json:"user"`
	Password string     `json:"password"`
}

// Collection kind constants.
const (
	CollectionManual  = "manual"
	CollectionSmart   = "smart"
	CollectionAuto    = "auto"
	CollectionReading = "reading"
)

// SmartQuery defines dynamic membership rules for smart/auto collections.
type SmartQuery struct {
	Format    string `json:"format,omitempty"`
	Author    string `json:"author,omitempty"`
	Series    string `json:"series,omitempty"`
	Search    string `json:"search,omitempty"`
	AddedDays int    `json:"addedDays,omitempty"`
}

// Collection is a manual shelf or a smart/auto collection of books.
type Collection struct {
	ID          int64       `json:"id"`
	UserID      int64       `json:"userId,omitempty"`
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Kind        string      `json:"kind"`
	Query       *SmartQuery `json:"query,omitempty"`
	BookCount   int64       `json:"bookCount"`
	CreatedAt   time.Time   `json:"createdAt"`
}
