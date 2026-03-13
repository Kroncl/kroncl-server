package wm

import "kroncl-server/internal/tenant/logs"

type Handlers struct {
	repository  *Repository
	logsService *logs.Service
}

func NewHandlers(repository *Repository, logsService *logs.Service) *Handlers {
	return &Handlers{
		repository:  repository,
		logsService: logsService,
	}
}
