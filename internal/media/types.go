package media

import (
	"time"
)

type File struct {
	ID           string    `json:"id"`
	Path         string    `json:"path"`
	URL          string    `json:"url"`
	Size         int64     `json:"size"`
	MimeType     string    `json:"mime_type"`
	CreatedAt    time.Time `json:"created_at"`
	CreatedBy    string    `json:"created_by"`
	OriginalName *string   `json:"original_name,omitempty"`
	Metadata     []byte    `json:"metadata,omitempty"`
	EntityType   *string   `json:"entity_type,omitempty"`
	EntityID     *string   `json:"entity_id,omitempty"`
}

type CreateFileParams struct {
	Path         string
	URL          string
	Size         int64
	MimeType     string
	CreatedBy    string
	OriginalName *string
	Metadata     []byte
	EntityType   *string
	EntityID     *string
}

type UploadResponse struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

const (
	MaxFileSize  = 10 * 1024 * 1024 // 10MB
	PublicBucket = "public"
	AvatarPath   = "avatars"
)

var AllowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}
