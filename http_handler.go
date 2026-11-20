package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"
)

//go:embed web/*
var webFiles embed.FS

type API struct {
	service *AccountService
}

func NewHTTPHandler(service *AccountService) http.Handler {
	api := &API{service: service}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", api.health)
	mux.HandleFunc("GET /api/accounts", api.listAccounts)
	mux.HandleFunc("POST /api/accounts", api.createAccount)
	mux.HandleFunc("GET /api/accounts/{id}", api.getAccount)
	mux.HandleFunc("POST /api/accounts/{id}/deposit", api.deposit)
	mux.HandleFunc("POST /api/accounts/{id}/withdraw", api.withdraw)
	mux.HandleFunc("POST /api/transfers", api.transfer)

	staticFiles, err := fs.Sub(webFiles, "web")
	if err != nil {
		panic(fmt.Sprintf("prepare frontend files: %v", err))
	}
	mux.Handle("/", http.FileServerFS(staticFiles))

	return securityHeaders(mux)
}

func (api *API) health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := api.service.Health(ctx); err != nil {
		slog.Error("health check failed", "error", err)
		writeAPIError(w, http.StatusServiceUnavailable, "service_unavailable", "account storage is unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (api *API) listAccounts(w http.ResponseWriter, r *http.Request) {
	accounts, err := api.service.ListAccounts(r.Context())
	if err != nil {
		api.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, accounts)
}

func (api *API) createAccount(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Owner string `json:"owner"`
	}
	if err := decodeJSON(w, r, &request); err != nil {
		writeJSONDecodeError(w, err)
		return
	}

	account, err := api.service.CreateAccount(r.Context(), request.Owner)
	if err != nil {
		api.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, account)
}

func (api *API) getAccount(w http.ResponseWriter, r *http.Request) {
	accountID, err := accountIDFromRequest(r)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_account_id", err.Error())
		return
	}

	account, err := api.service.GetAccount(r.Context(), accountID)
	if err != nil {
		api.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, account)
}

func (api *API) deposit(w http.ResponseWriter, r *http.Request) {
	api.changeBalance(w, r, api.service.Deposit)
}

func (api *API) withdraw(w http.ResponseWriter, r *http.Request) {
	api.changeBalance(w, r, api.service.Withdraw)
}

func (api *API) changeBalance(
	w http.ResponseWriter,
	r *http.Request,
	change func(context.Context, int, Money) error,
) {
	accountID, err := accountIDFromRequest(r)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_account_id", err.Error())
		return
	}

	var request struct {
		Amount Money `json:"amount"`
	}
	if err := decodeJSON(w, r, &request); err != nil {
		writeJSONDecodeError(w, err)
		return
	}

	if err := change(r.Context(), accountID, request.Amount); err != nil {
		api.writeError(w, err)
		return
	}

	account, err := api.service.GetAccount(r.Context(), accountID)
	if err != nil {
		api.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, account)
}

func (api *API) transfer(w http.ResponseWriter, r *http.Request) {
	var request struct {
		FromAccountID int   `json:"from_account_id"`
		ToAccountID   int   `json:"to_account_id"`
		Amount        Money `json:"amount"`
	}
	if err := decodeJSON(w, r, &request); err != nil {
		writeJSONDecodeError(w, err)
		return
	}
	if request.FromAccountID <= 0 || request.ToAccountID <= 0 {
		writeAPIError(w, http.StatusBadRequest, "invalid_account_id", "account IDs must be positive")
		return
	}

	if err := api.service.Transfer(
		r.Context(),
		request.FromAccountID,
		request.ToAccountID,
		request.Amount,
	); err != nil {
		api.writeError(w, err)
		return
	}

	fromAccount, err := api.service.GetAccount(r.Context(), request.FromAccountID)
	if err != nil {
		api.writeError(w, err)
		return
	}
	toAccount, err := api.service.GetAccount(r.Context(), request.ToAccountID)
	if err != nil {
		api.writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]Account{
		"from_account": fromAccount,
		"to_account":   toAccount,
	})
}

func (api *API) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidOwner), errors.Is(err, ErrInvalidAmount), errors.Is(err, ErrSameAccount):
		writeAPIError(w, http.StatusBadRequest, "validation_error", err.Error())
	case errors.Is(err, ErrAccountNotFound):
		writeAPIError(w, http.StatusNotFound, "account_not_found", err.Error())
	case errors.Is(err, ErrInsufficientFunds), errors.Is(err, ErrAccountExists), errors.Is(err, ErrBalanceLimit):
		writeAPIError(w, http.StatusConflict, "operation_rejected", err.Error())
	default:
		slog.Error("API request failed", "error", err)
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "an unexpected error occurred")
	}
}

func accountIDFromRequest(r *http.Request) (int, error) {
	accountID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || accountID <= 0 {
		return 0, fmt.Errorf("account ID must be a positive number")
	}
	return accountID, nil
}

var errUnsupportedMediaType = errors.New("content type must be application/json")

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) error {
	contentType := strings.TrimSpace(r.Header.Get("Content-Type"))
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil || mediaType != "application/json" {
		return errUnsupportedMediaType
	}

	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("decode JSON: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return fmt.Errorf("request body must contain one JSON object")
	}
	return nil
}

func writeJSONDecodeError(w http.ResponseWriter, err error) {
	if errors.Is(err, errUnsupportedMediaType) {
		writeAPIError(w, http.StatusUnsupportedMediaType, "unsupported_media_type", err.Error())
		return
	}
	var maxBytesError *http.MaxBytesError
	if errors.As(err, &maxBytesError) {
		writeAPIError(w, http.StatusRequestEntityTooLarge, "request_too_large", "request body must not exceed 1 MB")
		return
	}
	writeAPIError(w, http.StatusBadRequest, "invalid_request", err.Error())
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		slog.Error("write JSON response", "error", err)
	}
}

func writeAPIError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self'; script-src 'self'; connect-src 'self'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Cache-Control", "no-store")
		}
		next.ServeHTTP(w, r)
	})
}
