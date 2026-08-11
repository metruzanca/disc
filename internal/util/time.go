package util

import (
	"fmt"
	"time"
)

var timeLayouts = []string{
	time.RFC3339,
	"2006-01-02 15:04",
	"2006-01-02 15:04:05",
	"2006-01-02T15:04",
	"2006-01-02 3:04 PM",
	"2006-01-02",
}

// ParseTime parses a user-provided time string using several common layouts.
func ParseTime(s string) (time.Time, error) {
	for _, layout := range timeLayouts {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("could not parse time '%s'; use a format like %q or RFC3339", s, timeLayouts[1])
}
