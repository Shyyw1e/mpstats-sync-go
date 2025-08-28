package api

import (
	"net/http"

	"github.com/Shyyw1e/mpstats-sync-go/internal/config"
	"github.com/Shyyw1e/mpstats-sync-go/pkg/logger"
	"github.com/go-chi/chi/v5"
)

func HandleStartSync(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		logger.Log.Error("invalid slug: empty")
		http.Error(w, "invalid slug", http.StatusBadRequest)
		return
	}

	if _, err := config.LoadBySlug(slug); err != nil {
		logger.Log.Errorf("failed to load by slug: %v", err)
		http.Error(w, "config:" + err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"queued":true}`))
}

