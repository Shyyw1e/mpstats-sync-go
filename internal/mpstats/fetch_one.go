package mpstats

import (
	"context"
	"errors"
	"fmt"

	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Shyyw1e/mpstats-sync-go/pkg/logger"
)

func dateRange() (string, string) {
	now := time.Now()
	logger.Log.Info("Returning time. . .")
	return now.AddDate(0, 0, -10000).Format("2006-01-02"), now.Format("2006-01-02")
}

func FetchOne(ctx context.Context, c *Client, sku string, mapping map[string]string) (map[string]string, error) {
	d1, d2 := dateRange()

	var lastErr error
	backoff := 2 * time.Second   // базовая задержка между попытками
	maxBackoff := 20 * time.Second

	for attempt := 1; attempt <= 5; attempt++ {
		// --- versions ---
		reqCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
		vers, err := c.GetVersions(reqCtx, sku, d1, d2)
		cancel()

		// если 429 — ждём столько, сколько просит сервер (или backoff)
		if rle, ok := err.(RateLimitErr); ok {
			wait := rle.Wait
			if wait <= 0 { wait = backoff }
			if wait > maxBackoff { wait = maxBackoff }
			time.Sleep(wait)
			backoff *= 2
			if backoff > maxBackoff { backoff = maxBackoff }
			lastErr = err
			continue
		}
		if err == nil && len(vers) == 0 {
			err = errors.New("versions: empty")
		}
		if err != nil {
			lastErr = err
			time.Sleep(backoff)
			if backoff < maxBackoff { backoff *= 2 }
			continue
		}

		// --- full page ---
		reqCtx2, cancel2 := context.WithTimeout(ctx, 20*time.Second)
		fp, err := c.GetFullPage(reqCtx2, sku, vers[0].Version)
		cancel2()

		if rle, ok := err.(RateLimitErr); ok {
			wait := rle.Wait
			if wait <= 0 { wait = backoff }
			if wait > maxBackoff { wait = maxBackoff }
			time.Sleep(wait)
			backoff *= 2
			if backoff > maxBackoff { backoff = maxBackoff }
			lastErr = err
			continue
		}
		if err != nil {
			lastErr = err
			time.Sleep(backoff)
			if backoff < maxBackoff { backoff *= 2 }
			continue
		}

		info := ExtractByMapping(fp, mapping)
		info["sku"] = strings.TrimSpace(sku)
		return info, nil
	}

	return nil, fmt.Errorf("fetchOne %s failed: %w", sku, lastErr)
}


func sleepWithJitter(d time.Duration) {
	time.Sleep(d)
}

// (опционально) вытащить Retry-After, если хочешь уважать 429 от сервера
func parseRetryAfter(h http.Header) time.Duration {
	raw := h.Get("Retry-After")
	if raw == "" { return 0 }
	if secs, err := strconv.Atoi(raw); err == nil {
		return time.Duration(secs) * time.Second
	}
	if t, err := time.Parse(time.RFC1123, raw); err == nil {
		return time.Until(t)
	}
	return 0
}
