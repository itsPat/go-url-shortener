package main

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/itsPat/go-url-shortener/internal/httpapi"
	"github.com/itsPat/go-url-shortener/internal/links"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const port = 8080

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	envErr := godotenv.Load()
	if envErr != nil {
		slog.Error("Failed to load env file", "err", envErr)
		os.Exit(1)
	}

	dbUri := os.Getenv("DB_CONNECTION_URI")
	if dbUri == "" {
		slog.Error("DB_CONNECTION_URI not set or empty")
		os.Exit(1)
	}

	db, dbErr := gorm.Open(postgres.Open(dbUri), &gorm.Config{})
	if dbErr != nil {
		slog.Error("DB Error", "err", dbErr)
		os.Exit(1)
	}

	migrationErr := db.AutoMigrate(&links.Link{})
	if migrationErr != nil {
		slog.Error("DB Migration Error", "err", migrationErr)
		os.Exit(1)
	}

	linkStore := links.NewDatabase(db)

	server := httpapi.NewServer(linkStore)
	slog.Info("Server Listening", "port", port)
	serverErr := http.ListenAndServe(":"+strconv.Itoa(port), server.Handler())
	if serverErr != nil {
		slog.Error("Server Error", "err", serverErr)
		os.Exit(1)
	}
}
