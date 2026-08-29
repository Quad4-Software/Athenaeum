package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"athenaeum/internal/models"
)

func setOIDCClaims(tok *oidc.IDToken, claims map[string]any) {
	b, err := json.Marshal(claims)
	if err != nil {
		panic(err)
	}
	v := reflect.ValueOf(tok).Elem().FieldByName("claims")
	*(*[]byte)(unsafe.Pointer(v.UnsafeAddr())) = b
}

func TestOIDCGroupsFromClaimsAndUserinfo(t *testing.T) {
	srv, _ := testServer(t)

	idTok := &oidc.IDToken{}
	setOIDCClaims(idTok, map[string]any{"groups": []string{"from-id-token"}})
	got := srv.oidcGroups(context.Background(), idTok, nil, nil, models.OIDCConfig{GroupClaim: "groups"})
	if len(got) != 1 || got[0] != "from-id-token" {
		t.Fatalf("id token groups=%v", got)
	}

	got = srv.oidcGroups(context.Background(), idTok, nil, nil, models.OIDCConfig{})
	if len(got) != 1 {
		t.Fatalf("default claim key groups=%v", got)
	}

	userinfo := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"roles": "ops,admins"})
	}))
	t.Cleanup(userinfo.Close)

	emptyTok := &oidc.IDToken{}
	setOIDCClaims(emptyTok, map[string]any{"sub": "u1"})
	oauthCfg := &oauth2.Config{}
	token := &oauth2.Token{AccessToken: "access", TokenType: "Bearer", Expiry: time.Now().Add(time.Hour)}
	got = srv.oidcGroups(context.Background(), emptyTok, token, oauthCfg, models.OIDCConfig{
		GroupClaim:  "roles",
		UserinfoURL: userinfo.URL,
	})
	if len(got) != 2 {
		t.Fatalf("userinfo groups=%v", got)
	}

	got = srv.oidcGroups(context.Background(), emptyTok, token, oauthCfg, models.OIDCConfig{
		GroupClaim: "roles",
	})
	if got != nil {
		t.Fatalf("no userinfo url got=%v", got)
	}
}
