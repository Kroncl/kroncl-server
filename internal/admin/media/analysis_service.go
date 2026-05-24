package adminmedia

import (
	"context"
	"fmt"
	"strings"
	"time"

	"kroncl-server/internal/core"

	"github.com/minio/minio-go/v7"
)

func (s *Service) GetSystemStats(ctx context.Context) (*SystemMediaStats, error) {
	stats := &SystemMediaStats{}

	client := s.storageMediaService.GetClient()
	config := s.storageMediaService.GetConfig()

	buckets, err := client.ListBuckets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}

	stats.TotalBuckets = len(buckets)

	var totalObjects int
	var totalSize int64
	var tenantBucketsCount int
	var tenantTotalObjects int
	var tenantTotalSize int64
	var largestBucketName string
	var largestBucketObjects int
	var largestBucketSize int64

	for _, bucket := range buckets {
		bucketName := bucket.Name

		objCh := client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
			Recursive: true,
		})

		var bucketObjects int
		var bucketSize int64

		for obj := range objCh {
			if obj.Err != nil {
				continue
			}
			bucketObjects++
			bucketSize += obj.Size
		}

		totalObjects += bucketObjects
		totalSize += bucketSize

		if bucketName == config.PublicBucket {
			stats.PublicBucketObjects = bucketObjects
			stats.PublicBucketSizeMB = bucketSize / (1024 * 1024)
			continue
		}

		if bucketName == "temp" {
			stats.TempBucketObjects = bucketObjects
			stats.TempBucketSizeMB = bucketSize / (1024 * 1024)
			continue
		}

		if strings.HasPrefix(bucketName, "tenant-") {
			tenantBucketsCount++
			tenantTotalObjects += bucketObjects
			tenantTotalSize += bucketSize
		}

		if bucketSize > largestBucketSize {
			largestBucketSize = bucketSize
			largestBucketObjects = bucketObjects
			largestBucketName = bucketName
		}
	}

	stats.TotalObjects = totalObjects
	stats.TotalSizeMB = totalSize / (1024 * 1024)
	stats.TenantBucketsCount = tenantBucketsCount
	stats.TenantTotalObjects = tenantTotalObjects
	stats.TenantTotalSizeMB = tenantTotalSize / (1024 * 1024)

	if tenantBucketsCount > 0 {
		stats.AvgTenantObjects = float64(tenantTotalObjects) / float64(tenantBucketsCount)
		stats.AvgTenantSizeMB = float64(tenantTotalSize) / float64(tenantBucketsCount) / (1024 * 1024)
	}

	stats.LargestBucketName = largestBucketName
	stats.LargestBucketObjects = largestBucketObjects
	stats.LargestBucketSizeMB = largestBucketSize / (1024 * 1024)

	var lastSnapshotTime *time.Time
	query := `SELECT MAX(recorded_at) FROM metrics_media_history`
	err = s.pool.QueryRow(ctx, query).Scan(&lastSnapshotTime)
	if err == nil {
		stats.LastSnapshotTime = lastSnapshotTime
	}

	return stats, nil
}

func (s *Service) GetMetricsHistory(ctx context.Context, startDate, endDate *time.Time, limit int) ([]MediaMetricsHistory, error) {
	query := `
        SELECT 
            id, recorded_at, total_buckets, total_objects, total_size_mb,
            tenant_buckets_count, tenant_total_objects, tenant_total_size_mb,
            public_bucket_objects, public_bucket_size_mb,
            temp_bucket_objects, temp_bucket_size_mb
        FROM metrics_media_history
        WHERE 1=1
    `

	args := []interface{}{}
	argCounter := 1

	if startDate != nil {
		query += fmt.Sprintf(" AND recorded_at >= $%d", argCounter)
		args = append(args, *startDate)
		argCounter++
	}

	if endDate != nil {
		query += fmt.Sprintf(" AND recorded_at <= $%d", argCounter)
		args = append(args, *endDate)
		argCounter++
	}

	query += " ORDER BY recorded_at DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argCounter)
		args = append(args, limit)
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics history: %w", err)
	}
	defer rows.Close()

	var metrics []MediaMetricsHistory
	for rows.Next() {
		var m MediaMetricsHistory
		err := rows.Scan(
			&m.ID,
			&m.RecordedAt,
			&m.TotalBuckets,
			&m.TotalObjects,
			&m.TotalSizeMB,
			&m.TenantBucketsCount,
			&m.TenantTotalObjects,
			&m.TenantTotalSizeMB,
			&m.PublicBucketObjects,
			&m.PublicBucketSizeMB,
			&m.TempBucketObjects,
			&m.TempBucketSizeMB,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metric: %w", err)
		}
		metrics = append(metrics, m)
	}

	return metrics, nil
}

func (s *Service) GetBuckets(ctx context.Context, search string, params core.PaginationParams) (*BucketsResponse, error) {
	client := s.storageMediaService.GetClient()
	config := s.storageMediaService.GetConfig()

	allBuckets, err := client.ListBuckets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}

	var buckets []BucketInfo
	for _, bucket := range allBuckets {
		bucketName := bucket.Name

		if search != "" && !strings.Contains(strings.ToLower(bucketName), strings.ToLower(search)) {
			continue
		}

		objCh := client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
			Recursive: true,
		})

		var objectsCount int
		var totalSize int64
		var creationDate time.Time

		for obj := range objCh {
			if obj.Err != nil {
				continue
			}
			objectsCount++
			totalSize += obj.Size
			if creationDate.IsZero() || obj.LastModified.Before(creationDate) {
				creationDate = obj.LastModified
			}
		}

		isPublic := bucketName == config.PublicBucket
		isTemp := bucketName == "temp"
		isTenant := strings.HasPrefix(bucketName, "tenant-")
		var tenantID *string
		if isTenant {
			id := strings.TrimPrefix(bucketName, "tenant-")
			tenantID = &id
		}

		buckets = append(buckets, BucketInfo{
			Name:         bucketName,
			CreationDate: creationDate,
			ObjectsCount: objectsCount,
			SizeMB:       float64(totalSize) / (1024 * 1024),
			IsPublic:     isPublic,
			IsTemp:       isTemp,
			IsTenant:     isTenant,
			TenantID:     tenantID,
		})
	}

	total := len(buckets)
	start := params.Offset
	end := start + params.Limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paginatedBuckets := buckets[start:end]
	pagination := core.NewPagination(total, params.Page, params.Limit)

	return &BucketsResponse{
		Buckets:    paginatedBuckets,
		Pagination: pagination,
	}, nil
}

func (s *Service) GetBucket(ctx context.Context, bucketName string) (*BucketInfo, error) {
	client := s.storageMediaService.GetClient()
	config := s.storageMediaService.GetConfig()

	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("bucket not found")
	}

	objCh := client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Recursive: true,
	})

	var objectsCount int
	var totalSize int64
	var creationDate time.Time

	for obj := range objCh {
		if obj.Err != nil {
			continue
		}
		objectsCount++
		totalSize += obj.Size
		if creationDate.IsZero() || obj.LastModified.Before(creationDate) {
			creationDate = obj.LastModified
		}
	}

	isPublic := bucketName == config.PublicBucket
	isTemp := bucketName == "temp"
	isTenant := strings.HasPrefix(bucketName, "tenant-")
	var tenantID *string
	if isTenant {
		id := strings.TrimPrefix(bucketName, "tenant-")
		tenantID = &id
	}

	return &BucketInfo{
		Name:         bucketName,
		CreationDate: creationDate,
		ObjectsCount: objectsCount,
		SizeMB:       float64(totalSize) / (1024 * 1024),
		IsPublic:     isPublic,
		IsTemp:       isTemp,
		IsTenant:     isTenant,
		TenantID:     tenantID,
	}, nil
}
