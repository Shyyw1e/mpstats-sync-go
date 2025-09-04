package api

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/Shyyw1e/mpstats-sync-go/internal/config"
	"github.com/Shyyw1e/mpstats-sync-go/internal/sheets"
	"github.com/go-chi/chi/v5"
)

func HandleDebugSheets(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		http.Error(w, "use ?slug=...", http.StatusBadRequest); return
	}
	cat, err := config.LoadBySlug(slug)
	if err != nil { http.Error(w, "config: "+err.Error(), http.StatusBadRequest); return }

	credsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credsPath == "" { http.Error(w, "GOOGLE_APPLICATION_CREDENTIALS required", http.StatusInternalServerError); return }
	b, err := os.ReadFile(credsPath)
	if err != nil { http.Error(w, "read creds: "+err.Error(), http.StatusInternalServerError); return }

	ssID := os.Getenv("SPREADSHEET_ID")
	if ssID == "" { http.Error(w, "SPREADSHEET_ID required", http.StatusInternalServerError); return }

	cl, err := sheets.New(r.Context(), b, ssID)
	if err != nil { http.Error(w, "sheets init: "+err.Error(), http.StatusInternalServerError); return }

	if err := cl.EnsureHeaders(cat.Sheet, cat.Headers); err != nil {
		http.Error(w, "ensure headers: "+err.Error(), http.StatusBadGateway); return
	}
	if err := cl.WriteTest(cat.Sheet); err != nil {
		http.Error(w, "write test: "+err.Error(), http.StatusBadGateway); return
	}

	w.Header().Set("Content-Type","application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok": true, "sheet": cat.Sheet, "wrote": "A2",
	})
}
