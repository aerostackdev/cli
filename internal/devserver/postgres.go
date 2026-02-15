package devserver

import (
	"fmt"
	"net/url"
	"strings"
)

// ValidatePostgresConnectionString validates a Postgres connection string
// Supports both postgresql:// and postgres:// schemes
func ValidatePostgresConnectionString(connStr string) error {
	if connStr == "" {
		return fmt.Errorf("connection string cannot be empty")
	}

	// Check if it's still using environment variable syntax (not yet interpolated)
	if strings.Contains(connStr, "$") {
		return fmt.Errorf("connection string contains unresolved environment variables: %s", connStr)
	}

	// Parse the connection string as a URL
	u, err := url.Parse(connStr)
	if err != nil {
		return fmt.Errorf("invalid connection string format: %w", err)
	}

	// Validate scheme
	if u.Scheme != "postgresql" && u.Scheme != "postgres" {
		return fmt.Errorf("invalid scheme '%s': must be 'postgresql://' or 'postgres://'", u.Scheme)
	}

	// Validate host
	if u.Host == "" {
		return fmt.Errorf("connection string must include a host")
	}

	// Validate database name (path component)
	if u.Path == "" || u.Path == "/" {
		return fmt.Errorf("connection string must include a database name")
	}

	return nil
}
