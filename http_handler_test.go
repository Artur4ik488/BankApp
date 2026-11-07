package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPAccountWorkflow(t *testing.T) {
	handler := NewHTTPHandler(NewAccountService(NewAccountStore()))

	var arthur Account
	doJSONRequest(t, handler, http.MethodPost, "/api/accounts", map[string]any{
		"owner": "  Arthur  ",
	}, http.StatusCreated, &arthur)
	if arthur.ID != 1 || arthur.Owner != "Arthur" || arthur.Balance != 0 {
		t.Fatalf("unexpected account: %+v", arthur)
	}

	var daniel Account
	doJSONRequest(t, handler, http.MethodPost, "/api/accounts", map[string]any{
		"owner": "Daniel",
	}, http.StatusCreated, &daniel)

	var updated Account
	doJSONRequest(t, handler, http.MethodPost, "/api/accounts/1/deposit", map[string]any{
		"amount": 100_00,
	}, http.StatusOK, &updated)
	if updated.Balance != 100_00 {
		t.Fatalf("expected Arthur's balance to be 10000, got %d", updated.Balance)
	}

	doJSONRequest(t, handler, http.MethodPost, "/api/accounts/2/deposit", map[string]any{
		"amount": 50_00,
	}, http.StatusOK, &updated)

	var transferResult map[string]Account
	doJSONRequest(t, handler, http.MethodPost, "/api/transfers", map[string]any{
		"from_account_id": arthur.ID,
		"to_account_id":   daniel.ID,
		"amount":          25_00,
	}, http.StatusOK, &transferResult)
	if transferResult["from_account"].Balance != 75_00 {
		t.Fatalf("expected source balance 7500, got %d", transferResult["from_account"].Balance)
	}
	if transferResult["to_account"].Balance != 75_00 {
		t.Fatalf("expected destination balance 7500, got %d", transferResult["to_account"].Balance)
	}

	doJSONRequest(t, handler, http.MethodPost, "/api/accounts/2/withdraw", map[string]any{
		"amount": 10_00,
	}, http.StatusOK, &updated)
	if updated.Balance != 65_00 {
		t.Fatalf("expected Daniel's balance to be 6500, got %d", updated.Balance)
	}

	var accounts []Account
	doJSONRequest(t, handler, http.MethodGet, "/api/accounts", nil, http.StatusOK, &accounts)
	if len(accounts) != 2 || accounts[0].Balance != 75_00 || accounts[1].Balance != 65_00 {
		t.Fatalf("unexpected account list: %+v", accounts)
	}
}

func TestHTTPAPIValidation(t *testing.T) {
	handler := NewHTTPHandler(NewAccountService(NewAccountStore()))

	tests := []struct {
		name       string
		method     string
		path       string
		body       any
		wantStatus int
		wantCode   string
	}{
		{
			name:       "empty owner",
			method:     http.MethodPost,
			path:       "/api/accounts",
			body:       map[string]any{"owner": "   "},
			wantStatus: http.StatusBadRequest,
			wantCode:   "validation_error",
		},
		{
			name:       "account not found",
			method:     http.MethodGet,
			path:       "/api/accounts/999",
			wantStatus: http.StatusNotFound,
			wantCode:   "account_not_found",
		},
		{
			name:       "invalid account ID",
			method:     http.MethodGet,
			path:       "/api/accounts/nope",
			wantStatus: http.StatusBadRequest,
			wantCode:   "invalid_account_id",
		},
		{
			name:       "unknown JSON field",
			method:     http.MethodPost,
			path:       "/api/accounts",
			body:       map[string]any{"owner": "Arthur", "admin": true},
			wantStatus: http.StatusBadRequest,
			wantCode:   "invalid_request",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var response struct {
				Error struct {
					Code string `json:"code"`
				} `json:"error"`
			}
			doJSONRequest(t, handler, test.method, test.path, test.body, test.wantStatus, &response)
			if response.Error.Code != test.wantCode {
				t.Fatalf("expected error code %q, got %q", test.wantCode, response.Error.Code)
			}
		})
	}
}

func TestHTTPServesFrontend(t *testing.T) {
	handler := NewHTTPHandler(NewAccountService(NewAccountStore()))
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), "Your money") {
		t.Fatal("frontend response does not contain the dashboard")
	}
	if contentType := response.Header().Get("Content-Type"); !strings.Contains(contentType, "text/html") {
		t.Fatalf("expected HTML content type, got %q", contentType)
	}
}

func doJSONRequest(
	t *testing.T,
	handler http.Handler,
	method, path string,
	body any,
	wantStatus int,
	target any,
) {
	t.Helper()

	var requestBody io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("encode request body: %v", err)
		}
		requestBody = bytes.NewReader(encoded)
	}

	request := httptest.NewRequest(method, path, requestBody)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != wantStatus {
		t.Fatalf("expected status %d, got %d: %s", wantStatus, response.Code, response.Body.String())
	}
	if target != nil {
		if err := json.NewDecoder(response.Body).Decode(target); err != nil {
			t.Fatalf("decode response body: %v", err)
		}
	}
}
