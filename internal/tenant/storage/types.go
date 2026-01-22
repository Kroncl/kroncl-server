package storage

import (
	"time"
)

type Storage struct {
	ID          string                 `json:"id"`
	CompanyID   string                 `json:"company_id"`
	SchemaName  string                 `json:"schema_name"`
	Status      StorageStatus          `json:"status"`
	StorageType StorageType            `json:"storage_type"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type StorageSources struct {
	SchemaName            string  `json:"schema_name"`
	TotalSizeMB           float64 `json:"total_size_mb"`           // Общий размер в MB
	TotalSizePretty       string  `json:"total_size_pretty"`       // Человекочитаемый размер
	TableSizeMB           float64 `json:"table_size_mb"`           // Размер таблиц (без индексов)
	IndexSizeMB           float64 `json:"index_size_mb"`           // Размер индексов
	ToastSizeMB           float64 `json:"toast_size_mb"`           // Размер TOAST таблиц
	TableCount            int     `json:"table_count"`             // Количество таблиц
	IndexCount            int     `json:"index_count"`             // Количество индексов
	SequenceCount         int     `json:"sequence_count"`          // Количество последовательностей
	ViewCount             int     `json:"view_count"`              // Количество представлений
	MaterializedViewCount int     `json:"materialized_view_count"` // Количество материализованных представлений
	TotalRows             int64   `json:"total_rows"`              // Общее количество строк во всех таблицах
	DeadRows              int64   `json:"dead_rows"`               // Количество "мертвых" строк (нужен VACUUM)
	ActiveConnections     int     `json:"active_connections"`      // Активные соединения к этой схеме
	LastVacuum            *string `json:"last_vacuum"`             // Время последнего VACUUM
	LastAutovacuum        *string `json:"last_autovacuum"`         // Время последнего авто-VACUUM
	LastAnalyze           *string `json:"last_analyze"`            // Когда собрана статистика
	SchemaExists          bool    `json:"schema_exists"`           // Существует ли схема
	CreatedAt             *string `json:"created_at"`              // Когда создана схема
	UpdatedAt             *string `json:"updated_at"`              // Когда обновлена статистика
}

type StorageStatus string

const (
	StorageStatusNone         StorageStatus = "none"
	StorageStatusProvisioning StorageStatus = "provisioning"
	StorageStatusActive       StorageStatus = "active"
	StorageStatusFailed       StorageStatus = "failed"
	StorageStatusDeprecated   StorageStatus = "deprecated"
)

type StorageType string

const (
	StorageTypeSchema   StorageType = "schema"
	StorageTypeDatabase StorageType = "database"
)

type StorageStatusResponse struct {
	Storage      *Storage `json:"storage,omitempty"`       // Объект хранилища (может быть nil)
	Status       string   `json:"status"`                  // Текущий статус в виде строки
	Message      string   `json:"message"`                 // Человекочитаемое сообщение о статусе
	IsReady      bool     `json:"is_ready"`                // Готово ли хранилище к использованию
	SchemaName   string   `json:"schema_name,omitempty"`   // Имя схемы БД (если есть)
	SchemaExists bool     `json:"schema_exists,omitempty"` // Существует ли схема физически
}
