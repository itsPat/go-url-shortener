package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/itsPat/go-url-shortener/internal/links"
)

type createRequest struct {
	URL string `json:"url"`
}
type createResponse struct {
	Code     string `json:"code"`
	ShortURL string `json:"short_url"`
}

type linkResponse struct {
	Code      string    `json:"code"`
	URL       string    `json:"url"`
	Hits      int64     `json:"hits"`
	CreatedAt time.Time `json:"created_at"`
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "ok")
}

func (s *Server) shorten(w http.ResponseWriter, r *http.Request) {
	var body createRequest

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if body.URL == "" {
		writeJSONError(w, http.StatusBadRequest, "empty url")
		return
	}

	var link links.Link
	var err error

	for range 5 {
		code := links.NewCode()
		link, err = s.linkStore.Create(r.Context(), code, body.URL)
		if !errors.Is(err, links.ErrCodeTaken) {
			break
		}
	}

	if err != nil {
		slog.Error("shorten failed", "body", body, "err", err)
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}

	shortURL := "http://" + r.Host + "/" + link.Code

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createResponse{Code: link.Code, ShortURL: shortURL})
}

func (s *Server) redirect(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	link, err := s.linkStore.GetAndIncrement(r.Context(), code)

	if errors.Is(err, links.ErrNotFound) {
		writeJSONError(w, http.StatusNotFound, "code not found")
		return
	}
	if err != nil {
		slog.Error("redirect failed", "code", code, "err", err)
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}

	http.Redirect(w, r, link.URL, http.StatusFound)
}
