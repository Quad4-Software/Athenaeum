package storage

import (
	"context"
	"testing"
	"time"

	"athenaeum/internal/models"
)

func TestInviteCreateAcceptRevoke(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	adminID, err := s.CreateUser(ctx, "admin", "hash", true)
	if err != nil {
		t.Fatal(err)
	}
	expires := time.Now().Add(24 * time.Hour)
	inv, err := s.CreateInvite(ctx, models.Invite{
		Kind:        models.InviteKindPermanent,
		Email:       "a@example.com",
		Permissions: models.DefaultUserPermissions,
		CreatedBy:   adminID,
		ExpiresAt:   &expires,
	})
	if err != nil {
		t.Fatal(err)
	}
	if inv.Token == "" {
		t.Fatal("expected token")
	}
	got, err := s.GetInviteByToken(ctx, inv.Token)
	if err != nil {
		t.Fatal(err)
	}
	if got.Email != "a@example.com" {
		t.Fatalf("email = %q", got.Email)
	}
	uid, err := s.CreateInvitedUser(ctx, "alice", "hash", "a@example.com", models.DefaultUserPermissions)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.AcceptInvite(ctx, inv.ID, uid); err != nil {
		t.Fatal(err)
	}
	if err := s.AcceptInvite(ctx, inv.ID, uid); err != ErrConflict {
		t.Fatalf("second accept = %v", err)
	}

	inv2, err := s.CreateInvite(ctx, models.Invite{
		Kind:      models.InviteKindGuest,
		CreatedBy: adminID,
		ExpiresAt: &expires,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.RevokeInvite(ctx, inv2.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.RevokeInvite(ctx, inv2.ID); err != ErrNotFound {
		t.Fatalf("second revoke = %v", err)
	}
}

func TestWebhookCRUDAndDeliveries(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	wh, err := s.CreateWebhook(ctx, models.Webhook{
		URL:     "http://example.com/hook",
		Secret:  "sekrit",
		Events:  []string{models.WebhookEventUserCreate},
		Enabled: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	list, err := s.ListEnabledWebhooksForEvent(ctx, models.WebhookEventUserCreate)
	if err != nil || len(list) != 1 {
		t.Fatalf("list = %v err=%v", list, err)
	}
	id, err := s.InsertWebhookDelivery(ctx, models.WebhookDelivery{
		WebhookID:  wh.ID,
		Event:      models.WebhookEventUserCreate,
		Payload:    `{}`,
		StatusCode: 200,
		Success:    true,
		Attempts:   1,
		CreatedAt:  time.Now(),
	})
	if err != nil || id == 0 {
		t.Fatalf("delivery id=%d err=%v", id, err)
	}
	dels, err := s.ListWebhookDeliveries(ctx, wh.ID, 10, 0)
	if err != nil || len(dels) != 1 {
		t.Fatalf("deliveries = %v err=%v", dels, err)
	}
	if err := s.DeleteWebhook(ctx, wh.ID); err != nil {
		t.Fatal(err)
	}
}

func TestPocketIDSettings(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	cfg, err := s.GetPocketIDSettings(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Enabled {
		t.Fatal("expected disabled by default")
	}
	cfg.Enabled = true
	cfg.BaseURL = "https://id.example.com"
	cfg.APIKey = "key"
	cfg.DefaultGroupIDs = []string{"g1"}
	if err := s.SavePocketIDSettings(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetPocketIDSettings(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Enabled || got.BaseURL != "https://id.example.com" || got.APIKey != "key" {
		t.Fatalf("%+v", got)
	}
	if len(got.DefaultGroupIDs) != 1 || got.DefaultGroupIDs[0] != "g1" {
		t.Fatalf("groups=%v", got.DefaultGroupIDs)
	}
}
