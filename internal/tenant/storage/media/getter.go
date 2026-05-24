package storagemedia

import (
	"kroncl-server/internal/config"

	"github.com/minio/minio-go/v7"
)

func (s *Service) GetClient() *minio.Client {
	return s.client
}

func (s *Service) GetConfig() config.MinIOConfig {
	return s.config
}
