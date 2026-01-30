package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/marwan562/fintech-ecosystem/internal/ledger"
)

// MockDB for handler tests
func TestLedgerHandler_CreateAccount(t *testing.T) {
	tests := []struct {
		name           string
		reqBody        string
		mockSetup      func(*ledger.MockDB)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:    "Valid Request",
			reqBody: `{"name":"Checking","type":"asset","currency":"USD"}`,
			mockSetup: func(m *ledger.MockDB) {
				m.QueryRowContextFunc = func(ctx context.Context, query string, args ...any) ledger.Row {
					return &ledger.MockRow{
						ScanFunc: func(dest ...any) error {
							*(dest[0].(*string)) = "acc_123"
							return nil
						},
					}
				}
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `"id":"acc_123"`,
		},
		{
			name:           "Missing Name",
			reqBody:        `{"type":"asset"}`,
			mockSetup:      func(m *ledger.MockDB) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Name and Type are required",
		},
		{
			name:    "Repo Error",
			reqBody: `{"name":"Checking","type":"asset"}`,
			mockSetup: func(m *ledger.MockDB) {
				m.QueryRowContextFunc = func(ctx context.Context, query string, args ...any) ledger.Row {
					return &ledger.MockRow{
						ScanFunc: func(dest ...any) error {
							return errors.New("db error")
						},
					}
				}
			},
			expectedStatus: http.StatusBadRequest, // WriteErrorJSON defaults to 400
			expectedBody:   "Failed to create account",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mDB := &ledger.MockDB{}
			tt.mockSetup(mDB)
			repo := ledger.NewTestRepository(mDB)
			h := &LedgerHandler{repo: repo}

			req := httptest.NewRequest("POST", "/accounts", strings.NewReader(tt.reqBody))
			w := httptest.NewRecorder()

			h.CreateAccount(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
			if !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("Expected body to contain '%s', got '%s'", tt.expectedBody, w.Body.String())
			}
		})
	}
}

func TestLedgerHandler_RecordTransaction(t *testing.T) {
	tests := []struct {
		name           string
		reqBody        string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Missing ReferenceID",
			reqBody:        `{"entries":[{"account_id":"acc_1","amount":100},{"account_id":"acc_2","amount":-100}]}`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "ReferenceID required",
		},
		{
			name:           "Invalid JSON",
			reqBody:        `{invalid}`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid request body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &ledger.Repository{} // Empty repo is fine for validation-only tests
			h := &LedgerHandler{repo: repo}

			req := httptest.NewRequest("POST", "/transactions", strings.NewReader(tt.reqBody))
			w := httptest.NewRecorder()

			h.RecordTransaction(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
			if !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("Expected body to contain '%s', got '%s'", tt.expectedBody, w.Body.String())
			}
		})
	}
}
