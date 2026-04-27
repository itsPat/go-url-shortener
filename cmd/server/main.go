package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

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
		slog.Error("failed to load env file", "err", envErr)
		os.Exit(1)
	}

	dbUri := os.Getenv("DB_CONNECTION_URI")
	if dbUri == "" {
		slog.Error("DB_CONNECTION_URI not set or empty")
		os.Exit(1)
	}

	db, dbErr := gorm.Open(postgres.Open(dbUri), &gorm.Config{})
	if dbErr != nil {
		slog.Error("db error", "err", dbErr)
		os.Exit(1)
	}

	migrationErr := db.AutoMigrate(&links.Link{})
	if migrationErr != nil {
		slog.Error("db migration error", "err", migrationErr)
		os.Exit(1)
	}

	linkStore := links.NewDatabase(db)

	server := httpapi.NewServer(linkStore)
	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: server.Handler(),
	}

	go func() {
		slog.Info("server listening", "port", port)
		serverErr := srv.ListenAndServe()

		if !errors.Is(serverErr, http.ErrServerClosed) {
			slog.Error("server error", "err", serverErr)
			os.Exit(1)
		}
	}()

	// Listen for stop signals
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	slog.Info("server shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Stop accepting requests
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", "err", err)
	} else {
		slog.Info("server shutdown gracefully")
	}

	// Close DB
	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("get db handle error", "err", err)
		return
	}
	if err := sqlDB.Close(); err != nil {
		slog.Error("db close error", "err", err)
	}
	slog.Info("db shutdown gracefully")
}
