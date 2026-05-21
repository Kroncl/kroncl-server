package fm

import (
	"kroncl-server/internal/tenant/docs"
	"kroncl-server/internal/tenant/excelizer"
	"kroncl-server/internal/tenant/hrm"
	storagemedia "kroncl-server/internal/tenant/storage/media"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool                *pgxpool.Pool
	employeesRepository *hrm.Repository
	mediaService        *storagemedia.Service
	excelizer           *excelizer.Service
	docsService         *docs.Service
}

func NewRepository(pool *pgxpool.Pool, employeesRepository *hrm.Repository, mediaService *storagemedia.Service, excelizer *excelizer.Service, docsService *docs.Service) *Repository {
	return &Repository{pool: pool, employeesRepository: employeesRepository, mediaService: mediaService, excelizer: excelizer, docsService: docsService}
}
