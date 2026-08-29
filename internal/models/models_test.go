package models

import (
	"testing"
	"time"
)

func TestIsAudio(t *testing.T) {
	for _, f := range []string{FormatMP3, FormatM4B, FormatM4A, FormatOGG, FormatFLAC, FormatAudiobook} {
		if !IsAudio(f) {
			t.Errorf("IsAudio(%q) = false", f)
		}
	}
	if IsAudio(FormatEPUB) {
		t.Fatal("epub should not be audio")
	}
}

func TestIsComic(t *testing.T) {
	if !IsComic(FormatCBZ) || !IsComic(FormatCBR) {
		t.Fatal("expected comic formats")
	}
	if IsComic(FormatPDF) {
		t.Fatal("pdf should not be comic")
	}
}

func TestIsMobiFamily(t *testing.T) {
	for _, f := range []string{FormatMOBI, FormatAZW3, FormatAZW} {
		if !IsMobiFamily(f) {
			t.Errorf("IsMobiFamily(%q) = false", f)
		}
	}
	if IsMobiFamily(FormatKFX) || IsMobiFamily(FormatEPUB) {
		t.Fatal("kfx/epub should not be mobi family")
	}
}

func TestParsePermissions(t *testing.T) {
	mask := ParsePermissions([]string{"read", "edit_metadata", "unknown"})
	if !HasPermission(mask, PermRead) || !HasPermission(mask, PermEditMetadata) {
		t.Fatalf("mask=%d missing expected bits", mask)
	}
	if HasPermission(mask, PermManageUsers) {
		t.Fatal("unexpected manage_users")
	}
	if ParsePermissions(nil) != 0 {
		t.Fatal("empty should be 0")
	}
	names := PermissionList(PermRead | PermManageLibrary)
	if len(names) != 2 || names[0] != "read" || names[1] != "manage_library" {
		t.Fatalf("PermissionList = %v", names)
	}
}

func TestEffectivePermissions(t *testing.T) {
	if EffectivePermissions(User{IsAdmin: true}) != AllPermissions {
		t.Fatal("admin should have all")
	}
	if EffectivePermissions(User{Permissions: PermRead}) != PermRead {
		t.Fatal("explicit mask should win")
	}
	if EffectivePermissions(User{}) != DefaultUserPermissions {
		t.Fatal("zero should use defaults")
	}
}

func TestInvitePublicStatus(t *testing.T) {
	now := time.Now()
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)

	if got := (Invite{Token: "a"}).Public(); got.Status != "pending" {
		t.Fatalf("pending status=%q", got.Status)
	}
	if got := (Invite{RevokedAt: &past}).Public(); got.Status != "revoked" {
		t.Fatalf("revoked status=%q", got.Status)
	}
	if got := (Invite{AcceptedAt: &past}).Public(); got.Status != "accepted" {
		t.Fatalf("accepted status=%q", got.Status)
	}
	if got := (Invite{ExpiresAt: &past}).Public(); got.Status != "expired" {
		t.Fatalf("expired status=%q", got.Status)
	}
	if got := (Invite{ExpiresAt: &future, Permissions: PermRead}).Public(); got.Status != "pending" {
		t.Fatalf("future expiry status=%q", got.Status)
	}
	if len((Invite{Permissions: PermRead}).Public().Permissions) != 1 {
		t.Fatal("expected permission names")
	}
}

func TestPocketIDSettingsPublic(t *testing.T) {
	got := PocketIDSettings{Enabled: true, BaseURL: "https://id.example", APIKey: "secret"}.Public()
	if !got.Enabled || got.BaseURL != "https://id.example" || !got.APIKeySet {
		t.Fatalf("got %+v", got)
	}
	if got.DefaultGroupIDs == nil || len(got.DefaultGroupIDs) != 0 {
		t.Fatalf("nil groups should become empty slice: %#v", got.DefaultGroupIDs)
	}
	got = PocketIDSettings{DefaultGroupIDs: []string{"g1"}}.Public()
	if len(got.DefaultGroupIDs) != 1 || got.APIKeySet {
		t.Fatalf("got %+v", got)
	}
}

func TestSMTPSettingsPublic(t *testing.T) {
	got := SMTPSettings{Host: "mail", Password: "x", Port: 587}.Public()
	if !got.PasswordSet || got.Host != "mail" || got.Port != 587 {
		t.Fatalf("got %+v", got)
	}
	if (SMTPSettings{}).Public().PasswordSet {
		t.Fatal("empty password should not be set")
	}
}

func TestUserPublic(t *testing.T) {
	u := User{ID: 1, Username: "bob", IsAdmin: true, LocalAuth: true}
	p := u.Public()
	if p.Username != "bob" || !p.IsAdmin || len(p.Permissions) == 0 {
		t.Fatalf("got %+v", p)
	}
}
