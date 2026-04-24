package main

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
)

var port int = 8080


func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func (w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "ok")
	})

	slog.Info("Server Listening", "port", port)
	err := http.ListenAndServe(":" + strconv.Itoa(port), mux)

	if err != nil {
		slog.Error("Server Error", "msg", err)
		os.Exit(1)
	}
}