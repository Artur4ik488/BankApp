package main

import (
	"net/url"
	"testing"
)

func TestDatabaseURLFromPostgresEnvironment(t *testing.T) {
	t.Setenv("POSTGRES_DB", "bank app")
	t.Setenv("POSTGRES_USER", "bank@app")
	t.Setenv("POSTGRES_PASSWORD", "p@ss:word/with symbols")
	t.Setenv("POSTGRES_HOST", "127.0.0.1")
	t.Setenv("POSTGRES_PORT", "5433")
	t.Setenv("DATABASE_URL", "postgres://stale:wrong@localhost/old")

	databaseURL, err := databaseURLFromEnvironment()
	if err != nil {
		t.Fatal(err)
	}

	parsed, err := url.Parse(databaseURL)
	if err != nil {
		t.Fatalf("parse database URL: %v", err)
	}
	if parsed.Host != "127.0.0.1:5433" {
		t.Fatalf("expected host 127.0.0.1:5433, got %q", parsed.Host)
	}
	if parsed.Path != "/bank app" {
		t.Fatalf("expected database path /bank app, got %q", parsed.Path)
	}
	if parsed.User.Username() != "bank@app" {
		t.Fatalf("expected user bank@app, got %q", parsed.User.Username())
	}
	password, ok := parsed.User.Password()
	if !ok || password != "p@ss:word/with symbols" {
		t.Fatal("password was not encoded and decoded correctly")
	}
	if parsed.Query().Get("sslmode") != "disable" {
		t.Fatal("expected sslmode=disable")
	}
}

func TestDatabaseURLFallsBackToExplicitURL(t *testing.T) {
	t.Setenv("POSTGRES_DB", "")
	t.Setenv("POSTGRES_USER", "")
	t.Setenv("POSTGRES_PASSWORD", "")
	t.Setenv("DATABASE_URL", "postgres://example.test/bank")

	databaseURL, err := databaseURLFromEnvironment()
	if err != nil {
		t.Fatal(err)
	}
	if databaseURL != "postgres://example.test/bank" {
		t.Fatalf("unexpected database URL: %q", databaseURL)
	}
}

func TestDatabaseURLRequiresConfiguration(t *testing.T) {
	t.Setenv("POSTGRES_DB", "")
	t.Setenv("POSTGRES_USER", "")
	t.Setenv("POSTGRES_PASSWORD", "")
	t.Setenv("DATABASE_URL", "")

	if _, err := databaseURLFromEnvironment(); err == nil {
		t.Fatal("expected incomplete database configuration to fail")
	}
}
