package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"

	"athenaeum/internal/auth"
	"athenaeum/internal/models"
)

func TestTOTPCoverageFullFlow(t *testing.T) {
	srv, store := testServer(t)
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	session, csrf := loginAdmin(t, handler, store)

	do := func(method, path string, body any, sess, cs *http.Cookie) *httptest.ResponseRecorder {
		t.Helper()
		var rdr *bytes.Reader
		if body != nil {
			b, err := json.Marshal(body)
			if err != nil {
				t.Fatal(err)
			}
			rdr = bytes.NewReader(b)
		} else {
			rdr = bytes.NewReader(nil)
		}
		req := httptest.NewRequest(method, path, rdr)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		if sess != nil {
			req.AddCookie(sess)
		}
		if method != http.MethodGet && method != http.MethodHead && cs != nil {
			withCSRF(req, cs)
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec
	}

	rec := do(http.MethodPost, "/api/auth/totp/enable", map[string]any{"code": "000000"}, session, csrf)
	if rec.Code != http.StatusConflict {
		t.Fatalf("enable before setup status=%d", rec.Code)
	}

	rec = do(http.MethodPost, "/api/auth/totp/setup", nil, session, csrf)
	if rec.Code != http.StatusOK {
		t.Fatalf("setup status=%d body=%s", rec.Code, rec.Body.String())
	}
	var setup totpSetupResponse
	if err := json.NewDecoder(rec.Body).Decode(&setup); err != nil {
		t.Fatal(err)
	}
	if setup.Secret == "" || setup.OtpauthURL == "" {
		t.Fatalf("setup=%+v", setup)
	}

	rec = do(http.MethodPost, "/api/auth/totp/enable", map[string]any{"code": "000000"}, session, csrf)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("enable bad code status=%d", rec.Code)
	}

	code, err := totp.GenerateCode(setup.Secret, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	rec = do(http.MethodPost, "/api/auth/totp/enable", map[string]any{"code": code}, session, csrf)
	if rec.Code != http.StatusOK {
		t.Fatalf("enable status=%d body=%s", rec.Code, rec.Body.String())
	}

	logoutReq := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	logoutReq.AddCookie(session)
	withCSRF(logoutReq, csrf)
	logoutRec := httptest.NewRecorder()
	handler.ServeHTTP(logoutRec, logoutReq)
	if logoutRec.Code != http.StatusNoContent && logoutRec.Code != http.StatusOK {
		t.Fatalf("logout status=%d", logoutRec.Code)
	}

	loginCSRF := fetchCSRF(t, handler)
	loginRec := do(http.MethodPost, "/api/auth/login", map[string]any{
		"username": "admin", "password": "longpassword",
	}, nil, loginCSRF)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("login challenge status=%d body=%s", loginRec.Code, loginRec.Body.String())
	}
	var challenge models.LoginChallenge
	if err := json.NewDecoder(loginRec.Body).Decode(&challenge); err != nil {
		t.Fatal(err)
	}
	if !challenge.NeedsTOTP || challenge.TOTPToken == "" {
		t.Fatalf("challenge=%+v", challenge)
	}

	verifyCSRF := fetchCSRF(t, handler)
	badRec := do(http.MethodPost, "/api/auth/totp/verify", map[string]any{
		"totpToken": challenge.TOTPToken, "code": "000000",
	}, nil, verifyCSRF)
	if badRec.Code != http.StatusUnauthorized {
		t.Fatalf("verify bad code status=%d", badRec.Code)
	}

	verifyCode, err := totp.GenerateCode(setup.Secret, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	verifyCSRF = fetchCSRF(t, handler)
	verifyRec := do(http.MethodPost, "/api/auth/totp/verify", map[string]any{
		"totpToken": challenge.TOTPToken, "code": verifyCode,
	}, nil, verifyCSRF)
	if verifyRec.Code != http.StatusOK {
		t.Fatalf("verify status=%d body=%s", verifyRec.Code, verifyRec.Body.String())
	}
	for _, c := range verifyRec.Result().Cookies() {
		switch c.Name {
		case auth.SessionCookie:
			session = c
		case auth.CSRFCookie:
			csrf = c
		}
	}
	if session == nil || csrf == nil {
		t.Fatal("missing cookies after verify")
	}

	bogusCSRF := fetchCSRF(t, handler)
	bogusRec := do(http.MethodPost, "/api/auth/totp/verify", map[string]any{
		"totpToken": "not-a-token", "code": "123456",
	}, nil, bogusCSRF)
	if bogusRec.Code != http.StatusUnauthorized {
		t.Fatalf("verify bad token status=%d", bogusRec.Code)
	}

	disableCode, err := totp.GenerateCode(setup.Secret, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	rec = do(http.MethodPost, "/api/auth/totp/disable", map[string]any{
		"password": "wrongpassword", "code": disableCode,
	}, session, csrf)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("disable bad password status=%d", rec.Code)
	}

	disableCode, err = totp.GenerateCode(setup.Secret, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	rec = do(http.MethodPost, "/api/auth/totp/disable", map[string]any{
		"password": "longpassword", "code": "000000",
	}, session, csrf)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("disable bad code status=%d", rec.Code)
	}

	disableCode, err = totp.GenerateCode(setup.Secret, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	rec = do(http.MethodPost, "/api/auth/totp/disable", map[string]any{
		"password": "longpassword", "code": disableCode,
	}, session, csrf)
	if rec.Code != http.StatusOK {
		t.Fatalf("disable status=%d body=%s", rec.Code, rec.Body.String())
	}

	_ = store
}
