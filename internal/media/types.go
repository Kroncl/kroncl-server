package media

import (
	"time"
)

type File struct {
	ID           string    `json:"id"`
	Path         string    `json:"path"`
	URL          string    `json:"url"` // генерируется на лету
	Size         int64     `json:"size"`
	MimeType     string    `json:"mime_type"`
	CreatedAt    time.Time `json:"created_at"`
	CreatedBy    string    `json:"created_by"`
	OriginalName *string   `json:"original_name,omitempty"`
}

type CreateFileParams struct {
	ID           string
	Path         string
	Size         int64
	MimeType     string
	CreatedBy    string
	OriginalName *string
}

type UploadResponse struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

const (
	PublicBucket = "public"
	MaxFileSize  = 10 * 1024 * 1024 // 10MB
	Bucket       = PublicBucket
	AvatarPath   = "avatars"
)

var AllowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}
