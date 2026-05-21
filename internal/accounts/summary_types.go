package accounts

type SummaryCounters struct {
	OrganizationsCount int `json:"organizations_count"`
	InvitationsCount   int `json:"invitations_count"`
	FingerprintsCount  int `json:"fingerprints_count"`
}
