package storagemedia

import (
	"encoding/json"
	"fmt"
	"io"
	"kroncl-server/internal/core"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Handlers struct {
	service *Service
}

func NewHandlers(service *Service) *Handlers {
	return &Handlers{service: service}
}

func (h *Handlers) UploadFile(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		core.SendError(w, http.StatusBadRequest, "Failed to parse form data")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		core.SendError(w, http.StatusBadRequest, "File is required")
		return
	}
	defer file.Close()

	tag := strings.TrimSpace(r.FormValue("tag"))

	if !allowedTags[tag] {
		core.SendError(w, http.StatusBadRequest, "Invalid tag. Allowed: invoices, avatars, attachments, reports")
		return
	}

	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".bin"
	}

	var objectPath string
	if tag == "" {
		objectPath = fmt.Sprintf("%s%s", uuid.New().String(), ext)
	} else {
		objectPath = fmt.Sprintf("%s/%s%s", tag, uuid.New().String(), ext)
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	err = h.service.UploadFileToBucket(r.Context(), objectPath, file, header.Size, contentType)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to upload file: %s", err.Error()))
		return
	}

	previewURL, _ := h.service.GeneratePresignedURL(r.Context(), objectPath, 15*time.Minute)

	core.SendSuccess(w, map[string]interface{}{
		"path":        objectPath,
		"size":        header.Size,
		"preview_url": previewURL,
	}, "File uploaded successfully")
}

func (h *Handlers) GetFile(w http.ResponseWriter, r *http.Request) {
	objectPath := r.URL.Query().Get("path")
	if objectPath == "" {
		core.SendError(w, http.StatusBadRequest, "Path is required")
		return
	}

	reader, err := h.service.GetFileFromBucket(r.Context(), objectPath)
	if err != nil {
		core.SendNotFound(w, fmt.Sprintf("File not found: %s", err.Error()))
		return
	}
	defer reader.Close()

	contentType := r.URL.Query().Get("content_type")
	if contentType == "" {
		ext := filepath.Ext(objectPath)
		switch ext {
		case ".pdf":
			contentType = "application/pdf"
		case ".jpg", ".jpeg":
			contentType = "image/jpeg"
		case ".png":
			contentType = "image/png"
		default:
			contentType = "application/octet-stream"
		}
	}
	w.Header().Set("Content-Type", contentType)

	_, err = io.Copy(w, reader)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to write file: %s", err.Error()))
		return
	}
}

func (h *Handlers) DeleteFile(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		core.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Path == "" {
		core.SendError(w, http.StatusBadRequest, "Path is required")
		return
	}

	err := h.service.DeleteFileFromBucket(r.Context(), req.Path)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to delete file: %s", err.Error()))
		return
	}

	core.SendSuccess(w, nil, "File deleted successfully")
}

func (h *Handlers) GetBucketStats(w http.ResponseWriter, r *http.Request) {
	companyID, ok := core.GetCompanyIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusBadRequest, "Company context not found")
		return
	}

	info, err := h.service.GetBucketInfo(r.Context(), companyID)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to get bucket stats: %s", err.Error()))
		return
	}

	core.SendSuccess(w, info, "Bucket stats retrieved successfully")
}

func (h *Handlers) GeneratePresignedURL(w http.ResponseWriter, r *http.Request) {
	objectPath := r.URL.Query().Get("path")
	if objectPath == "" {
		core.SendError(w, http.StatusBadRequest, "Path is required")
		return
	}

	expiry := 15 * time.Minute
	if expiryStr := r.URL.Query().Get("expiry"); expiryStr != "" {
		if d, err := time.ParseDuration(expiryStr); err == nil && d > 0 {
			expiry = d
		}
	}

	url, err := h.service.GeneratePresignedURL(r.Context(), objectPath, expiry)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to generate URL: %s", err.Error()))
		return
	}

	core.SendSuccess(w, url, "URL generated successfully")
}
