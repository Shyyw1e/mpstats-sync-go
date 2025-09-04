package main

import (
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	"github.com/Shyyw1e/mpstats-sync-go/internal/api"
	"github.com/Shyyw1e/mpstats-sync-go/pkg/logger"
)

func main() {
	logger.InitLog("debug")

	if err := godotenv.Load(); err != nil {
		logger.Log.Errorf("failed to load .env: %v", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// debug-эндпоинт: проверяем MPStats+маппинг по одному SKU
	r.Get("/debug/extract", api.HandleDebugExtract)
	r.Get("/debug/sheets", api.HandleDebugSheets)
	// Заготовка под боевой запуск задачи
	r.Post("/sync/{slug}", api.HandleSync)

	if err := http.ListenAndServe(":8080", r); err != nil {
		logger.Log.Fatalf("listen: %v", err)
	}
	logger.Log.Info("Listening at :8080")
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" { return v }
	return def
}
