package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds the application configuration
type Config struct {
	// Database
	DatabasePath string

	// JWT
	JWTSecret          string
	JWTExpirationHours int

	// Admin
	AdminEmails []string

	// Gmail API
	GmailClientID     string
	GmailClientSecret string
	GmailRefreshToken string
	GmailFromEmail    string

	// Uploads
	UploadDir        string
	MaxUploadSizeMB  int

	// System Settings
	BookingAdvanceDays      int
	CancellationNoticeHours int
	AutoDeactivationDays    int

	// Server
	Port        string
	Environment string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		// Database
		DatabasePath: getEnv("DATABASE_PATH", "./gassigeher.db"),

		// JWT
		JWTSecret:          getEnv("JWT_SECRET", "change-this-in-production"),
		JWTExpirationHours: getEnvAsInt("JWT_EXPIRATION_HOURS", 24),

		// Admin
		AdminEmails: getEnvAsSlice("ADMIN_EMAILS", ","),

		// Gmail API
		GmailClientID:     getEnv("GMAIL_CLIENT_ID", ""),
		GmailClientSecret: getEnv("GMAIL_CLIENT_SECRET", ""),
		GmailRefreshToken: getEnv("GMAIL_REFRESH_TOKEN", ""),
		GmailFromEmail:    getEnv("GMAIL_FROM_EMAIL", "noreply@gassigeher.com"),

		// Uploads
		UploadDir:       getEnv("UPLOAD_DIR", "./uploads"),
		MaxUploadSizeMB: getEnvAsInt("MAX_UPLOAD_SIZE_MB", 5),

		// System Settings
		BookingAdvanceDays:      getEnvAsInt("BOOKING_ADVANCE_DAYS", 14),
		CancellationNoticeHours: getEnvAsInt("CANCELLATION_NOTICE_HOURS", 12),
		AutoDeactivationDays:    getEnvAsInt("AUTO_DEACTIVATION_DAYS", 365),

		// Server
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
	}
}

// IsAdmin checks if the given email is an admin
func (c *Config) IsAdmin(email string) bool {
	email = strings.TrimSpace(strings.ToLower(email))
	for _, adminEmail := range c.AdminEmails {
		if strings.TrimSpace(strings.ToLower(adminEmail)) == email {
			return true
		}
	}
	return false
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsSlice(key, sep string) []string {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return []string{}
	}
	return strings.Split(valueStr, sep)
}
