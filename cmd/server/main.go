package main

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/itsPat/go-url-shortener/internal/httpapi"
	"github.com/itsPat/go-url-shortener/internal/links"
)

const port = 8080

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	linkStore := links.NewInMemory()
	server := httpapi.NewServer(linkStore)

	slog.Info("Server Listening", "port", port)
	err := http.ListenAndServe(":"+strconv.Itoa(port), server.Handler())

	if err != nil {
		slog.Error("Server Error", "err", err)
		os.Exit(1)
	}
}
