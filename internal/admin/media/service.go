package adminmedia

import (
	coreworkers "kroncl-server/internal/core/workers"
	storagemedia "kroncl-server/internal/tenant/storage/media"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool                *pgxpool.Pool
	metricsService      *coreworkers.Service
	storageMediaService *storagemedia.Service
}

func NewService(
	pool *pgxpool.Pool,
	metricsService *coreworkers.Service,
	storageMediaService *storagemedia.Service,
) *Service {
	return &Service{
		pool:                pool,
		metricsService:      metricsService,
		storageMediaService: storageMediaService,
	}
}
