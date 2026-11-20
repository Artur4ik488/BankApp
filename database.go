package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresPool(
	ctx context.Context,
	databaseURL string,
) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}

	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	var lastErr error
	for {
		if err := pool.Ping(connectCtx); err == nil {
			return pool, nil
		} else {
			lastErr = err
		}

		select {
		case <-connectCtx.Done():
			pool.Close()
			return nil, fmt.Errorf("database did not become ready: %w", lastErr)
		case <-ticker.C:
		}
	}
}
