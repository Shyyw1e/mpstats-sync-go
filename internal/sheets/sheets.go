package sheets

import (
	"context"
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

func (c *Client) WriteTest(sheet string) error {
	vr := &googlesheets.ValueRange{Values: [][]interface{}{{"ping", time.Now().Format(time.RFC3339)}}}
	_, err := c.svc.Spreadsheets.Values.
		Update(c.id, sheet+"!A2", vr).
		ValueInputOption("RAW").Do()
	return err
}

func toIface(ss []string) []interface{} {
	out := make([]interface{}, len(ss))
	for i, s := range ss { out[i] = s }
	return out
}