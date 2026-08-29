package server

import (
	"errors"
	"fmt"
	"html"
	"log/slog"
	"net/http"
	"path"
	"strings"

	"athenaeum/internal/telemetry"
)

var errPanic = errors.New("internal server error")

// looksLikeStaticAsset reports whether the path looks like a missing file
// (has an extension) rather than an SPA client route.
func looksLikeStaticAsset(name string) bool {
	base := path.Base(name)
	if base == "." || base == "/" || base == "" {
		return false
	}
	return strings.Contains(base, ".")
}

func writeHTMLError(w http.ResponseWriter, status int, title, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	safeTitle := html.EscapeString(title)
	safeMessage := html.EscapeString(message)
	_, _ = fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1"/>
<title>%s</title>
<style>
body{font-family:system-ui,sans-serif;margin:0;min-height:100vh;display:grid;place-items:center;background:#0a0a0b;color:#e8e8ea}
main{max-width:28rem;padding:2rem;text-align:center}
h1{font-size:1.5rem;margin:0 0 .75rem}
p{color:#a1a1aa;margin:0 0 1.5rem;line-height:1.5}
a{color:#e8e8ea}
</style>
</head>
<body>
<main>
<h1>%s</h1>
<p>%s</p>
<p><a href="/">Back to library</a></p>
</main>
</body>
</html>`, safeTitle, safeTitle, safeMessage)
}

func writePanicResponse(w http.ResponseWriter, r *http.Request) {
	if prefersHTML(r) {
		writeHTMLError(w, http.StatusInternalServerError, "Something went wrong",
			"The server hit an unexpected error. Try again in a moment.")
		return
	}
	writeError(w, http.StatusInternalServerError, errPanic)
}

// recoverMiddleware catches panics so a single bad request cannot take down the process.
// It sits outside Sentry middleware so Repanic still reaches this handler.
func recoverMiddleware(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				rec := recover()
				if rec == nil {
					return
				}
				err := fmt.Errorf("panic: %v", rec)
				if log != nil {
					log.Error("panic recovered", "err", err, "path", r.URL.Path)
				}
				telemetry.CaptureException(err)
				writePanicResponse(w, r)
			}()
			next.ServeHTTP(w, r)
		})
	}
}
