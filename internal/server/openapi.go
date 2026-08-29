package server

import (
	"net/http"
	"strings"

	"athenaeum/internal/brand"
	"athenaeum/internal/version"
)

type openAPISpec struct {
	OpenAPI    string                     `json:"openapi"`
	Info       openAPIInfo                `json:"info"`
	Servers    []openAPIServer            `json:"servers"`
	Paths      map[string]openAPIPathItem `json:"paths"`
	Components openAPIComponents          `json:"components"`
	Tags       []openAPITag               `json:"tags"`
}

// OpenAPIExport is the JSON-serializable OpenAPI document used by genapi.
type OpenAPIExport = openAPISpec

// BuildOpenAPI returns the OpenAPI 3 document derived from the API docs.
func BuildOpenAPI() OpenAPIExport {
	return openAPIFromDoc(apiDocumentation())
}

type openAPIInfo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

type openAPIServer struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

type openAPIPathItem map[string]openAPIOperation

type openAPIOperation struct {
	Tags        []string               `json:"tags,omitempty"`
	Summary     string                 `json:"summary"`
	Description string                 `json:"description,omitempty"`
	OperationID string                 `json:"operationId,omitempty"`
	Parameters  []openAPIParameter     `json:"parameters,omitempty"`
	RequestBody *openAPIRequestBody    `json:"requestBody,omitempty"`
	Responses   map[string]openAPIResp `json:"responses"`
	Security    []map[string][]string  `json:"security,omitempty"`
}

type openAPIParameter struct {
	Name        string         `json:"name"`
	In          string         `json:"in"`
	Required    bool           `json:"required,omitempty"`
	Description string         `json:"description,omitempty"`
	Schema      map[string]any `json:"schema"`
}

type openAPIRequestBody struct {
	Required bool                        `json:"required,omitempty"`
	Content  map[string]openAPIMediaType `json:"content"`
}

type openAPIMediaType struct {
	Schema map[string]any `json:"schema"`
}

type openAPIResp struct {
	Description string `json:"description"`
}

type openAPIComponents struct {
	SecuritySchemes map[string]openAPISecurityScheme `json:"securitySchemes"`
}

type openAPISecurityScheme struct {
	Type        string `json:"type"`
	Scheme      string `json:"scheme,omitempty"`
	Name        string `json:"name,omitempty"`
	In          string `json:"in,omitempty"`
	Description string `json:"description,omitempty"`
}

type openAPITag struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

func openAPIFromDoc(doc APIDoc) openAPISpec {
	ver := strings.TrimSpace(version.Version)
	if ver == "" {
		ver = doc.Version
	}
	descParts := append([]string{}, doc.Auth...)
	descParts = append(descParts, doc.ContentTypes...)
	spec := openAPISpec{
		OpenAPI: "3.0.3",
		Info: openAPIInfo{
			Title:       doc.Title,
			Description: strings.Join(descParts, "\n\n"),
			Version:     ver,
		},
		Servers: []openAPIServer{{URL: "/", Description: "This server"}},
		Paths:   map[string]openAPIPathItem{},
		Components: openAPIComponents{
			SecuritySchemes: map[string]openAPISecurityScheme{
				"sessionCookie": {
					Type:        "apiKey",
					In:          "cookie",
					Name:        "access_token",
					Description: "Browser session after POST /api/auth/login",
				},
				"basicAuth": {
					Type:        "http",
					Scheme:      "basic",
					Description: "Username and password (OPDS and scripts)",
				},
				"apiKeyHeader": {
					Type:        "apiKey",
					In:          "header",
					Name:        "X-API-Key",
					Description: "API key (" + brand.APIKeyPrefix + "…)",
				},
				"bearerAuth": {
					Type:        "http",
					Scheme:      "bearer",
					Description: "API key as Bearer token",
				},
			},
		},
	}

	for _, section := range doc.Sections {
		spec.Tags = append(spec.Tags, openAPITag{Name: section.Title})
		for _, ep := range section.Endpoints {
			pathKey, pathParams := openAPIPath(ep.Path)
			item := spec.Paths[pathKey]
			if item == nil {
				item = openAPIPathItem{}
				spec.Paths[pathKey] = item
			}
			method := strings.ToLower(ep.Method)
			op := openAPIOperation{
				Tags:        []string{section.Title},
				Summary:     ep.Summary,
				OperationID: openAPIOpID(ep.Method, ep.Path),
				Responses: map[string]openAPIResp{
					"200": {Description: "Success"},
					"401": {Description: "Unauthorized"},
				},
			}
			if ep.Auth != "" && ep.Auth != "public" {
				op.Description = "Auth: " + ep.Auth
				op.Security = []map[string][]string{
					{"sessionCookie": {}},
					{"basicAuth": {}},
					{"apiKeyHeader": {}},
					{"bearerAuth": {}},
				}
			}
			for _, p := range pathParams {
				op.Parameters = append(op.Parameters, openAPIParameter{
					Name:     p,
					In:       "path",
					Required: true,
					Schema:   map[string]any{"type": "string"},
				})
			}
			for _, q := range splitCSV(ep.Query) {
				op.Parameters = append(op.Parameters, openAPIParameter{
					Name:   q,
					In:     "query",
					Schema: map[string]any{"type": "string"},
				})
			}
			if ep.Body != "" {
				schema := map[string]any{"type": "object"}
				if strings.Contains(strings.ToLower(ep.Body), "octet-stream") {
					schema = map[string]any{"type": "string", "format": "binary"}
					op.RequestBody = &openAPIRequestBody{
						Required: true,
						Content: map[string]openAPIMediaType{
							"application/octet-stream": {Schema: schema},
						},
					}
				} else {
					op.RequestBody = &openAPIRequestBody{
						Required: true,
						Content: map[string]openAPIMediaType{
							"application/json": {Schema: schema},
						},
					}
					op.Description = strings.TrimSpace(op.Description + "\nBody: " + ep.Body)
				}
			}
			if ep.Response != "" {
				op.Description = strings.TrimSpace(op.Description + "\nResponse: " + ep.Response)
			}
			item[method] = op
		}
	}
	return spec
}

func openAPIPath(path string) (string, []string) {
	parts := strings.Split(path, "/")
	var params []string
	for i, p := range parts {
		if strings.HasPrefix(p, "{") && strings.HasSuffix(p, "}") {
			name := strings.TrimSuffix(strings.TrimPrefix(p, "{"), "}")
			params = append(params, name)
			parts[i] = "{" + name + "}"
		}
	}
	return strings.Join(parts, "/"), params
}

func openAPIOpID(method, path string) string {
	clean := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			return r
		default:
			return '_'
		}
	}, method+"_"+path)
	return strings.Trim(clean, "_")
}

func splitCSV(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	var out []string
	for part := range strings.SplitSeq(s, ",") {
		p := strings.TrimSpace(part)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func (s *Server) handleOpenAPI(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, openAPIFromDoc(apiDocumentation()))
}

func (s *Server) handleDocsUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = w.Write([]byte(docsUIHTML))
}

func (s *Server) handleDocsAppJS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = w.Write([]byte(docsUIJS))
}

const docsUIHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1"/>
<title>API docs</title>
<style>
:root { color-scheme: light dark; --bg:#0f1115; --fg:#e8eaed; --muted:#9aa0a6; --card:#1a1d24; --border:#2a2f3a; --accent:#7aa2ff; --get:#3dd68c; --post:#6ea8fe; --put:#ffc107; --delete:#ff6b6b; }
@media (prefers-color-scheme: light) {
  :root { --bg:#f6f7f9; --fg:#1a1d24; --muted:#5f6368; --card:#fff; --border:#dadce0; --accent:#1a73e8; }
}
* { box-sizing: border-box; }
body { margin:0; font:14px/1.45 system-ui,sans-serif; background:var(--bg); color:var(--fg); }
header { position:sticky; top:0; z-index:2; display:flex; flex-wrap:wrap; gap:.75rem; align-items:center; padding:1rem 1.25rem; border-bottom:1px solid var(--border); background:color-mix(in srgb, var(--bg) 92%, transparent); backdrop-filter:blur(8px); }
h1 { margin:0; font-size:1.15rem; font-weight:650; }
.meta { color:var(--muted); font-size:.85rem; }
input[type=search] { flex:1; min-width:12rem; max-width:28rem; padding:.55rem .75rem; border:1px solid var(--border); border-radius:.5rem; background:var(--card); color:var(--fg); }
a { color:var(--accent); }
main { max-width:960px; margin:0 auto; padding:1.25rem; }
.auth, .section { background:var(--card); border:1px solid var(--border); border-radius:.75rem; padding:1rem; margin-bottom:1rem; }
.section h2 { margin:0 0 .75rem; font-size:1rem; }
.ep { display:grid; grid-template-columns:5.5rem 1fr; gap:.35rem .75rem; padding:.65rem 0; border-top:1px solid var(--border); }
.ep:first-of-type { border-top:0; }
.method { font:600 .75rem/1.8 ui-monospace,monospace; text-align:center; border-radius:.35rem; text-transform:uppercase; }
.method.get { background:color-mix(in srgb, var(--get) 22%, transparent); color:var(--get); }
.method.post { background:color-mix(in srgb, var(--post) 22%, transparent); color:var(--post); }
.method.put,.method.patch { background:color-mix(in srgb, var(--put) 22%, transparent); color:var(--put); }
.method.delete { background:color-mix(in srgb, var(--delete) 22%, transparent); color:var(--delete); }
.path { font:600 .85rem/1.4 ui-monospace,monospace; word-break:break-all; }
.summary { color:var(--muted); font-size:.85rem; }
.detail { grid-column:2; color:var(--muted); font-size:.8rem; }
.hidden { display:none !important; }
.error { color:var(--delete); padding:1rem; }
</style>
</head>
<body>
<header>
  <div>
    <h1 id="title">API docs</h1>
    <div class="meta"><a href="/api/openapi.json">openapi.json</a> · <a href="/api/docs">docs JSON</a> · <a href="/">app</a></div>
  </div>
  <input id="q" type="search" placeholder="Search endpoints…" autocomplete="off"/>
</header>
<main id="root"><p class="meta">Loading…</p></main>
<script src="/docs/app.js"></script>
</body>
</html>
`

const docsUIJS = `(function () {
  const root = document.getElementById("root");
  const title = document.getElementById("title");
  const q = document.getElementById("q");

  function esc(s) {
    return String(s || "")
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;");
  }

  function methodClass(m) {
    const x = String(m || "").toLowerCase();
    if (x === "get" || x === "post" || x === "put" || x === "patch" || x === "delete") return x;
    return "get";
  }

  function render(doc) {
    title.textContent = doc.title || "API docs";
    const auth = (doc.auth || []).map((line) => "<li>" + esc(line) + "</li>").join("");
    let html = '<section class="auth"><h2>Authentication</h2><ul>' + auth + "</ul></section>";
    for (const section of doc.sections || []) {
      html += '<section class="section" data-section="' + esc(section.title) + '"><h2>' + esc(section.title) + "</h2>";
      for (const ep of section.endpoints || []) {
        const hay = [ep.method, ep.path, ep.summary, ep.auth, ep.query, ep.body].join(" ").toLowerCase();
        html +=
          '<div class="ep" data-hay="' +
          esc(hay) +
          '"><div class="method ' +
          methodClass(ep.method) +
          '">' +
          esc(ep.method) +
          '</div><div><div class="path">' +
          esc(ep.path) +
          '</div><div class="summary">' +
          esc(ep.summary) +
          "</div>";
        const bits = [];
        if (ep.auth) bits.push("Auth: " + ep.auth);
        if (ep.query) bits.push("Query: " + ep.query);
        if (ep.body) bits.push("Body: " + ep.body);
        if (bits.length) html += '<div class="detail">' + esc(bits.join(" · ")) + "</div>";
        html += "</div></div>";
      }
      html += "</section>";
    }
    root.innerHTML = html;
  }

  function filter() {
    const needle = (q.value || "").trim().toLowerCase();
    for (const section of root.querySelectorAll(".section")) {
      let visible = 0;
      for (const ep of section.querySelectorAll(".ep")) {
        const show = !needle || (ep.getAttribute("data-hay") || "").includes(needle);
        ep.classList.toggle("hidden", !show);
        if (show) visible += 1;
      }
      section.classList.toggle("hidden", visible === 0);
    }
  }

  fetch("/api/docs")
    .then((r) => {
      if (!r.ok) throw new Error("Failed to load docs (" + r.status + ")");
      return r.json();
    })
    .then(render)
    .catch((err) => {
      root.innerHTML = '<p class="error">' + esc(err.message || String(err)) + "</p>";
    });

  q.addEventListener("input", filter);
})();
`
