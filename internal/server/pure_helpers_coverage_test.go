package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"athenaeum/internal/auth"
	"athenaeum/internal/models"
)

func TestPureHelpersCoverage(t *testing.T) {
	if got := emailLocalPart("user@example.com"); got != "user" {
		t.Fatalf("emailLocalPart=%q", got)
	}
	if got := emailLocalPart("nopesign"); got != "nopesign" {
		t.Fatalf("emailLocalPart no at=%q", got)
	}

	if parentRel("a/b/c.pdf") != "a/b" {
		t.Fatal("parentRel nested")
	}
	if parentRel("file.pdf") != "" {
		t.Fatal("parentRel root")
	}

	if stringsTrim("  hi\t") != "hi" {
		t.Fatal("stringsTrim")
	}

	if sanitizeUsername("Alice Bob@x") != "Alice_Bob_x" {
		t.Fatalf("sanitizeUsername=%q", sanitizeUsername("Alice Bob@x"))
	}
	if sanitizeUsername("!") != "" {
		t.Fatal("sanitizeUsername short")
	}
	if sanitizeUsername("") != "" {
		t.Fatal("sanitizeUsername empty")
	}
	longName := strings.Repeat("a", 80)
	if got := sanitizeUsername(longName); len(got) != 64 {
		t.Fatalf("sanitizeUsername truncate len=%d", len(got))
	}
	if sanitizeUsername("..ab..") != "ab" {
		t.Fatalf("sanitizeUsername trim edges=%q", sanitizeUsername("..ab.."))
	}

	groups := groupsFromClaim(map[string]any{"groups": []any{"a", "b", 3}}, "groups")
	if len(groups) != 2 {
		t.Fatalf("groups slice=%v", groups)
	}
	groups = groupsFromClaim(map[string]any{"groups": "x, y"}, "groups")
	if len(groups) != 2 {
		t.Fatalf("groups csv=%v", groups)
	}
	groups = groupsFromClaim(map[string]any{"groups": []string{"z"}}, "groups")
	if len(groups) != 1 || groups[0] != "z" {
		t.Fatalf("groups []string=%v", groups)
	}
	if groupsFromClaim(map[string]any{}, "groups") != nil {
		t.Fatal("missing claim")
	}
	if groupsFromClaim(map[string]any{"groups": 42}, "groups") != nil {
		t.Fatal("unsupported claim type")
	}

	if !oidcGroupMatchesAdmin([]string{"Readers", "Admins"}, "admins,ops") {
		t.Fatal("oidcGroupMatchesAdmin")
	}
	if oidcGroupMatchesAdmin([]string{"Readers"}, "") {
		t.Fatal("empty admin groups")
	}
	if oidcGroupMatchesAdmin(nil, "admins") {
		t.Fatal("empty groups")
	}
	if oidcGroupMatchesAdmin([]string{"Readers"}, " , ") {
		t.Fatal("blank admin list parts")
	}

	doc := BuildOpenAPI()
	if doc.OpenAPI == "" || len(doc.Paths) == 0 {
		t.Fatal("BuildOpenAPI empty")
	}

	if absoluteOrigin("https://ex.test/path") != "https://ex.test" {
		t.Fatalf("absoluteOrigin=%q", absoluteOrigin("https://ex.test/path"))
	}
	if absoluteOrigin("/relative") != "" {
		t.Fatal("absoluteOrigin relative")
	}

	if !corsOriginAllowed("https://a.test", "*") {
		t.Fatal("cors *")
	}
	if !corsOriginAllowed("https://a.test", "https://b.test, https://a.test") {
		t.Fatal("cors list")
	}
	if corsOriginAllowed("https://nope", "https://a.test") {
		t.Fatal("cors deny")
	}

	policy := "default-src 'self'; connect-src 'self';"
	got := withAltchaConnectSrc(policy, "https://chal.example/challenge")
	if !strings.Contains(got, "https://chal.example") {
		t.Fatalf("withAltchaConnectSrc=%q", got)
	}
	if withAltchaConnectSrc(policy, "/local") != policy {
		t.Fatal("withAltchaConnectSrc local")
	}

	long := make([]byte, 250)
	for i := range long {
		long[i] = 'a'
	}
	if trunc := truncateErr(long); len(trunc) <= 200 || !strings.HasSuffix(trunc, "…") {
		t.Fatalf("truncateErr=%q", trunc)
	}
	if truncateErr([]byte(" short ")) != "short" {
		t.Fatal("truncateErr short")
	}

	name, err := uploadPartName("0123456789abcdef0123456789abcdef")
	if err != nil || name != "0123456789abcdef0123456789abcdef.part" {
		t.Fatalf("uploadPartName=%q err=%v", name, err)
	}
	if _, err := uploadPartName("bad"); err == nil {
		t.Fatal("uploadPartName invalid")
	}

	start, end, total, err := parseContentRange("bytes 0-9/100")
	if err != nil || start != 0 || end != 9 || total != 100 {
		t.Fatalf("parseContentRange %d-%d/%d err=%v", start, end, total, err)
	}
	if _, _, _, err := parseContentRange("nope"); err == nil {
		t.Fatal("parseContentRange bad")
	}

	if _, err := sanitizeUploadRelPath("/books/../x.pdf"); err == nil {
		t.Fatal("sanitizeUploadRelPath traversal")
	}
	rel, err := sanitizeUploadRelPath("books/a.pdf")
	if err != nil || rel != "books/a.pdf" {
		t.Fatalf("sanitizeUploadRelPath=%q err=%v", rel, err)
	}

	if uploadFormatFromExt("x.epub") == "" {
		t.Fatal("uploadFormatFromExt")
	}
	id, err := newUploadID()
	if err != nil || len(id) != 32 {
		t.Fatalf("newUploadID=%q err=%v", id, err)
	}

	ev := sanitizeWebhookEvents([]string{"user.create", "ping", "user.create", "nope"})
	if len(ev) != 1 || ev[0] != "user.create" {
		t.Fatalf("sanitizeWebhookEvents=%v", ev)
	}

	if formatInt64List(nil) != "" || formatInt64List([]int64{1, 2}) != "1,2" {
		t.Fatal("formatInt64List")
	}

	if mapEnabled(true) != "enabled" || mapEnabled(false) != "disabled" {
		t.Fatal("mapEnabled")
	}

	mime := buildMIMEText("a@b.c", "d@e.f", "subj", "body")
	if !strings.Contains(string(mime), "Subject: subj") || !strings.Contains(string(mime), "body") {
		t.Fatalf("buildMIMEText=%s", mime)
	}

	srv, store := testServer(t)
	uname, err := srv.uniqueUsername(context.Background(), "helperuser")
	if err != nil || uname != "helperuser" {
		t.Fatalf("uniqueUsername=%q err=%v", uname, err)
	}
	hash, err := auth.HashPassword("longpassword")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreateUser(context.Background(), "helperuser", hash, false); err != nil {
		t.Fatal(err)
	}
	uname2, err := srv.uniqueUsername(context.Background(), "helperuser")
	if err != nil || uname2 != "helperuser-2" {
		t.Fatalf("uniqueUsername taken=%q err=%v", uname2, err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/auth/users/3/libraries", nil)
	req.SetPathValue("id", "3")
	uid, ok := userPathID(httptest.NewRecorder(), req)
	if !ok || uid != 3 {
		t.Fatalf("userPathID=%d ok=%v", uid, ok)
	}
	req.SetPathValue("id", "0")
	if _, ok := userPathID(httptest.NewRecorder(), req); ok {
		t.Fatal("userPathID invalid")
	}

	req = httptest.NewRequest(http.MethodGet, "/api/libraries/9", nil)
	req.SetPathValue("id", "9")
	lid, ok := libraryPathID(httptest.NewRecorder(), req)
	if !ok || lid != 9 {
		t.Fatalf("libraryPathID=%d ok=%v", lid, ok)
	}

	q, err := srv.opdsBookQuery(httptest.NewRequest(http.MethodGet, "/opds/recent", nil), models.BookQuery{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if q.Limit != 10 {
		t.Fatalf("opdsBookQuery limit=%d", q.Limit)
	}
}

func TestSanitizeUploadRelPathAttacks(t *testing.T) {
	cases := []struct {
		in      string
		wantErr bool
		want    string
	}{
		{"books/a.pdf", false, "books/a.pdf"},
		{"/books/a.pdf", false, "books/a.pdf"},
		{"../escape.pdf", true, ""},
		{"books/../../etc/passwd", true, ""},
		{"", true, ""},
		{".", true, ""},
		{"books/./a.pdf", false, "books/a.pdf"},
	}
	for _, tc := range cases {
		got, err := sanitizeUploadRelPath(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Fatalf("%q: expected error, got %q", tc.in, got)
			}
			continue
		}
		if err != nil || got != tc.want {
			t.Fatalf("%q: got %q err=%v want %q", tc.in, got, err, tc.want)
		}
	}
}

func TestParseContentRangeBounds(t *testing.T) {
	start, end, total, err := parseContentRange("bytes 0-9/100")
	if err != nil || start != 0 || end != 9 || total != 100 {
		t.Fatalf("got %d-%d/%d err=%v", start, end, total, err)
	}
	if _, _, _, err := parseContentRange("bytes 5-4/10"); err == nil {
		t.Fatal("expected invalid range where end < start")
	}
	if _, _, _, err := parseContentRange("bytes 0-9/0"); err == nil {
		t.Fatal("expected invalid total 0")
	}
	if _, _, _, err := parseContentRange("0-9/100"); err == nil {
		t.Fatal("expected missing bytes prefix")
	}
}
