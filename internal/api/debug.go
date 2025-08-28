package api

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/Shyyw1e/mpstats-sync-go/internal/config"
	"github.com/Shyyw1e/mpstats-sync-go/internal/mpstats"
	"github.com/Shyyw1e/mpstats-sync-go/pkg/logger"
)


func HandleDebugExtract(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Query().Get("slug")
	sku := r.URL.Query().Get("sku")
	if slug == "" || sku == "" {
		logger.Log.Errorf("invalid values:\nslug:%v\tsku:%v", slug, sku)
		http.Error(w, "use ?slug=...&sku=...", http.StatusBadRequest)
		return
	}

	cat, err := config.LoadBySlug(slug)
	if err != nil {
		logger.Log.Errorf("failed to load by slug: %v", err)
		http.Error(w, "config: "+err.Error(), http.StatusBadRequest)
		return
	}

	token := getenv("MPSTATS_API_TOKEN", "")
	if token == "" {
		logger.Log.Errorf("empty API token")
		http.Error(w, "MPSTATS_API_TOKEN required", http.StatusInternalServerError)
		return
	}
	client := mpstats.New(token)

	now := time.Now()
	d2 := now.Format("2006-01-02")
	d1 := now.AddDate(0, 0, -10000).Format("2006-01-02")
	ctx := r.Context()

	vers, err := client.GetVersions(ctx, sku, d1, d2)
	if err != nil || len(vers) == 0 {
		logger.Log.Errorf("invalid ")
		http.Error(w, "versions: "+errMsg(err, "empty"), http.StatusBadGateway)
		return
	}
	fp, err := client.GetFullPage(ctx, sku, vers[0].Version)
	if err != nil {
		http.Error(w, "full_page: "+err.Error(), http.StatusBadGateway)
		return
	}

	kv := mpstats.ExtractByMapping(fp, cat.FieldMapping)
	kv["sku"] = sku

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(kv)
}

func errMsg(err error, fallback string) string {
	if err != nil { return err.Error() }
	return fallback
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" { return v }
	return def
}
