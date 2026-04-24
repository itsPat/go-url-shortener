package httpapi

import (
	"io"
	"net/http"
)

func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "ok")
}
