package server

import (
	"testing"
)

func TestVerifyWebhookSignature(t *testing.T) {
	payload := []byte(`{"event":"ping"}`)
	sig := signWebhookPayload("secret", payload)
	header := "sha256=" + sig
	if !VerifyWebhookSignature("secret", header, payload) {
		t.Fatal("expected valid signature")
	}
	if VerifyWebhookSignature("wrong", header, payload) {
		t.Fatal("expected invalid signature")
	}
}
