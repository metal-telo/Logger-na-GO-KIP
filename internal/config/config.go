package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	// Server
	Port string

	// Logging
	LogDir            string
	LogFile           string
	LogLevel          string // debug|info|warn|error
	EnableHTTPLogging bool   // request logging middleware

	// Metrics
	MetricsDir  string
	MetricsFile string

	// Tracing (Jaeger)
	EnableTracing bool
	JaegerURL      string
	ServiceName    string

	// Load testing helpers
	AllowDuplicatePassports bool
}

func Load() Config {
	return Config{
		Port:                 envString("PORT", ":8080"),
		LogDir:               envString("LOG_DIR", "logs"),
		LogFile:              envString("LOG_FILE", "app.log"),
		LogLevel:             envString("LOG_LEVEL", "info"),
		EnableHTTPLogging:    envBool("HTTP_LOGGING", true),
		MetricsDir:           envString("METRICS_DIR", "metrics"),
		MetricsFile:          envString("METRICS_FILE", "metrics.log"),
		EnableTracing:        envBool("TRACING_ENABLED", false),
		JaegerURL:            envString("JAEGER_URL", "http://localhost:14268/api/traces"),
		ServiceName:          envString("SERVICE_NAME", "employee-management-system"),
		AllowDuplicatePassports: envBool("ALLOW_DUPLICATE_PASSPORTS", false),
	}
}

func envString(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func envBool(key string, def bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	// Accept: 1/0, true/false, yes/no
	if b, err := strconv.ParseBool(v); err == nil {
		return b
	}
	switch strings.ToLower(v) {
	case "y", "yes", "on", "enable", "enabled":
		return true
	case "n", "no", "off", "disable", "disabled":
		return false
	default:
		return def
	}
}




