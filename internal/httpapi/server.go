package httpapi

import (
	"net/http"

	"github.com/itsPat/go-url-shortener/internal/links"
)

type Server struct {
	linkStore links.Store
}

func NewServer(linkStore links.Store) *Server {
	return &Server{linkStore: linkStore}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthz)
	mux.HandleFunc("POST /shorten", s.shorten)
	mux.HandleFunc("GET /{code}", s.redirect)
	return mux
}
