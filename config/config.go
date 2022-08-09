package config

import (
	"os"
	"strings"
)

var config *Config

// LoadConfig reads the configuration variables
func LoadConfig() error {
	config = &Config{}

	config.AuthServiceAccounts = parseCsv("AUTH_SERVICE_ACCOUNTS")
	config.AuthAudience = os.Getenv("AUTH_AUDIENCE")

	return nil
}

// Config contains all the application configuration
type Config struct {
	AuthServiceAccounts []string
	AuthAudience        string
}

// GetConfig returns the current application configuration settings
func GetConfig() *Config {
	return config
}

func parseCsv(name string) []string {
	items := strings.Split(os.Getenv(name), ",")
	result := []string{}
	for _, v := range items {
		if v != "" {
			result = append(result, v)
		}
	}
	return result
}