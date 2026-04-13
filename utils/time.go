package utils

import (
	"fmt"
	"log"
	"time"
)

var LocationCache = make(map[string]*time.Location)

func FormatTimeInLocation(t time.Time, locationCode string, format string) (string, error) {
	loc, exists := LocationCache[locationCode]
	if !exists {
		var err error
		loc, err = time.LoadLocation(locationCode)
		if err != nil {
			log.Printf("Warning: failed to load location %s: %v, falling back to UTC", locationCode, err)
			loc = time.UTC
			return t.In(loc).Format(format), fmt.Errorf("unknown location %s, using UTC", locationCode)
		}
		LocationCache[locationCode] = loc
	}

	return t.In(loc).Format(format), nil
}

func MustFormatTimeInLocation(t time.Time, locationCode string, format string) string {
	s, err := FormatTimeInLocation(t, locationCode, format)
	if err != nil {
		panic(err)
	}
	return s
}

func GetTimeInLocation(locationCode string) (time.Time, error) {
	loc, exists := LocationCache[locationCode]
	if !exists {
		var err error
		loc, err = time.LoadLocation(locationCode)
		if err != nil {
			log.Printf("Warning: failed to load location %s: %v, falling back to UTC", locationCode, err)
			return time.Now().UTC(), err
		}
		LocationCache[locationCode] = loc
	}
	return time.Now().In(loc), nil
}
