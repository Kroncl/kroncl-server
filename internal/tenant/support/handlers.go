package support

import "kroncl-server/internal/tenant/logs"

type Handlers struct {
	service     *Service
	logsService *logs.Service
}

func NewHandlers(service *Service, logsService *logs.Service) *Handlers {
	return &Handlers{
		service:     service,
		logsService: logsService,
	}
}
