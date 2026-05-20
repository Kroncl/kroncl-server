package storagemedia

import (
	"context"
	"fmt"
	"io"
	"kroncl-server/internal/config"
	"log"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Service struct {
	client *minio.Client
	config config.MinIOConfig
}

func NewService(cfg config.MinIOConfig) (*Service, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.RootUser, cfg.RootPassword, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	return &Service{
		client: client,
		config: cfg,
	}, nil
}

func (s *Service) InitTenantBucket(ctx context.Context, tenantID string) error {
	bucketName := fmt.Sprintf("tenant-%s", tenantID)
	go s.runProvisioningWorker(bucketName)
	return nil
}

func (s *Service) runProvisioningWorker(bucketName string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	log.Printf("Starting bucket provisioning for %s", bucketName)

	err := s.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	if err != nil {
		exists, errExists := s.client.BucketExists(ctx, bucketName)
		if errExists != nil {
			log.Printf("Failed to check bucket existence: %v", errExists)
			return
		}
		if exists {
			log.Printf("Bucket %s already exists", bucketName)
			return
		}
		log.Printf("Failed to create bucket: %v", err)
		return
	}

	policy := config.GetTenantBucketPolicy(bucketName)

	if err := s.client.SetBucketPolicy(ctx, bucketName, policy); err != nil {
		log.Printf("Failed to set private policy: %v", err)
		return
	}

	log.Printf("✅ Bucket %s created successfully", bucketName)
}

func (s *Service) DeleteTenantBucket(ctx context.Context, tenantID string) error {
	bucketName := fmt.Sprintf("tenant-%s", tenantID)
	return s.client.RemoveBucket(ctx, bucketName)
}

func (s *Service) GetBucketStatus(ctx context.Context, tenantID string) (*BucketStatusResponse, error) {
	bucketName := fmt.Sprintf("tenant-%s", tenantID)

	exists, err := s.client.BucketExists(ctx, bucketName)
	if err != nil {
		return &BucketStatusResponse{
			IsReady: false,
			Message: fmt.Sprintf("Failed to check bucket: %v", err),
			Exists:  false,
		}, nil
	}

	if !exists {
		return &BucketStatusResponse{
			IsReady: false,
			Message: "Bucket not created yet",
			Exists:  false,
		}, nil
	}

	info := &BucketInfo{
		Name: bucketName,
	}

	objCh := s.client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{})
	var totalSize int64
	var count int

	for obj := range objCh {
		if obj.Err != nil {
			continue
		}
		count++
		totalSize += obj.Size
		if info.CreationDate.IsZero() || obj.LastModified.Before(info.CreationDate) {
			info.CreationDate = obj.LastModified
		}
	}

	info.ObjectCount = count
	info.SizeMB = float64(totalSize) / (1024 * 1024)

	return &BucketStatusResponse{
		IsReady:    true,
		Message:    "Bucket is ready",
		BucketInfo: info,
		Exists:     true,
	}, nil
}

func (s *Service) GetBucketInfo(ctx context.Context, tenantID string) (*BucketInfo, error) {
	bucketName := fmt.Sprintf("tenant-%s", tenantID)

	exists, err := s.client.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("bucket not found")
	}

	info := &BucketInfo{
		Name: bucketName,
	}

	objCh := s.client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{})
	var totalSize int64
	var count int

	for obj := range objCh {
		if obj.Err != nil {
			continue
		}
		count++
		totalSize += obj.Size
		if info.CreationDate.IsZero() || obj.LastModified.Before(info.CreationDate) {
			info.CreationDate = obj.LastModified
		}
	}

	info.ObjectCount = count
	info.SizeMB = float64(totalSize) / (1024 * 1024)

	return info, nil
}

func (s *Service) UploadFileToBucket(ctx context.Context, objectPath string, reader io.Reader, size int64, contentType string) error {
	bucketName, ok := s.GetBucketFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant bucket not found in context")
	}

	_, err := s.client.PutObject(ctx, bucketName, objectPath, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	return nil
}

func (s *Service) GetFileFromBucket(ctx context.Context, objectPath string) (io.ReadCloser, error) {
	bucketName, ok := s.GetBucketFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant bucket not found in context")
	}

	obj, err := s.client.GetObject(ctx, bucketName, objectPath, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}
	return obj, nil
}

func (s *Service) DeleteFileFromBucket(ctx context.Context, objectPath string) error {
	bucketName, ok := s.GetBucketFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant bucket not found in context")
	}

	err := s.client.RemoveObject(ctx, bucketName, objectPath, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func (s *Service) GeneratePresignedURL(ctx context.Context, objectPath string, expiry time.Duration) (string, error) {
	bucketName, ok := s.GetBucketFromContext(ctx)
	if !ok {
		return "", fmt.Errorf("tenant bucket not found in context")
	}

	url, err := s.client.PresignedGetObject(ctx, bucketName, objectPath, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url.String(), nil
}
