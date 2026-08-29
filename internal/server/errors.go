package server

import (
	"net/http"
	"net/url"
	"strings"
)

func prefersHTML(r *http.Request) bool {
	if r.Header.Get("Sec-Fetch-Mode") == "navigate" {
		return true
	}
	if r.Header.Get("X-Requested-With") == "XMLHttpRequest" {
		return false
	}
	accept := r.Header.Get("Accept")
	if accept == "" {
		return false
	}
	if strings.Contains(accept, "application/json") && !strings.Contains(accept, "text/html") {
		return false
	}
	return strings.Contains(accept, "text/html")
}

func redirectAPIError(w http.ResponseWriter, r *http.Request, status int) bool {
	if r == nil || !strings.HasPrefix(r.URL.Path, "/api/") || !prefersHTML(r) {
		return false
	}
	switch status {
	case http.StatusUnauthorized:
		http.Redirect(w, r, loginRedirectURL(r, "required"), http.StatusSeeOther)
		return true
	case http.StatusForbidden:
		http.Redirect(w, r, "/error/forbidden", http.StatusSeeOther)
		return true
	case http.StatusNotFound:
		http.Redirect(w, r, "/error/not-found", http.StatusSeeOther)
		return true
	default:
		if status >= 500 {
			http.Redirect(w, r, "/error/server", http.StatusSeeOther)
			return true
		}
	}
	return false
}

func loginRedirectURL(r *http.Request, reason string) string {
	params := url.Values{}
	params.Set("reason", reason)
	next := r.URL.Path
	if r.URL.RawQuery != "" {
		next += "?" + r.URL.RawQuery
	}
	if next != "" && next != "/login" && next != "/setup" {
		params.Set("next", next)
	}
	return "/login?" + params.Encode()
}

func writeErrorReq(w http.ResponseWriter, r *http.Request, status int, err error) {
	if redirectAPIError(w, r, status) {
		return
	}
	writeError(w, status, err)
}
