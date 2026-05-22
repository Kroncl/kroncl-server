package dm

import (
	"kroncl-server/internal/tenant/crm"
	"kroncl-server/internal/tenant/docs"
	"kroncl-server/internal/tenant/fm"
	"kroncl-server/internal/tenant/hrm"
	"kroncl-server/internal/tenant/pdfgen"
	storagemedia "kroncl-server/internal/tenant/storage/media"
	"kroncl-server/internal/tenant/wm"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool          *pgxpool.Pool
	fmRepository  *fm.Repository
	hrmRepository *hrm.Repository
	crmRepository *crm.Repository
	wmRepository  *wm.Repository
	pdfgen        *pdfgen.Service
	mediaService  *storagemedia.Service
	docsService   *docs.Service
}

func NewRepository(
	pool *pgxpool.Pool,
	fmRepository *fm.Repository,
	hrmRepository *hrm.Repository,
	crmRepository *crm.Repository,
	wmRepository *wm.Repository,
	pdfgen *pdfgen.Service,
	mediaService *storagemedia.Service,
	docsService *docs.Service,
) *Repository {
	return &Repository{
		pool:          pool,
		fmRepository:  fmRepository,
		crmRepository: crmRepository,
		wmRepository:  wmRepository,
		hrmRepository: hrmRepository,
		pdfgen:        pdfgen,
		mediaService:  mediaService,
		docsService:   docsService,
	}
}
