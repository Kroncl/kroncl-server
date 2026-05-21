package storage

import (
	storagedb "kroncl-server/internal/tenant/storage/db"
	storagemedia "kroncl-server/internal/tenant/storage/media"
)

type Service struct {
	Db    *storagedb.Service
	Media *storagemedia.Service
}

func NewService(
	dbService *storagedb.Service,
	mediaService *storagemedia.Service,
) *Service {
	return &Service{
		Db:    dbService,
		Media: mediaService,
	}
}
