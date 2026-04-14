package config

import (
	"net/http"
	"os"
)

type EnvMode string

const (
	EnvDevelopment EnvMode = "development"
	EnvProduction  EnvMode = "production"

	AUTH_REFRESH_PATH = "/api/account/refresh"
)

func GetEnvMode() EnvMode {
	env := os.Getenv("ENV")
	if env == "" {
		env = os.Getenv("ENVIRONMENT")
	}
	if env == "" {
		return EnvDevelopment
	}
	if env == "production" || env == "prod" {
		return EnvProduction
	}
	return EnvDevelopment
}

func IsProduction() bool {
	return GetEnvMode() == EnvProduction
}

func GetBaseDomain() string {
	domain := os.Getenv("BASE_DOMAIN")
	if domain == "" {
		if IsProduction() {
			return "kroncl.com"
		}
		return "localhost"
	}
	return domain
}

func GetClientDomain() string {
	domain := os.Getenv("CLIENT_DOMAIN")
	if domain == "" {
		if IsProduction() {
			return "kroncl.com"
		}
		return "localhost:3000"
	}
	return domain
}

func GetCookieDomain() string {
	if IsProduction() {
		return "." + GetBaseDomain()
	}
	return ""
}

func GetCookieSecure() bool {
	return IsProduction()
}

func GetCookieSameSite() http.SameSite {
	if IsProduction() {
		return http.SameSiteNoneMode
	}
	return http.SameSiteNoneMode
}
