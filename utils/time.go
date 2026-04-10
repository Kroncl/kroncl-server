package utils

import (
	"log"
	"time"
)

var MoscowLocation *time.Location

func init() {
	var err error
	MoscowLocation, err = time.LoadLocation("Europe/Moscow")
	if err != nil {
		log.Printf("Warning: failed to load Europe/Moscow location: %v, using UTC+3 fixed zone", err)
		MoscowLocation = time.FixedZone("MSK", 3*60*60)
	}
}

// GetMoscowTime возвращает текущее время в московском часовом поясе
func GetMoscowTime() time.Time {
	return time.Now().In(MoscowLocation)
}

// FormatMoscowTime форматирует время в московском поясе
func FormatMoscowTime(t time.Time, format string) string {
	return t.In(MoscowLocation).Format(format)
}
