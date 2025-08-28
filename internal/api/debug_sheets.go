package api

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/Shyyw1e/mpstats-sync-go/internal/config"
	"github.com/Shyyw1e/mpstats-sync-go/internal/sheets"
)

func HandleDebugSheets(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Query().Get("slug")
	if slug == "" {
		http.Error(w, "use ?slug=...", http.StatusBadRequest); return
	}
	cat, err := config.LoadBySlug(slug)
	if err != nil { http.Error(w, "config: "+err.Error(), http.StatusBadRequest); return }

	credsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credsPath == "" { http.Error(w, "GOOGLE_APPLICATION_CREDENTIALS required", 500); return }
	b, err := os.ReadFile(credsPath)
	if err != nil { http.Error(w, "read creds: "+err.Error(), 500); return }

	ssID := os.Getenv("SPREADSHEET_ID")
	if ssID == "" { http.Error(w, "SPREADSHEET_ID required", 500); return }

	cl, err := sheets.New(r.Context(), b, ssID)
	if err != nil { http.Error(w, "sheets init: "+err.Error(), 500); return }

	if err := cl.EnsureHeaders(cat.Sheet, cat.Headers); err != nil {
		http.Error(w, "ensure headers: "+err.Error(), 502); return
	}
	if err := cl.WriteTest(cat.Sheet); err != nil {
		http.Error(w, "write test: "+err.Error(), 502); return
	}

	w.Header().Set("Content-Type","application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok": true, "sheet": cat.Sheet, "wrote": "A2",
	})
}
