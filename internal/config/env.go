package config

import "os"

func GetEnv(key, defaultValue string) string {
	// Look up the environment variable by key.
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}
