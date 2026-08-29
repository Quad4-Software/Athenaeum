package server

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/smtp"
	"strings"

	"athenaeum/internal/models"
)

func (s *Server) registerSMTPRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/admin/smtp", s.handleGetSMTP)
	mux.HandleFunc("PUT /api/admin/smtp", s.handlePutSMTP)
	mux.HandleFunc("GET /api/auth/kindle-email", s.handleGetKindleEmail)
	mux.HandleFunc("PUT /api/auth/kindle-email", s.handlePutKindleEmail)
	mux.HandleFunc("POST /api/books/{id}/send", s.handleSendBook)
}

func (s *Server) handleGetSMTP(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	cfg, err := s.store.GetSMTPSettings(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, cfg.Public())
}

func (s *Server) handlePutSMTP(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	var cfg models.SMTPSettings
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&cfg); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if cfg.Password == "" {
		existing, err := s.store.GetSMTPSettings(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		cfg.Password = existing.Password
	}
	if err := s.store.SaveSMTPSettings(r.Context(), cfg); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, cfg.Public())
}

func (s *Server) handleGetKindleEmail(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	email, err := s.store.GetKindleEmail(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"email": email})
}

func (s *Server) handlePutKindleEmail(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<12)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	userID := UserIDFromContext(r.Context())
	if err := s.store.SaveKindleEmail(r.Context(), userID, strings.TrimSpace(req.Email)); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"email": strings.TrimSpace(req.Email)})
}

func (s *Server) handleSendBook(w http.ResponseWriter, r *http.Request) {
	book, err := s.bookByIDChecked(w, r)
	if err != nil {
		return
	}
	var req struct {
		To     string `json:"to"`
		Kindle bool   `json:"kindle"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<12)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	to := strings.TrimSpace(req.To)
	userID := UserIDFromContext(r.Context())
	if req.Kindle || to == "" {
		email, err := s.store.GetKindleEmail(r.Context(), userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		if email == "" {
			writeError(w, http.StatusBadRequest, errors.New("kindle email not configured"))
			return
		}
		to = email
	}
	smtpCfg, err := s.store.GetSMTPSettings(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if !smtpCfg.Enabled || smtpCfg.Host == "" {
		writeError(w, http.StatusBadRequest, errors.New("smtp is not configured"))
		return
	}
	fs, err := s.openLibraryFS(r.Context(), book.LibraryID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	rc, err := fs.Open(r.Context(), book.RelPath)
	if err != nil {
		writeError(w, http.StatusNotFound, errors.New("file missing on disk"))
		return
	}
	data, err := io.ReadAll(rc)
	_ = rc.Close()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	filename := safeFilename(book)
	from := smtpCfg.FromAddr
	if from == "" {
		from = smtpCfg.Username
	}
	msg := buildMIMEAttachment(from, to, book.Title, filename, data)
	addr := fmt.Sprintf("%s:%d", smtpCfg.Host, smtpCfg.Port)
	var auth smtp.Auth
	if smtpCfg.Username != "" {
		auth = smtp.PlainAuth("", smtpCfg.Username, smtpCfg.Password, smtpCfg.Host)
	}
	if err := smtp.SendMail(addr, auth, from, []string{to}, msg); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "sent", "to": to})
}

func buildMIMEAttachment(from, to, subject, filename string, data []byte) []byte {
	boundary := "athenaeum-boundary"
	var b strings.Builder
	b.WriteString("From: " + from + "\r\n")
	b.WriteString("To: " + to + "\r\n")
	b.WriteString("Subject: " + subject + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: multipart/mixed; boundary=" + boundary + "\r\n\r\n")
	b.WriteString("--" + boundary + "\r\n")
	b.WriteString("Content-Type: text/plain; charset=utf-8\r\n\r\n")
	b.WriteString("Sent from Athenaeum\r\n\r\n")
	b.WriteString("--" + boundary + "\r\n")
	b.WriteString("Content-Type: application/octet-stream\r\n")
	b.WriteString("Content-Disposition: attachment; filename=\"" + filename + "\"\r\n")
	b.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")
	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(encoded, data)
	for i := 0; i < len(encoded); i += 76 {
		end := min(i+76, len(encoded))
		b.Write(encoded[i:end])
		b.WriteString("\r\n")
	}
	b.WriteString("--" + boundary + "--\r\n")
	return []byte(b.String())
}
