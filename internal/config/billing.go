package config

import (
	"os"
	"strings"
)

const (
	BILLING_MODE_ON  = "on"
	BILLING_MODE_OFF = "off"
)

func GetBillingMode() string {
	mode := strings.ToLower(os.Getenv("BILLING_MODE"))
	if mode == BILLING_MODE_OFF {
		return BILLING_MODE_OFF
	}
	return BILLING_MODE_ON
}

func IsBillingEnabled() bool {
	return GetBillingMode() == BILLING_MODE_ON
}
