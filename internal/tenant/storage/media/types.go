package storagemedia

import "time"

type BucketInfo struct {
	Name         string    `json:"name"`
	CreationDate time.Time `json:"creation_date"`
	SizeMB       float64   `json:"size_mb"`
	ObjectCount  int       `json:"object_count"`
}

type BucketStatusResponse struct {
	IsReady    bool        `json:"is_ready"`
	Message    string      `json:"message"`
	BucketInfo *BucketInfo `json:"storage,omitempty"`
	Exists     bool        `json:"exists"`
}

type FileUploadRequest struct {
	FilePath string `json:"file_path"`
}

type FileResponse struct {
	URL string `json:"url"`
}

type BucketStats struct {
	BucketName   string    `json:"bucket_name"`
	SizeMB       float64   `json:"size_mb"`
	ObjectCount  int       `json:"object_count"`
	CreationDate time.Time `json:"creation_date"`
}
