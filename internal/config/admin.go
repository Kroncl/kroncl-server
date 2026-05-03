package config

import "os"

const (
	ADMIN_LEVEL_1 = 1 // нищий
	ADMIN_LEVEL_2 = 2
	ADMIN_LEVEL_3 = 3
	ADMIN_LEVEL_4 = 4
	ADMIN_LEVEL_5 = 5 // самый пиздатый

	ADMIN_LEVEL_MIN = 1
	ADMIN_LEVEL_MAX = 5
)

func GetFirstAdminEmail() string {
	email := os.Getenv("FIRST_ADMIN_EMAIL")
	return email
}
