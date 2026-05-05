package admincompanies

import "kroncl-server/internal/companies"

type AdminCompany struct {
	companies.Company
	StorageStatus string `json:"storage_status"`
	StorageReady  bool   `json:"storage_ready"`
	SchemaName    string `json:"schema_name"`
}
