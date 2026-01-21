package migrator

import (
	"kroncl-server/utils"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SchemaType string

const (
	SchemaTypePublic SchemaType = "public"
	SchemaTypeTenant SchemaType = "tenant"
)

type Migrator struct {
	pool     *pgxpool.Pool
	basePath string
	config   utils.DBConfig
}

type Config struct {
	MigrationsPath string
	SchemaType     SchemaType
}
