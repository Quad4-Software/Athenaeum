package models

import "time"

// Invite kinds.
const (
	InviteKindPermanent = "permanent"
	InviteKindGuest     = "guest"
)

// Invite is a tokenized account invitation.
type Invite struct {
	ID             int64      `json:"id"`
	Token          string     `json:"token"`
	Kind           string     `json:"kind"`
	Email          string     `json:"email,omitempty"`
	Permissions    int64      `json:"-"`
	CreatedBy      int64      `json:"createdBy"`
	ExpiresAt      *time.Time `json:"expiresAt,omitempty"`
	GuestExpiresAt *time.Time `json:"guestExpiresAt,omitempty"`
	PocketIDUserID string     `json:"pocketIdUserId,omitempty"`
	AcceptedAt     *time.Time `json:"acceptedAt,omitempty"`
	AcceptedUserID *int64     `json:"acceptedUserId,omitempty"`
	RevokedAt      *time.Time `json:"revokedAt,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
}

// InvitePublic is the admin-facing invite view with permission names.
type InvitePublic struct {
	ID             int64      `json:"id"`
	Token          string     `json:"token"`
	Kind           string     `json:"kind"`
	Email          string     `json:"email,omitempty"`
	Permissions    []string   `json:"permissions"`
	CreatedBy      int64      `json:"createdBy"`
	ExpiresAt      *time.Time `json:"expiresAt,omitempty"`
	GuestExpiresAt *time.Time `json:"guestExpiresAt,omitempty"`
	PocketIDUserID string     `json:"pocketIdUserId,omitempty"`
	AcceptedAt     *time.Time `json:"acceptedAt,omitempty"`
	AcceptedUserID *int64     `json:"acceptedUserId,omitempty"`
	RevokedAt      *time.Time `json:"revokedAt,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	Status         string     `json:"status"`
}

// Public returns the admin-facing invite representation.
func (inv Invite) Public() InvitePublic {
	status := "pending"
	now := time.Now()
	if inv.RevokedAt != nil {
		status = "revoked"
	} else if inv.AcceptedAt != nil {
		status = "accepted"
	} else if inv.ExpiresAt != nil && now.After(*inv.ExpiresAt) {
		status = "expired"
	}
	return InvitePublic{
		ID:             inv.ID,
		Token:          inv.Token,
		Kind:           inv.Kind,
		Email:          inv.Email,
		Permissions:    PermissionList(inv.Permissions),
		CreatedBy:      inv.CreatedBy,
		ExpiresAt:      inv.ExpiresAt,
		GuestExpiresAt: inv.GuestExpiresAt,
		PocketIDUserID: inv.PocketIDUserID,
		AcceptedAt:     inv.AcceptedAt,
		AcceptedUserID: inv.AcceptedUserID,
		RevokedAt:      inv.RevokedAt,
		CreatedAt:      inv.CreatedAt,
		Status:         status,
	}
}

// InviteMeta is public metadata for the accept page.
type InviteMeta struct {
	Kind               string     `json:"kind"`
	EmailPresent       bool       `json:"emailPresent"`
	ExpiresAt          *time.Time `json:"expiresAt,omitempty"`
	Valid              bool       `json:"valid"`
	Reason             string     `json:"reason,omitempty"`
	PocketIDConfigured bool       `json:"pocketIdConfigured"`
}

// InviteCreateResult is returned when an admin creates an invite.
type InviteCreateResult struct {
	Invite           InvitePublic `json:"invite"`
	URL              string       `json:"url"`
	PocketIDSetupURL string       `json:"pocketIdSetupUrl,omitempty"`
	EmailSent        bool         `json:"emailSent"`
}
