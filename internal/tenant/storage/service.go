package storage

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"kroncl-server/internal/core"
	"kroncl-server/internal/migrator"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	repository  *Repository
	migrator    *migrator.Migrator
	globalPool  *pgxpool.Pool
	tenantPools sync.Map
}

func NewService(repository *Repository, migrator *migrator.Migrator, globalPool *pgxpool.Pool) *Service {
	return &Service{
		repository: repository,
		migrator:   migrator,
		globalPool: globalPool,
	}
}

func (s *Service) GetTenantPool(ctx context.Context, companyID string) (*pgxpool.Pool, error) {
	// Пытаемся получить из кэша
	if pool, ok := s.tenantPools.Load(companyID); ok {
		return pool.(*pgxpool.Pool), nil
	}

	// Создаём новый пул
	storage, err := s.repository.GetStorageByCompanyID(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage: %w", err)
	}

	dsn := s.buildTenantDSN(storage.SchemaName)
	tenantPool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	// Сохраняем в кэш
	s.tenantPools.Store(companyID, tenantPool)

	// Очистка при закрытии приложения
	go func() {
		<-ctx.Done()
		tenantPool.Close()
		s.tenantPools.Delete(companyID)
	}()

	return tenantPool, nil
}

// buildTenantDSN строит DSN для подключения к схеме тенанта
func (s *Service) buildTenantDSN(schemaName string) string {
	// Получаем конфиг из глобального пула
	config := s.globalPool.Config()

	// Создаем DSN на основе конфига
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		config.ConnConfig.User,
		config.ConnConfig.Password,
		config.ConnConfig.Host,
		config.ConnConfig.Port,
		config.ConnConfig.Database,
	)

	// Добавляем параметры
	dsn += "?sslmode=disable"

	// Добавляем search_path
	dsn += fmt.Sprintf("&search_path=%s", schemaName)

	return dsn
}

// GetTenantPoolFromContext получает пул тенанта из контекста (из companyID)
func (s *Service) GetTenantPoolFromContext(ctx context.Context) (*pgxpool.Pool, error) {
	companyID, ok := core.GetCompanyIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("company ID not found in context")
	}

	return s.GetTenantPool(ctx, companyID)
}

// TenantPoolMiddleware - переиспользуем существующий пул
func (s *Service) TenantPoolMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		companyID, ok := core.GetCompanyIDFromContext(r.Context())
		if !ok {
			core.SendError(w, http.StatusBadRequest, "Company context not found")
			return
		}

		// Получаем или создаём пул
		tenantPool, err := s.GetTenantPool(r.Context(), companyID)
		if err != nil {
			core.SendError(w, http.StatusInternalServerError, "Failed to get company storage")
			return
		}

		// Проверяем, что пул жив
		if err := tenantPool.Ping(r.Context()); err != nil {
			// Удаляем битый пул из кэша
			s.tenantPools.Delete(companyID)
			core.SendError(w, http.StatusInternalServerError, "Storage connection lost")
			return
		}

		// Добавляем в контекст
		ctx := context.WithValue(r.Context(), "tenant_pool", tenantPool)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetTenantPoolFromRequest извлекает пул тенанта из контекста запроса
func (s *Service) GetTenantPoolFromRequest(r *http.Request) (*pgxpool.Pool, bool) {
	pool, ok := r.Context().Value("tenant_pool").(*pgxpool.Pool)
	return pool, ok
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

func (s *Service) runProvisioningWorker(storageID, schemaName string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	log.Printf("Starting provisioning for storage %s, schema: %s", storageID, schemaName)

	// 1. Обновляем статус
	s.repository.UpdateStorageStatus(ctx, storageID, string(StorageStatusProvisioning))

	// 2. Создаем схему
	if err := s.migrator.CreateSchema(ctx, schemaName); err != nil {
		log.Printf("Failed to create schema: %v", err)
		s.repository.UpdateStorageStatus(ctx, storageID, string(StorageStatusFailed))
		return
	}

	// 3. Применяем миграции тенантов
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

	log.Printf("✅ Provisioning completed for storage %s", storageID)
}

func (s *Service) GetStorageStatus(ctx context.Context, companyID string) (*StorageStatusResponse, error) {
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
