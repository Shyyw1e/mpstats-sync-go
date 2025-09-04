package mpstats

import "github.com/Shyyw1e/mpstats-sync-go/pkg/logger"

func BuildRow(headers []string, info map[string]string) []interface{} {
	row := make([]interface{}, len(headers))
	for i, h := range headers {
		if v, ok := info[h]; ok && v != "" {
			row[i] = v
		} else {
			row[i] = ""
		}
	}
	logger.Log.Infof("Builded row:\t%v", row)
	return row
}
