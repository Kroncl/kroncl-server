package adminmedia

import (
	"kroncl-server/internal/core"
	"time"
)

type SystemMediaStats struct {
	TotalBuckets         int        `json:"total_buckets"`
	TotalObjects         int        `json:"total_objects"`
	TotalSizeMB          int64      `json:"total_size_mb"`
	PublicBucketObjects  int        `json:"public_bucket_objects"`
	PublicBucketSizeMB   int64      `json:"public_bucket_size_mb"`
	TempBucketObjects    int        `json:"temp_bucket_objects"`
	TempBucketSizeMB     int64      `json:"temp_bucket_size_mb"`
	TenantBucketsCount   int        `json:"tenant_buckets_count"`
	TenantTotalObjects   int        `json:"tenant_total_objects"`
	TenantTotalSizeMB    int64      `json:"tenant_total_size_mb"`
	AvgTenantObjects     float64    `json:"avg_tenant_objects"`
	AvgTenantSizeMB      float64    `json:"avg_tenant_size_mb"`
	LargestBucketName    string     `json:"largest_bucket_name"`
	LargestBucketObjects int        `json:"largest_bucket_objects"`
	LargestBucketSizeMB  int64      `json:"largest_bucket_size_mb"`
	LastSnapshotTime     *time.Time `json:"last_snapshot_time,omitempty"`
}

type MediaMetricsHistory struct {
	ID                  int64     `json:"id"`
	RecordedAt          time.Time `json:"recorded_at"`
	TotalBuckets        int       `json:"total_buckets"`
	TotalObjects        int       `json:"total_objects"`
	TotalSizeMB         int64     `json:"total_size_mb"`
	TenantBucketsCount  int       `json:"tenant_buckets_count"`
	TenantTotalObjects  int       `json:"tenant_total_objects"`
	TenantTotalSizeMB   int64     `json:"tenant_total_size_mb"`
	PublicBucketObjects int       `json:"public_bucket_objects"`
	PublicBucketSizeMB  int64     `json:"public_bucket_size_mb"`
	TempBucketObjects   int       `json:"temp_bucket_objects"`
	TempBucketSizeMB    int64     `json:"temp_bucket_size_mb"`
}

type BucketInfo struct {
	Name         string    `json:"name"`
	CreationDate time.Time `json:"creation_date"`
	ObjectsCount int       `json:"objects_count"`
	SizeMB       float64   `json:"size_mb"`
	IsPublic     bool      `json:"is_public"`
	IsTemp       bool      `json:"is_temp"`
	IsTenant     bool      `json:"is_tenant"`
	TenantID     *string   `json:"tenant_id,omitempty"`
}

type BucketsResponse struct {
	Buckets    []BucketInfo    `json:"buckets"`
	Pagination core.Pagination `json:"pagination"`
}
