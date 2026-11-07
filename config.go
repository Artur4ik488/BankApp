package main

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
)

func databaseURLFromEnvironment() (string, error) {
	databaseName := strings.TrimSpace(os.Getenv("POSTGRES_DB"))
	user := strings.TrimSpace(os.Getenv("POSTGRES_USER"))
	password := os.Getenv("POSTGRES_PASSWORD")

	if databaseName != "" && user != "" && password != "" {
		host := strings.TrimSpace(os.Getenv("POSTGRES_HOST"))
		if host == "" {
			host = "localhost"
		}

		port := strings.TrimSpace(os.Getenv("POSTGRES_PORT"))
		if port == "" {
			port = "5433"
		}

		connectionURL := &url.URL{
			Scheme: "postgres",
			User:   url.UserPassword(user, password),
			Host:   net.JoinHostPort(host, port),
			Path:   databaseName,
		}
		query := connectionURL.Query()
		query.Set("sslmode", "disable")
		connectionURL.RawQuery = query.Encode()
		return connectionURL.String(), nil
	}

	if databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL")); databaseURL != "" {
		return databaseURL, nil
	}

	return "", fmt.Errorf(
		"database configuration is incomplete: set POSTGRES_DB, POSTGRES_USER, and POSTGRES_PASSWORD, or set DATABASE_URL (use BANKAPP_STORE=memory for a temporary demo)",
	)
}
