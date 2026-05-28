package coreworkers

import (
	"context"
	"log"
	"time"

	"kroncl-server/internal/metrics"

	"github.com/robfig/cron/v3"
)

func NewMediaWorker(service *Service, interval string) *Worker {
	return &Worker{
		service:  service,
		cron:     cron.New(),
		interval: interval,
	}
}

func (w *Worker) StartMediaMetricsWorker() error {
	_, err := w.cron.AddFunc(w.interval, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		stats, err := w.service.CollectMediaMetrics(ctx)
		if err != nil {
			log.Printf("❌ Failed to collect media metrics: %v", err)
			metrics.SetMediaWorkerLastSuccess(false)
			return
		}

		if err := w.service.SaveMediaMetricsSnapshot(ctx, stats); err != nil {
			log.Printf("❌ Failed to save media metrics: %v", err)
			metrics.SetMediaWorkerLastSuccess(false)
			return
		}

		metrics.SetMediaWorkerLastSuccess(true)

		log.Printf("📊 Media metrics saved: buckets=%d, objects=%d, size=%dMB, tenants=%d",
			stats.TotalBuckets, stats.TotalObjects, stats.TotalSizeMB, stats.TenantBucketsCount)
	})

	if err != nil {
		return err
	}

	w.cron.Start()
	log.Printf("✅ Media metrics worker started with interval: %s", w.interval)
	return nil
}
