package dm

import (
	"kroncl-server/internal/tenant/crm"
	"kroncl-server/internal/tenant/fm"
	"kroncl-server/internal/tenant/hrm"
	"kroncl-server/internal/tenant/wm"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool          *pgxpool.Pool
	fmRepository  *fm.Repository
	hrmRepository *hrm.Repository
	crmRepository *crm.Repository
	wmRepository  *wm.Repository
}

func NewRepository(
	pool *pgxpool.Pool,
	fmRepository *fm.Repository,
	hrmRepository *hrm.Repository,
	crmRepository *crm.Repository,
	wmRepository *wm.Repository,
) *Repository {
	return &Repository{
		pool:          pool,
		fmRepository:  fmRepository,
		crmRepository: crmRepository,
		wmRepository:  wmRepository,
		hrmRepository: hrmRepository,
	}
}
