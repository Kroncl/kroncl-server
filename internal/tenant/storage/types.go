package storage

import "time"

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
