package utils

import (
	"math"
	"net/url"
	"strconv"
)

// Builds query string for URL
func BuildQueries(merchant, category, channel, date, start, end string, min, max, limit int, summary bool) string {
	queries := map[string]string{
		"merchant": merchant,
		"category": category,
		"channel":  channel,
		"date":     date,
		"start":    start,
		"end":      end,
	}
	if min != math.MinInt64 {
		queries["min"] = strconv.Itoa(min)
	}
	if max != math.MaxInt64 {
		queries["max"] = strconv.Itoa(max)
	}
	if limit != 100 { // only if not default
		queries["limit"] = strconv.Itoa(limit)
	}
	if summary {
		queries["summary"] = "true"
	}

	q := url.Values{}
	for key, val := range queries {
		if val != "" {
			q.Set(key, val)
		}
	}

	return "?" + q.Encode()
}
