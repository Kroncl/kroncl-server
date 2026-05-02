package admin

import (
	adminauth "kroncl-server/internal/admin/auth"
	adminhealth "kroncl-server/internal/admin/health"
	"kroncl-server/internal/auth"
	"kroncl-server/internal/config"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
)

type Deps struct {
	JWTService       *auth.JWTService
	AdminAuthService *adminauth.Service
}

func NewRoutes(deps Deps) chi.Router {
	r := chi.NewRouter()

	r.Use(deps.JWTService.RequireAuth)
	r.Use(httprate.LimitByIP(config.RATE_LIMIT_PRIVATE_ROUTES_PER_MINUTE, 1*time.Minute))

	r.Group(func(r chi.Router) {
		r.Use(deps.AdminAuthService.RequireAdmin)
		r.Get("/health", adminhealth.SendResult)
	})

	return r
}
