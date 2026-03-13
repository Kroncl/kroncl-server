package media

import (
	"log"
	"net/http"
	"strings"

	"kroncl-server/internal/core"
)

type Handlers struct {
	service *Service
}

func NewHandlers(service *Service) *Handlers {
	return &Handlers{
		service: service,
	}
}

func (h *Handlers) UploadFile(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	err := r.ParseMultipartForm(MaxFileSize)
	if err != nil {
		log.Printf("Error parsing form: %v", err)
		core.SendError(w, http.StatusBadRequest, "File too large or invalid form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		core.SendError(w, http.StatusBadRequest, "No file provided")
		return
	}
	defer file.Close()

	fileInfo, err := h.service.SaveFile(r.Context(), file, header, accountID)
	if err != nil {
		status := http.StatusInternalServerError
		msg := err.Error()

		if strings.Contains(msg, "unsupported file type") ||
			strings.Contains(msg, "file too large") {
			status = http.StatusBadRequest
		}

		core.SendError(w, status, msg)
		return
	}

	url, err := h.service.GetFileURL(r.Context(), fileInfo.ID)
	if err != nil {
		core.SendInternalError(w, "File uploaded but failed to generate access URL")
		return
	}

	core.SendSuccess(w, UploadResponse{
		ID:  fileInfo.ID,
		URL: url,
	}, "File uploaded successfully")
}

func (h *Handlers) GetFile(w http.ResponseWriter, r *http.Request) {
	fileID := r.PathValue("fileId")
	if fileID == "" {
		core.SendError(w, http.StatusBadRequest, "File ID is required")
		return
	}

	file, err := h.service.GetFile(r.Context(), fileID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			core.SendNotFound(w, "File not found")
			return
		}
		core.SendInternalError(w, "Failed to get file")
		return
	}

	core.SendSuccess(w, file, "File retrieved successfully")
}

func (h *Handlers) GetFileURL(w http.ResponseWriter, r *http.Request) {
	fileID := r.PathValue("fileId")
	if fileID == "" {
		core.SendError(w, http.StatusBadRequest, "File ID is required")
		return
	}

	url, err := h.service.GetFileURL(r.Context(), fileID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			core.SendNotFound(w, "File not found")
			return
		}
		core.SendInternalError(w, "Failed to generate URL")
		return
	}

	core.SendSuccess(w, map[string]string{
		"url": url,
	}, "File URL generated successfully")
}
