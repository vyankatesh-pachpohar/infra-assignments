package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/lib/pq"

	"github.com/vyankatesh-pachpohar/infra-assignments/internal/handler"
	"github.com/vyankatesh-pachpohar/infra-assignments/internal/repository"
	"github.com/vyankatesh-pachpohar/infra-assignments/internal/service"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "configdb")
	appPort := getEnv("APP_PORT", "8080")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		slog.Error("failed to open db connection", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	repo := repository.NewConfigRepository(db)
	svc := service.NewConfigService(repo)
	h := handler.NewConfigHandler(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping", h.Ping)
	mux.HandleFunc("GET /configs/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		h.GetConfig(w, r, id)
	})
	mux.HandleFunc("POST /configs", h.UpsertConfig)

	slog.Info("starting server", "port", appPort, "db_host", dbHost)
	if err := http.ListenAndServe(":"+appPort, handler.LoggingMiddleware(mux)); err != nil {	
	slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
