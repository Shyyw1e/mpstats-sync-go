package sheets

import (
	"context"
	"fmt"
	"time"

	"github.com/Shyyw1e/mpstats-sync-go/pkg/logger"
	"google.golang.org/api/option"
	googlesheets "google.golang.org/api/sheets/v4"
)

type Client struct {
	svc *googlesheets.Service
	id string
}

func New(ctx context.Context, credentialsJSON []byte, spreadsheetid string) (*Client, error) {
	svc, err := googlesheets.NewService(ctx, 
		option.WithCredentialsJSON(credentialsJSON),
		option.WithScopes(googlesheets.SpreadsheetsScope),
		)
	if err != nil {
		logger.Log.Errorf("failed to create new google sheets service: %v", err)
		return nil ,err
	}	
	return &Client{svc: svc, id: spreadsheetid}, nil
}

func (c *Client) EnsureHeaders(sheet string, headers []string) error {
	vr := &googlesheets.ValueRange{Values: [][]interface{}{toIface(headers)}}
	_, err := c.svc.Spreadsheets.Values.
		Update(c.id, sheet+"!A1", vr).
		ValueInputOption("RAW").Do()
	return err
}

func (c *Client) ClearTail(ctx context.Context, sheet string, fromRow, toRow int) error {
    rng := fmt.Sprintf("%s!A%d:Z%d", sheet, fromRow, toRow)
    _, err := c.svc.Spreadsheets.Values.Clear(c.id, rng, &googlesheets.ClearValuesRequest{}).
        Context(ctx).Do()
    return err
}

func (c *Client) WriteTest(sheet string) error {
	vr := &googlesheets.ValueRange{Values: [][]interface{}{{"ping", time.Now().Format(time.RFC3339)}}}
	_, err := c.svc.Spreadsheets.Values.
		Update(c.id, sheet+"!A2", vr).
		ValueInputOption("RAW").Do()
	return err
}

func (c *Client) WriteRows(ctx context.Context, sheetName string, startRow int, rows [][]interface{}) error {
	if len(rows) == 0 {
		logger.Log.Info("empty rows")
		return nil
	}

	const chunkSize = 300
	const maxRetries = 3
	const baseDelay = 300 * time.Millisecond

	for i := 0; i < len(rows); i += chunkSize {
		end := i + chunkSize
		if end > len(rows) {
			end = len(rows)
		}
		chunk := rows[i:end]

		start := startRow + i
		endRow := start + len(chunk) - 1
		rangeStr := fmt.Sprintf("%s!A%d", sheetName, start)

		req := &googlesheets.ValueRange{Values: chunk}

		var err error
		delay := baseDelay
		for attempt := 1; attempt <= maxRetries; attempt++ {
			_, err = c.svc.Spreadsheets.Values.Update(c.id, rangeStr, req).
				ValueInputOption("RAW").Context(ctx).Do()
			if err == nil {
				break
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				delay *= 2
			}
		}
		if err != nil {
			err1 := fmt.Errorf("write rows %d-%d: %w", start, endRow, err)
			logger.Log.Error(err1)
			return err1
		}

		if end < len(rows) {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(700 * time.Millisecond):
			}
		}
	}

	return nil
}

func toIface(ss []string) []interface{} {
	out := make([]interface{}, len(ss))
	for i, s := range ss { out[i] = s }
	return out
}

