package api

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"net/http"
	"os"

	"github.com/Shyyw1e/mpstats-sync-go/internal/config"
	"github.com/Shyyw1e/mpstats-sync-go/internal/mpstats"
	"github.com/Shyyw1e/mpstats-sync-go/internal/sheets"
	"github.com/Shyyw1e/mpstats-sync-go/internal/worker"
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


func HandleSync(w http.ResponseWriter, r *http.Request) {
    slug := chi.URLParam(r, "slug")
    if slug == "" {
        http.Error(w, "empty slug", http.StatusBadRequest)
        return
    }

    cat, err := config.LoadBySlug(slug)
    if err != nil {
        http.Error(w, "config: "+err.Error(), http.StatusBadRequest)
        return
    }

    base := context.Background()

    token := os.Getenv("MPSTATS_API_TOKEN")
    sprID := os.Getenv("SPREADSHEET_ID")
    credJSONPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")

    credentialsJSON, err := os.ReadFile(credJSONPath)
    if err != nil {
        http.Error(w, "credentials: "+err.Error(), http.StatusInternalServerError)
        return
    }
    sc, err := sheets.New(base, credentialsJSON, sprID)
    if err != nil {
        http.Error(w, "sheets client: "+err.Error(), http.StatusInternalServerError)
        return
    }

    skus, err := sc.ReadColumn(base, cat.Sheet)
    if err != nil {
        http.Error(w, "read column: "+err.Error(), http.StatusInternalServerError)
        return
    }

    mpclient := mpstats.New(token)

    workers := atoiDefault(os.Getenv("WORKERS"), 48)
    rps     := atoiDefault(os.Getenv("RPS"),     8)

    const chunkSize = 100 
    startRow := 2

    var total, ok, fail int
    headers := cat.Headers

    for i := 0; i < len(skus); i += chunkSize {
        j := i + chunkSize
        if j > len(skus) { j = len(skus) }
        chunk := skus[i:j]

        chunkCtx, cancel := context.WithTimeout(base, 3*time.Minute)
        rows, errs := worker.ProcessSKUs(chunkCtx, mpclient, headers, cat.FieldMapping, chunk, workers, rps)
        cancel()

        for _, e := range errs {
            if e == nil { ok++ } else { fail++; logger.Log.Warnf("sku failed: %v", e) }
        }
        total += len(chunk)

        if err := sc.WriteRows(context.Background(), cat.Sheet, startRow+i, rows); err != nil {
            http.Error(w, "write rows: "+err.Error(), http.StatusBadGateway)
            return
        }
    }

    resp := map[string]any{
        "slug":  slug,
        "total": total,
        "ok":    ok,
        "fail":  fail,
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}

func atoiDefault(s string, def int) int {
    if s == "" { return def }
    n, err := strconv.Atoi(s)
    if err != nil || n == 0 { return def }
    return n
}
