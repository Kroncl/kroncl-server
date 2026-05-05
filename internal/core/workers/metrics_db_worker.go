package coreworkers

import (
	"context"
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

func NewDbWorker(service *Service, interval string) *Worker {
	return &Worker{
		service:  service,
		cron:     cron.New(),
		interval: interval,
	}
}

func (w *Worker) StartDbMetricsWorker() error {
	_, err := w.cron.AddFunc(w.interval, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		stats, err := w.service.CollectMetrics(ctx)
		if err != nil {
			log.Printf("❌ Failed to collect metrics: %v", err)
			return
		}

		if err := w.service.SaveMetricsSnapshot(ctx, stats); err != nil {
			log.Printf("❌ Failed to save metrics: %v", err)
			return
		}

		log.Printf("📊 Metrics saved: size=%dMB, schemas=%d, companies=%d",
			stats.TotalDatabaseSizeMB, stats.TotalSchemasCount, stats.CompanySchemasCount)
	})

	if err != nil {
		return err
	}

	w.cron.Start()
	log.Printf("✅ Metrics worker started with interval: %s", w.interval)
	return nil
}
