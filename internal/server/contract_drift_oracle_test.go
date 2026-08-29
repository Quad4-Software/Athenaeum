package server

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// PROVED_CONTRACT_DOCUMENTED_ROUTES: every apiDocumentation endpoint is registered.

func TestContractDocumentedRoutesAreRegistered(t *testing.T) {
	registered := loadRegisteredRoutes(t)
	var missing []string
	for _, ep := range documentedEndpoints() {
		key := routeKey(ep.Method, ep.Path)
		if _, ok := registered[key]; !ok {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		t.Fatalf("documented routes missing from HandleFunc registrations:\n%s", strings.Join(missing, "\n"))
	}
}

func TestContractOpenAPIMatchesAPIDocs(t *testing.T) {
	doc := apiDocumentation()
	spec := openAPIFromDoc(doc)
	if spec.OpenAPI == "" || spec.Info.Title == "" {
		t.Fatal("openapi spec missing required fields")
	}

	want := map[string]struct{}{}
	for _, ep := range documentedEndpoints() {
		pathKey, _ := openAPIPath(ep.Path)
		want[routeKey(ep.Method, pathKey)] = struct{}{}
	}

	got := map[string]struct{}{}
	for path, item := range spec.Paths {
		for method := range item {
			got[routeKey(strings.ToUpper(method), path)] = struct{}{}
		}
	}

	var missing, extra []string
	for k := range want {
		if _, ok := got[k]; !ok {
			missing = append(missing, k)
		}
	}
	for k := range got {
		if _, ok := want[k]; !ok {
			extra = append(extra, k)
		}
	}
	if len(missing) > 0 || len(extra) > 0 {
		sort.Strings(missing)
		sort.Strings(extra)
		t.Fatalf("openapi drift vs api docs\nmissing:\n%s\nextra:\n%s",
			strings.Join(missing, "\n"), strings.Join(extra, "\n"))
	}
}

func TestContractOpenAPIHTTPEndpoint(t *testing.T) {
	srv, _ := testServer(t)
	handler, err := srv.Handler()
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/openapi.json", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["openapi"] == nil || body["paths"] == nil {
		t.Fatal("openapi.json missing openapi/paths")
	}
	paths, _ := body["paths"].(map[string]any)
	if len(paths) == 0 {
		t.Fatal("openapi.json has empty paths")
	}
}

func TestContractFrontendAPIPathsAreRegistered(t *testing.T) {
	registered := loadRegisteredRoutes(t)
	paths := loadFrontendAPIPaths(t)
	var missing []string
	for _, p := range paths {
		if frontendPathKnown(registered, p) {
			continue
		}
		missing = append(missing, p)
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		t.Fatalf("frontend API paths not registered:\n%s", strings.Join(missing, "\n"))
	}
}

func documentedEndpoints() []APIDocEndpoint {
	var out []APIDocEndpoint
	for _, sec := range apiDocumentation().Sections {
		out = append(out, sec.Endpoints...)
	}
	return out
}

func routeKey(method, path string) string {
	return strings.ToUpper(strings.TrimSpace(method)) + " " + strings.TrimSpace(path)
}

func loadRegisteredRoutes(t *testing.T) map[string]struct{} {
	t.Helper()
	dir := serverPackageDir(t)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	out := map[string]struct{}{}
	re := regexp.MustCompile(`^([A-Z]+)\s+(\S+)$`)
	for _, ent := range entries {
		name := ent.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || len(call.Args) < 1 {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "HandleFunc" {
				return true
			}
			lit, ok := call.Args[0].(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			pattern, err := strconv.Unquote(lit.Value)
			if err != nil {
				return true
			}
			m := re.FindStringSubmatch(pattern)
			if m == nil {
				return true
			}
			out[routeKey(m[1], m[2])] = struct{}{}
			return true
		})
	}
	if len(out) < 50 {
		t.Fatalf("expected many HandleFunc routes, got %d", len(out))
	}
	return out
}

func serverPackageDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Dir(file)
}

func repoRoot(t *testing.T) string {
	t.Helper()
	return filepath.Clean(filepath.Join(serverPackageDir(t), "../.."))
}

func loadFrontendAPIPaths(t *testing.T) []string {
	t.Helper()
	apiDir := filepath.Join(repoRoot(t), "web/src/lib/api")
	entries, err := os.ReadDir(apiDir)
	if err != nil {
		t.Fatal(err)
	}
	reSeg := regexp.MustCompile(`/\$\{[^}]+\}`)
	reOther := regexp.MustCompile(`\$\{[^}]+\}`)
	rePath := regexp.MustCompile(`/api/[A-Za-z0-9_./{}-]+`)
	reDigits := regexp.MustCompile(`/\d+(/|$)`)
	seen := map[string]struct{}{}
	var out []string
	for _, ent := range entries {
		name := ent.Name()
		if !strings.HasSuffix(name, ".ts") || strings.Contains(name, ".test.") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(apiDir, name))
		if err != nil {
			t.Fatal(err)
		}
		normalized := reSeg.ReplaceAllString(string(raw), "/{id}")
		normalized = reOther.ReplaceAllString(normalized, "")
		for _, m := range rePath.FindAllString(normalized, -1) {
			p := m
			if i := strings.IndexByte(p, '?'); i >= 0 {
				p = p[:i]
			}
			p = strings.TrimRight(p, ".")
			if reDigits.MatchString(p) {
				continue
			}
			p = canonicalizePathParams(p)
			if p == "" || !strings.HasPrefix(p, "/api/") {
				continue
			}
			if _, ok := seen[p]; ok {
				continue
			}
			seen[p] = struct{}{}
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		t.Fatal("no frontend /api paths found")
	}
	return out
}

func canonicalizePathParams(p string) string {
	parts := strings.Split(p, "/")
	for i, part := range parts {
		if part == "{id}" {
			continue
		}
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			continue
		}
		switch part {
		case "id", "bookId", "uploadId", "bookmarkId", "highlightId",
			"shareId", "tagId", "token", "locale", "page", "name", "document":
			parts[i] = "{" + part + "}"
		}
	}
	return strings.Join(parts, "/")
}

func frontendPathKnown(registered map[string]struct{}, path string) bool {
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE"}
	for _, m := range methods {
		if _, ok := registered[routeKey(m, path)]; ok {
			return true
		}
	}
	// Allow undocumented-but-registered dynamic forms by matching prefix families.
	for key := range registered {
		_, regPath, ok := strings.Cut(key, " ")
		if !ok {
			continue
		}
		if routeTemplateEqual(regPath, path) {
			return true
		}
	}
	return false
}

func routeTemplateEqual(a, b string) bool {
	as := strings.Split(a, "/")
	bs := strings.Split(b, "/")
	if len(as) != len(bs) {
		return false
	}
	for i := range as {
		ap, bp := as[i], bs[i]
		if ap == bp {
			continue
		}
		aDyn := strings.HasPrefix(ap, "{") && strings.HasSuffix(ap, "}")
		bDyn := strings.HasPrefix(bp, "{") && strings.HasSuffix(bp, "}")
		if aDyn || bDyn {
			continue
		}
		return false
	}
	return true
}
