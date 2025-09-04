package worker

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/Shyyw1e/mpstats-sync-go/internal/mpstats"
)

type Job struct {
	Index int
	SKU   string
}

type Result struct {
	Index int
	Row   []interface{}
	Err   error
}

func ProcessSKUs(
	ctx context.Context,
	client *mpstats.Client,
	headers []string,
	mapping map[string]string,
	skus []string,
	workers int, // напр. 48
	rps int,     // напр. 8
) ([][]interface{}, []error) {

	if workers <= 0 { workers = 32 }
	if rps <= 0 { rps = 1 }

	limiter := time.NewTicker(time.Second / time.Duration(rps))
	defer limiter.Stop()
//	waitToken := func() { <-limiter.C }
	

	jobs := make(chan Job)
	results := make(chan Result)

	rows := make([][]interface{}, len(skus))
	errs := make([]error, len(skus))

	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func() {
			defer wg.Done()
			for job := range jobs {
				select {
				case <-ctx.Done():
					return
				case <-limiter.C:
				}
				oneCtx, cancel := context.WithTimeout(ctx, 45*time.Second) // общий лимит на один SKU
				info, err := mpstats.FetchOne(oneCtx, client, job.SKU, mapping)
				cancel()

				var row []interface{}
				if err == nil {
					row = mpstats.BuildRow(headers, info)
				}
				results <- Result{Index: job.Index, Row: row, Err: err}
			}
		}()
	}

	go func() {
		for i, s := range skus {
			sku := strings.TrimSpace(s)
			if sku == "" {
				continue
			}
			jobs <- Job{Index: i, SKU: sku}
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	for res := range results {
		rows[res.Index] = res.Row
		errs[res.Index] = res.Err
	}

	return rows, errs
}
