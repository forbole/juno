package utils

import (
	"os"
)

// GetEnvOr returns the value of the ENV variable having the given key, or the provided orValue
func GetEnvOr(envKey string, orValue string) string {
	if envValue := os.Getenv(envKey); envValue != "" {
		return envValue
	}
	return orValue
}
