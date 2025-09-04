package mpstats

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Shyyw1e/mpstats-sync-go/pkg/logger"
)

func (c *Client) GetVersions(ctx context.Context, sku, d1, d2 string) (VersionsResp, error) {
	url := fmt.Sprintf("%s/item/%s/full_page/versions?d1=%s&d2=%s", baseURL, sku, d1, d2)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil { return nil, err }

	if c.token != "" { req.Header.Set("X-Mpstats-TOKEN", c.token) }

	res, err := c.http.Do(req)
	if err != nil { return nil, err }
	defer res.Body.Close()

	if res.StatusCode == http.StatusTooManyRequests {
		// 429 — подскажем, сколько ждать
		logger.Log.Warn("Too Many requests: 429")
		return nil, RateLimitErr{Wait: parseRetryAfter(res.Header), Status: res.StatusCode}
	}
	if res.StatusCode >= 300 {
		return nil, fmt.Errorf("versions %s: %s", sku, res.Status)
	}

	var out VersionsResp
	return out, json.NewDecoder(res.Body).Decode(&out)
}

func (c *Client) GetFullPage(ctx context.Context, sku, version string) (*FullPage, error) {
	url := fmt.Sprintf("%s/item/%s/full_page?version=%s", baseURL, sku, version)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if c.token != "" { req.Header.Set("X-Mpstats-TOKEN", c.token) }

	res, err := c.http.Do(req)
	if err != nil { return nil, err }
	defer res.Body.Close()

	if res.StatusCode == http.StatusTooManyRequests {
		return nil, RateLimitErr{Wait: parseRetryAfter(res.Header), Status: res.StatusCode}
	}
	if res.StatusCode >= 300 {
		return nil, fmt.Errorf("full_page %s: %s", sku, res.Status)
	}

	var fp FullPage
	return &fp, json.NewDecoder(res.Body).Decode(&fp)
}