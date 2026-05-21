package accounts

import (
	"kroncl-server/internal/core"
	"net/http"
)

func (h *Handlers) GetSummary(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	orgsCount, err := h.service.GetAccountCompaniesCount(r.Context(), accountID)
	if err != nil {
		core.SendError(w, http.StatusInternalServerError, "Failed to get organizations count")
		return
	}

	invitesCount, err := h.service.GetPendingInvitationsCount(r.Context(), accountID)
	if err != nil {
		core.SendError(w, http.StatusInternalServerError, "Failed to get invitations count")
		return
	}

	fingerprintsCount, err := h.service.GetActiveFingerprintsCount(r.Context(), accountID)
	if err != nil {
		core.SendError(w, http.StatusInternalServerError, "Failed to get fingerprints count")
		return
	}

	summary := SummaryCounters{
		OrganizationsCount: orgsCount,
		InvitationsCount:   invitesCount,
		FingerprintsCount:  fingerprintsCount,
	}

	core.SendSuccess(w, summary, "Account summary retrieved successfully")
}
