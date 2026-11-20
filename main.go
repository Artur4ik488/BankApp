package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil {
		slog.Error("application failed", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	var repo AccountRepository
	if os.Getenv("BANKAPP_STORE") == "memory" {
		repo = NewAccountStore()
		slog.Info("using in-memory account store")
	} else {
		databaseURL, err := databaseURLFromEnvironment()
		if err != nil {
			return err
		}

		pool, err := NewPostgresPool(ctx, databaseURL)
		if err != nil {
			return fmt.Errorf("connect to database: %w", err)
		}
		defer pool.Close()

		repo = NewPostgresAccountRepository(pool)
		slog.Info("database connection established")
	}

	service := NewAccountService(repo)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return fmt.Errorf("PORT must be a number between 1 and 65535")
	}
	host := strings.TrimSpace(os.Getenv("HOST"))
	if host == "" {
		host = "127.0.0.1"
	}
	address := net.JoinHostPort(host, port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", address, err)
	}
	defer listener.Close()

	server := &http.Server{
		Addr:              address,
		Handler:           NewHTTPHandler(service),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverError := make(chan error, 1)
	go func() {
		serverError <- server.Serve(listener)
	}()
	slog.Info("BankApp is ready", "url", "http://localhost:"+port)

	select {
	case err := <-serverError:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("serve HTTP: %w", err)
		}
		return nil
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown HTTP server: %w", err)
		}
		slog.Info("BankApp stopped")
		return nil
	}
}
