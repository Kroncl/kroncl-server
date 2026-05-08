package coreworkers

import (
	"os"
	"strconv"
	"strings"
)

// getOpenFDs возвращает количество открытых файловых дескрипторов
func getOpenFDs() int {
	// Linux: /proc/self/fd
	f, err := os.Open("/proc/self/fd")
	if err != nil {
		return 0
	}
	defer f.Close()

	names, err := f.Readdirnames(-1)
	if err != nil {
		return 0
	}

	return len(names)
}

// getMemoryUsage возвращает RSS память процесса в MB
func getMemoryUsage() int {
	// Linux: /proc/self/statm
	data, err := os.ReadFile("/proc/self/statm")
	if err != nil {
		return 0
	}

	parts := strings.Split(string(data), " ")
	if len(parts) < 2 {
		return 0
	}

	// statm: size resident shared text lib data dt
	residentPages, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0
	}

	// страница обычно 4096 байт
	pageSize := int64(os.Getpagesize())
	residentBytes := residentPages * pageSize

	return int(residentBytes / 1024 / 1024)
}

// getCPUUsage возвращает использование CPU в процентах
// Примечание: полная реализация требует замера разницы между вызовами,
// для простоты возвращаем 0 или можно использовать runtime/metrics
func getCPUUsage() float64 {
	// Базовая реализация: можно вернуть 0 или добавить более сложную логику
	// с использованием runtime/metrics в будущем
	return 0.0
}
