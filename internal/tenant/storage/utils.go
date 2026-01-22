package storage

import (
	"fmt"
	"time"
)

// formatBytes форматирует байты в читаемый вид
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// formatTime форматирует время в строку
func formatTime(t *time.Time) *string {
	if t == nil || t.IsZero() {
		return nil
	}
	formatted := t.Format(time.RFC3339)
	return &formatted
}
