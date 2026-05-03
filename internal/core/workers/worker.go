package coreworkers

import "github.com/robfig/cron/v3"

type Worker struct {
	service  *Service
	cron     *cron.Cron
	interval string
}
