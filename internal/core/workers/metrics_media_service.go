package coreworkers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
)

func (s *Service) CollectMediaMetrics(ctx context.Context) (*MetricsMediaSnapshot, error) {
	stats := &MetricsMediaSnapshot{
		RecordedAt: time.Now(),
	}

	// Получаем клиент через геттер
	client := s.storageMediaService.GetClient()

	// Получаем список всех бакетов
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

	config := s.storageMediaService.GetConfig()

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

	return stats, nil
}

func (s *Service) SaveMediaMetricsSnapshot(ctx context.Context, stats *MetricsMediaSnapshot) error {
	query := `
        INSERT INTO metrics_media_history (
            recorded_at, total_buckets, total_objects, total_size_mb,
            public_bucket_objects, public_bucket_size_mb,
            temp_bucket_objects, temp_bucket_size_mb,
            tenant_buckets_count, tenant_total_objects, tenant_total_size_mb,
            avg_tenant_objects, avg_tenant_size_mb,
            largest_bucket_name, largest_bucket_objects, largest_bucket_size_mb
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
        )
    `

	_, err := s.pool.Exec(ctx, query,
		stats.RecordedAt,
		stats.TotalBuckets,
		stats.TotalObjects,
		stats.TotalSizeMB,
		stats.PublicBucketObjects,
		stats.PublicBucketSizeMB,
		stats.TempBucketObjects,
		stats.TempBucketSizeMB,
		stats.TenantBucketsCount,
		stats.TenantTotalObjects,
		stats.TenantTotalSizeMB,
		stats.AvgTenantObjects,
		stats.AvgTenantSizeMB,
		stats.LargestBucketName,
		stats.LargestBucketObjects,
		stats.LargestBucketSizeMB,
	)

	if err != nil {
		return fmt.Errorf("failed to save media metrics snapshot: %w", err)
	}

	return nil
}

func (s *Service) GetMediaMetricsHistory(ctx context.Context, startDate, endDate *time.Time, limit int) ([]MetricsMediaSnapshot, error) {
	query := `
        SELECT 
            recorded_at, total_buckets, total_objects, total_size_mb,
            public_bucket_objects, public_bucket_size_mb,
            temp_bucket_objects, temp_bucket_size_mb,
            tenant_buckets_count, tenant_total_objects, tenant_total_size_mb,
            avg_tenant_objects, avg_tenant_size_mb,
            largest_bucket_name, largest_bucket_objects, largest_bucket_size_mb
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
		return nil, fmt.Errorf("failed to get media metrics history: %w", err)
	}
	defer rows.Close()

	var metrics []MetricsMediaSnapshot
	for rows.Next() {
		var m MetricsMediaSnapshot
		err := rows.Scan(
			&m.RecordedAt,
			&m.TotalBuckets,
			&m.TotalObjects,
			&m.TotalSizeMB,
			&m.PublicBucketObjects,
			&m.PublicBucketSizeMB,
			&m.TempBucketObjects,
			&m.TempBucketSizeMB,
			&m.TenantBucketsCount,
			&m.TenantTotalObjects,
			&m.TenantTotalSizeMB,
			&m.AvgTenantObjects,
			&m.AvgTenantSizeMB,
			&m.LargestBucketName,
			&m.LargestBucketObjects,
			&m.LargestBucketSizeMB,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan media metric: %w", err)
		}
		metrics = append(metrics, m)
	}

	return metrics, nil
}
