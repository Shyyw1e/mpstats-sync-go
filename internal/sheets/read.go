package sheets

import (
	"context"
	"fmt"
	"strings"

	"github.com/Shyyw1e/mpstats-sync-go/pkg/logger"
)


func (c *Client) ReadColumn(ctx context.Context, sheet string) ([]string, error) {
	rng := fmt.Sprintf("%s!A2:A", sheet)
	resp, err := c.svc.Spreadsheets.Values.Get(c.id, rng).Do()
	if err != nil {
		logger.Log.Errorf("failed to get range of values: %v", err)
		return nil, err
	}
	
	out := make([]string, 0, len(resp.Values))
	for _, row := range resp.Values {
		if len(row) == 0 {
			logger.Log.Warn("empty sku: skipping")
			continue
		}
		sku := strings.TrimSpace(fmt.Sprint(row[0]))
		if sku == "" {
			logger.Log.Warn("empty sku: skipping")
			continue
		}
		out = append(out, sku)
	}

	logger.Log.Infof("Read sku column.\tGot skus:%v", len(out))
	return out, nil
}