package media

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Service struct {
	minioClient *minio.Client
	repo        *Repository
	bucket      string
	publicHost  string
	useSSL      bool
}

type Config struct {
	Endpoint   string // для подключения к MinIO
	AccessKey  string
	SecretKey  string
	UseSSL     bool
	Bucket     string
	PublicHost string // для формирования публичных URL через Nginx
}

func NewService(cfg Config, repo *Repository) (*Service, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	return &Service{
		minioClient: client,
		repo:        repo,
		bucket:      cfg.Bucket,
		publicHost:  cfg.PublicHost,
		useSSL:      cfg.UseSSL,
	}, nil
}

func (s *Service) SaveFile(ctx context.Context, file multipart.File, header *multipart.FileHeader, accountID string) (*File, error) {
	// Проверка типа файла
	contentType := header.Header.Get("Content-Type")
	if !AllowedImageTypes[contentType] {
		return nil, fmt.Errorf("unsupported file type: %s", contentType)
	}

	// Проверка размера
	if header.Size > MaxFileSize {
		return nil, fmt.Errorf("file too large: %d bytes (max %d)", header.Size, MaxFileSize)
	}

	// Генерация уникального пути в MinIO
	ext := strings.ToLower(filepath.Ext(header.Filename))
	filename := fmt.Sprintf("%s/%s%s", AvatarPath, uuid.New().String(), ext)

	// Загрузка в MinIO
	info, err := s.minioClient.PutObject(ctx, s.bucket, filename, file, header.Size, minio.PutObjectOptions{
		ContentType: contentType,
		UserMetadata: map[string]string{
			"original-name": header.Filename,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload to minio: %w", err)
	}

	// Формируем публичный URL через Nginx
	scheme := "http"
	if s.useSSL {
		scheme = "https"
	}
	// URL теперь идет через Nginx, который проксирует /files/ в MinIO
	url := fmt.Sprintf("%s://%s/files/%s", scheme, s.publicHost, filename)

	// Сохраняем метаданные в БД
	fileInfo, err := s.repo.CreateFile(ctx, CreateFileParams{
		Path:         filename,
		URL:          url,
		Size:         info.Size,
		MimeType:     contentType,
		CreatedBy:    accountID,
		OriginalName: &header.Filename,
		Metadata:     nil,
	})
	if err != nil {
		// Если не удалось сохранить в БД — удаляем из MinIO
		_ = s.minioClient.RemoveObject(ctx, s.bucket, filename, minio.RemoveObjectOptions{})
		return nil, fmt.Errorf("failed to save file metadata: %w", err)
	}

	return fileInfo, nil
}

func (s *Service) GetFile(ctx context.Context, id string) (*File, error) {
	return s.repo.GetFile(ctx, id)
}
