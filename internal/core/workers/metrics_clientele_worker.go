package coreworkers

import (
	"context"
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

func NewClienteleWorker(service *Service, interval string) *Worker {
	return &Worker{
		service:  service,
		cron:     cron.New(),
		interval: interval,
	}
}

func (w *Worker) StartClienteleMetricsWorker() error {
	_, err := w.cron.AddFunc(w.interval, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		stats, err := w.service.CollectClienteleMetrics(ctx)
		if err != nil {
			log.Printf("❌ Failed to collect clientele metrics: %v", err)
			return
		}

		if err := w.service.SaveClienteleMetricsSnapshot(ctx, stats); err != nil {
			log.Printf("❌ Failed to save clientele metrics: %v", err)
			return
		}

		log.Printf("📊 Clientele metrics saved: accounts=%d, companies=%d, transactions=%d",
			stats.TotalAccounts, stats.TotalCompanies, stats.TotalTransactions)
	})

	if err != nil {
		return err
	}

	w.cron.Start()
	log.Printf("✅ Clientele metrics worker started with interval: %s", w.interval)
	return nil
}
