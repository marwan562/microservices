package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/marwan562/fintech-ecosystem/internal/payment"
)

func TestPaymentHandler_CreatePaymentIntent(t *testing.T) {
	tests := []struct {
		name           string
		reqBody        string
		headers        map[string]string
		mockSetup      func(*payment.MockDB)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:    "Valid Request",
			reqBody: `{"amount":1000,"currency":"USD"}`,
			headers: map[string]string{"X-User-ID": "user_123"},
			mockSetup: func(m *payment.MockDB) {
				m.QueryRowContextFunc = func(ctx context.Context, query string, args ...any) payment.Row {
					return &payment.MockRow{
						ScanFunc: func(dest ...any) error {
							*(dest[0].(*string)) = "pi_123"
							return nil
						},
					}
				}
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `"id":"pi_123"`,
		},
		{
			name:           "Unauthorized",
			reqBody:        `{"amount":1000,"currency":"USD"}`,
			headers:        map[string]string{},
			mockSetup:      func(m *payment.MockDB) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Authentication required",
		},
		{
			name:           "Invalid Amount",
			reqBody:        `{"amount":-50,"currency":"USD"}`,
			headers:        map[string]string{"X-User-ID": "user_123"},
			mockSetup:      func(m *payment.MockDB) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Amount and Currency are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mDB := &payment.MockDB{}
			tt.mockSetup(mDB)
			repo := payment.NewTestRepository(mDB)
			h := &PaymentHandler{repo: repo}

			req := httptest.NewRequest("POST", "/payment_intents", strings.NewReader(tt.reqBody))
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			w := httptest.NewRecorder()

			h.CreatePaymentIntent(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
			if !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("Expected body to contain '%s', got '%s'", tt.expectedBody, w.Body.String())
			}
		})
	}
}
