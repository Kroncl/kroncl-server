package coreworkers

import "time"

type MetricsMediaSnapshot struct {
	RecordedAt time.Time `json:"recorded_at"`

	// Общая статистика по всем бакетам
	TotalBuckets int   `json:"total_buckets"`
	TotalObjects int   `json:"total_objects"`
	TotalSizeMB  int64 `json:"total_size_mb"`

	// Статистика по публичному бакету
	PublicBucketObjects int   `json:"public_bucket_objects"`
	PublicBucketSizeMB  int64 `json:"public_bucket_size_mb"`

	// Статистика по временному бакету
	TempBucketObjects int   `json:"temp_bucket_objects"`
	TempBucketSizeMB  int64 `json:"temp_bucket_size_mb"`

	// Статистика по арендным бакетам
	TenantBucketsCount int     `json:"tenant_buckets_count"`
	TenantTotalObjects int     `json:"tenant_total_objects"`
	TenantTotalSizeMB  int64   `json:"tenant_total_size_mb"`
	AvgTenantObjects   float64 `json:"avg_tenant_objects"`
	AvgTenantSizeMB    float64 `json:"avg_tenant_size_mb"`

	// Дополнительная аналитика
	LargestBucketName    string `json:"largest_bucket_name"`
	LargestBucketObjects int    `json:"largest_bucket_objects"`
	LargestBucketSizeMB  int64  `json:"largest_bucket_size_mb"`
}
