package excelizer

import (
	"bytes"
	"context"
	"fmt"
	storagemedia "kroncl-server/internal/tenant/storage/media"
	"time"

	"github.com/xuri/excelize/v2"
)

type SheetGenerator func(ctx context.Context, f *excelize.File, sheetName string) (int, error)

type Service struct {
	mediaService storagemedia.MediaUploader
}

func NewService(mediaService storagemedia.MediaUploader) *Service {
	return &Service{mediaService: mediaService}
}

type ReportResult struct {
	ObjectPath   string
	TotalRows    int
	PresignedURL string
}

func (s *Service) GenerateSingleSheetReport(ctx context.Context, generator SheetGenerator, filePrefix string, expiry time.Duration) (*ReportResult, error) {
	f := excelize.NewFile()
	defer f.Close()

	f.DeleteSheet("Sheet1")
	_, err := f.NewSheet("Sheet1")
	if err != nil {
		return nil, fmt.Errorf("failed to create sheet: %w", err)
	}

	total, err := generator(ctx, f, "Sheet1")
	if err != nil {
		return nil, fmt.Errorf("failed to generate sheet: %w", err)
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write excel: %w", err)
	}

	objectPath := fmt.Sprintf("%s%s.xlsx", filePrefix, time.Now().Format("20060102_150405"))

	err = s.mediaService.UploadFileToBucket(ctx, objectPath, &buf, int64(buf.Len()), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	if err != nil {
		return nil, fmt.Errorf("failed to upload report: %w", err)
	}

	presignedURL, err := s.mediaService.GeneratePresignedURL(ctx, objectPath, expiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return &ReportResult{
		ObjectPath:   objectPath,
		TotalRows:    total,
		PresignedURL: presignedURL,
	}, nil
}

func (s *Service) GenerateMultiSheetReport(ctx context.Context, generators map[string]SheetGenerator, filePrefix string, expiry time.Duration) (*ReportResult, error) {
	f := excelize.NewFile()
	defer f.Close()

	var totalRows int
	first := true

	for sheetName, generator := range generators {
		if first {
			// Переименовываем дефолтный лист в первый
			if err := f.SetSheetName("Sheet1", sheetName); err != nil {
				return nil, fmt.Errorf("failed to rename sheet: %w", err)
			}
			first = false
		} else {
			// Создаём остальные листы
			_, err := f.NewSheet(sheetName)
			if err != nil {
				return nil, fmt.Errorf("failed to create sheet %s: %w", sheetName, err)
			}
		}

		total, err := generator(ctx, f, sheetName)
		if err != nil {
			return nil, fmt.Errorf("failed to generate sheet %s: %w", sheetName, err)
		}
		totalRows += total
	}

	if len(generators) == 0 {
		_, err := f.NewSheet("Sheet1")
		if err != nil {
			return nil, fmt.Errorf("failed to create default sheet: %w", err)
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write excel: %w", err)
	}

	objectPath := fmt.Sprintf("%s%s.xlsx", filePrefix, time.Now().Format("20060102_150405"))

	err := s.mediaService.UploadFileToBucket(ctx, objectPath, &buf, int64(buf.Len()), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	if err != nil {
		return nil, fmt.Errorf("failed to upload report: %w", err)
	}

	presignedURL, err := s.mediaService.GeneratePresignedURL(ctx, objectPath, expiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return &ReportResult{
		ObjectPath:   objectPath,
		TotalRows:    totalRows,
		PresignedURL: presignedURL,
	}, nil
}
