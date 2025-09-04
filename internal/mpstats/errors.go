package mpstats

import "time"

type RateLimitErr struct {
	Wait   time.Duration
	Status int
}

func (e RateLimitErr) Error() string {
	if e.Wait > 0 {
		return "rate limited: wait " + e.Wait.String()
	}
	return "rate limited"
}
