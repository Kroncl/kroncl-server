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
	publicHost  string // localhost:9000 или domain.com
	useSSL      bool
}

type Config struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	UseSSL     bool
	Bucket     string
	PublicHost string
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
	contentType := header.Header.Get("Content-Type")
	if !AllowedImageTypes[contentType] {
		return nil, fmt.Errorf("unsupported file type: %s", contentType)
	}

	if header.Size > MaxFileSize {
		return nil, fmt.Errorf("file too large: %d bytes (max %d)", header.Size, MaxFileSize)
	}

	fileUUID := uuid.New().String()
	ext := strings.ToLower(filepath.Ext(header.Filename))
	filename := fmt.Sprintf("%s/%s%s", AvatarPath, fileUUID, ext)

	info, err := s.minioClient.PutObject(ctx, s.bucket, filename, file, header.Size, minio.PutObjectOptions{
		ContentType: contentType,
		UserMetadata: map[string]string{
			"original-name": header.Filename,
			"uploaded-by":   accountID,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload to minio: %w", err)
	}

	fileInfo, err := s.repo.CreateFile(ctx, CreateFileParams{
		ID:           fileUUID,
		Path:         filename,
		Size:         info.Size,
		MimeType:     contentType,
		CreatedBy:    accountID,
		OriginalName: &header.Filename,
	})
	if err != nil {
		_ = s.minioClient.RemoveObject(ctx, s.bucket, filename, minio.RemoveObjectOptions{})
		return nil, fmt.Errorf("failed to save file metadata: %w", err)
	}

	return fileInfo, nil
}

// GetFile — прямой URL без подписей
func (s *Service) GetFile(ctx context.Context, id string) (*File, error) {
	file, err := s.repo.GetFile(ctx, id)
	if err != nil {
		return nil, err
	}

	scheme := "http"
	if s.useSSL {
		scheme = "https"
	}

	// ⭐ Простой прямой URL: scheme://publicHost/bucket/path
	file.URL = fmt.Sprintf("%s://%s/%s/%s", scheme, s.publicHost, s.bucket, file.Path)
	return file, nil
}

// GetFileURL — только URL
func (s *Service) GetFileURL(ctx context.Context, id string) (string, error) {
	file, err := s.repo.GetFile(ctx, id)
	if err != nil {
		return "", err
	}

	scheme := "http"
	if s.useSSL {
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s/%s/%s", scheme, s.publicHost, s.bucket, file.Path), nil
}
