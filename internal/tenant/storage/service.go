package storage

import (
	"context"
	"fmt"
	"log"
	"time"

	"kroncl-server/internal/tenant/migrator"
)

type Service struct {
	repository *Repository
	migrator   *migrator.Migrator
}

func NewService(repository *Repository, migrator *migrator.Migrator) *Service {
	return &Service{
		repository: repository,
		migrator:   migrator,
	}
}

func (s *Service) InitStorage(ctx context.Context, companyID string) (*Storage, error) {
	// создаём запись хранилища
	storage, err := s.repository.CreateStorageRecord(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed init storage: %w", err)
	}

	// запускаем воркер миграций в фоне
	go s.runProvisioningWorker(storage.ID, storage.SchemaName)

	return storage, nil
}

// runProvisioningWorker фоновый воркер
func (s *Service) runProvisioningWorker(storageID, schemaName string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	log.Printf("Starting provisioning for storage %s, schema: %s", storageID, schemaName)

	// 1. Обновляем статус на 'provisioning' (хотя уже такой, но для ясности)
	s.repository.UpdateStorageStatus(ctx, storageID, string(StorageStatusProvisioning))

	// 2. Создаем схему
	if err := s.migrator.CreateSchema(ctx, schemaName); err != nil {
		log.Printf("Failed to create schema: %v", err)
		s.repository.UpdateStorageStatus(ctx, storageID, string(StorageStatusFailed))
		return
	}

	// 3. Применяем миграции
	if err := s.migrator.Up(ctx, schemaName); err != nil {
		log.Printf("Failed to apply migrations: %v", err)
		s.repository.UpdateStorageStatus(ctx, storageID, string(StorageStatusFailed))
		return
	}

	// 4. Обновляем статус на 'active'
	if err := s.repository.UpdateStorageStatus(ctx, storageID, string(StorageStatusActive)); err != nil {
		log.Printf("Failed to update status to active: %v", err)
		return
	}

	log.Printf("Provisioning completed for storage %s", storageID)
}

// GetStorageStatus возвращает статус хранилища
func (s *Service) GetStorageStatus(ctx context.Context, companyID string) (*StorageStatusResponse, error) {
	// Нужно добавить метод в репозиторий
	storage, err := s.repository.GetStorageByCompanyID(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage: %w", err)
	}

	if storage == nil {
		return &StorageStatusResponse{
			Status:  "not_created",
			Message: "Storage not initialized",
			IsReady: false,
		}, nil
	}

	// Проверяем существует ли схема (опционально)
	schemaExists := false
	if storage.Status == StorageStatusActive {
		var err error
		schemaExists, err = s.migrator.CheckSchemaExists(ctx, storage.SchemaName)
		if err != nil {
			log.Printf("Failed to check schema existence: %v", err)
		}
	}

	return &StorageStatusResponse{
		Storage:      storage,
		Status:       string(storage.Status),
		Message:      s.getStatusMessage(storage.Status),
		IsReady:      storage.Status == StorageStatusActive && schemaExists,
		SchemaName:   storage.SchemaName,
		SchemaExists: schemaExists,
	}, nil
}

func (s *Service) getStatusMessage(status StorageStatus) string {
	switch status {
	case StorageStatusProvisioning:
		return "Creating schema and applying migrations..."
	case StorageStatusActive:
		return "Storage is ready"
	case StorageStatusFailed:
		return "Storage creation failed"
	default:
		return string(status)
	}
}
