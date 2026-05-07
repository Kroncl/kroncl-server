package coreworkers

import (
	"context"
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

func NewServerWorker(service *Service, interval string) *Worker {
	return &Worker{
		service:  service,
		cron:     cron.New(),
		interval: interval,
	}
}

func (w *Worker) StartServerMetricsWorker() error {
	_, err := w.cron.AddFunc(w.interval, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		stats, err := w.service.CollectServerMetrics(ctx)
		if err != nil {
			log.Printf("❌ Failed to collect server metrics: %v", err)
			// НЕ ставим флаги воркеров здесь!
			return
		}

		if err := w.service.SaveServerMetricsSnapshot(ctx, stats); err != nil {
			log.Printf("❌ Failed to save server metrics: %v", err)
			return
		}

		log.Printf("📊 Server metrics saved: requests=%d, goroutines=%d, heap=%dMB",
			stats.RequestsTotal, stats.GoroutinesCount, stats.HeapAllocMB)
	})

	if err != nil {
		return err
	}

	w.cron.Start()
	log.Printf("✅ Server metrics worker started with interval: %s", w.interval)
	return nil
}
