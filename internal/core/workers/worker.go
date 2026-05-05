package coreworkers

import (
	"log"

	"github.com/robfig/cron/v3"
)

type Worker struct {
	service  *Service
	cron     *cron.Cron
	interval string
}

func (w *Worker) Stop() {
	<-w.cron.Stop().Done()
	log.Println("⏹️ Metrics worker stopped")
}
