package admindb

type MigrateTenantRequest struct {
	Command string `json:"command"` // up, down, version, force
	Steps   int    `json:"steps"`   // для down и force
}
