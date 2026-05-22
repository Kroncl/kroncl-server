package hrm

import (
	"kroncl-server/internal/accounts"
	"kroncl-server/internal/companies"
	"kroncl-server/internal/tenant/docs"
	"kroncl-server/internal/tenant/excelizer"
	storagemedia "kroncl-server/internal/tenant/storage/media"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool             *pgxpool.Pool
	accountsService  *accounts.Service
	companiesService *companies.Service
	mediaService     *storagemedia.Service
	excelizer        *excelizer.Service
	docsService      *docs.Service
}

func NewRepository(
	pool *pgxpool.Pool,
	accountsService *accounts.Service,
	companiesService *companies.Service,
	mediaService *storagemedia.Service,
	excelizer *excelizer.Service,
	docsService *docs.Service,
) *Repository {
	return &Repository{
		pool:             pool,
		accountsService:  accountsService,
		companiesService: companiesService,
		mediaService:     mediaService,
		excelizer:        excelizer,
		docsService:      docsService,
	}
}
