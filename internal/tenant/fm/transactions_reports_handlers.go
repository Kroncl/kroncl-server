package fm

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"kroncl-server/internal/config"
	"kroncl-server/internal/core"
	"kroncl-server/internal/tenant/logs"
)

func (h *Handlers) GenerateTransactionsReport(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var startDate, endDate time.Time
	var err error

	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	if startDateStr == "" && endDateStr == "" {
		now := time.Now()
		startDate = now.AddDate(0, -1, 0)
		endDate = now
	} else {
		if startDateStr == "" || endDateStr == "" {
			core.SendError(w, http.StatusBadRequest, "Both start_date and end_date must be provided together")
			return
		}

		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			core.SendError(w, http.StatusBadRequest, "Invalid start_date format. Use YYYY-MM-DD")
			return
		}

		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			core.SendError(w, http.StatusBadRequest, "Invalid end_date format. Use YYYY-MM-DD")
			return
		}
	}

	endDate = endDate.Add(24*time.Hour - time.Second)

	comment := r.URL.Query().Get("comment")
	var commentPtr *string
	if comment != "" {
		commentPtr = &comment
	}

	report, total, err := h.repository.CreateTransactionReport(r.Context(), startDate, endDate, commentPtr)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS_REPORTS_CREATE, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
		)

		if strings.Contains(err.Error(), "too many transactions") {
			core.SendError(w, http.StatusBadRequest, err.Error())
			return
		}

		core.SendInternalError(w, fmt.Sprintf("Failed to generate report: %s", err.Error()))
		return
	}

	presignedURL, err := h.repository.mediaService.GeneratePresignedURL(r.Context(), report.ObjectPath, 1*time.Hour)
	if err != nil {
		core.SendInternalError(w, fmt.Sprintf("Failed to generate download URL: %s", err.Error()))
		return
	}

	h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS_REPORTS_CREATE, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("report_id", report.ID),
		logs.WithMetadata("total_transactions", total),
	)

	core.SendSuccess(w, map[string]interface{}{
		"download_url": presignedURL,
		"report":       report,
		"total":        total,
	}, "Report generated successfully")
}

func (h *Handlers) GetTransactionsReports(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pagination := core.GetDefaultPaginationParams(r)

	search := r.URL.Query().Get("search")
	var searchPtr *string
	if search != "" {
		searchPtr = &search
	}

	reports, total, err := h.repository.GetTransactionReports(r.Context(), pagination.Offset, pagination.Limit, searchPtr)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_FM_TRANSACTIONS_REPORTS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
		)
		core.SendInternalError(w, "Failed to get reports")
		return
	}

	response := map[string]interface{}{
		"reports":    reports,
		"pagination": core.NewPagination(int(total), pagination.Page, pagination.Limit),
	}

	core.SendSuccess(w, response, "Reports retrieved successfully")
}
