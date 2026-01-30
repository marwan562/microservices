package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/marwan562/fintech-ecosystem/internal/auth"
)

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name           string
		reqBody        string
		mockSetup      func(*auth.MockDB)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:    "Valid Login",
			reqBody: `{"email":"test@example.com","password":"password123"}`,
			mockSetup: func(m *auth.MockDB) {
				m.QueryRowContextFunc = func(ctx context.Context, query string, args ...any) auth.Row {
					return &auth.MockRow{
						ScanFunc: func(dest ...any) error {
							if strings.Contains(query, "SELECT id, email, password_hash") {
								// password_hash for "password123"
								*(dest[0].(*string)) = "user_123"
								*(dest[1].(*string)) = "test@example.com"
								*(dest[2].(*string)) = "$2a$10$8K1p/a06gr71Z6S.p.8.9eS6l.T7.P/.k1e0k2O2.Y2.Y2.Y2.Y2"
								return nil
							}
							return nil
						},
					}
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "token",
		},
		{
			name:           "Missing Email",
			reqBody:        `{"password":"password123"}`,
			mockSetup:      func(m *auth.MockDB) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Email and password are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mDB := &auth.MockDB{}
			tt.mockSetup(mDB)
			repo := auth.NewTestRepository(mDB)
			h := &AuthHandler{repo: repo}

			req := httptest.NewRequest("POST", "/login", strings.NewReader(tt.reqBody))
			w := httptest.NewRecorder()

			h.Login(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("%s: Expected status %d, got %d. Body: %s", tt.name, tt.expectedStatus, w.Code, w.Body.String())
			}
			if !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("%s: Expected body to contain '%s', got '%s'", tt.name, tt.expectedBody, w.Body.String())
			}
		})
	}
}

func TestAuthHandler_Register(t *testing.T) {
	tests := []struct {
		name           string
		reqBody        string
		mockSetup      func(*auth.MockDB)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:    "Successful Registration",
			reqBody: `{"email":"new@example.com","password":"password123"}`,
			mockSetup: func(m *auth.MockDB) {
				m.QueryRowContextFunc = func(ctx context.Context, query string, args ...any) auth.Row {
					return &auth.MockRow{
						ScanFunc: func(dest ...any) error {
							*(dest[0].(*string)) = "user_456"
							*(dest[1].(*string)) = "new@example.com"
							return nil
						},
					}
				}
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "new@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mDB := &auth.MockDB{}
			tt.mockSetup(mDB)
			repo := auth.NewTestRepository(mDB)
			h := &AuthHandler{repo: repo}

			req := httptest.NewRequest("POST", "/register", strings.NewReader(tt.reqBody))
			w := httptest.NewRecorder()

			h.Register(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("%s: Expected status %d, got %d", tt.name, tt.expectedStatus, w.Code)
			}
			if !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("%s: Expected body to contain '%s', got '%s'", tt.name, tt.expectedBody, w.Body.String())
			}
		})
	}
}
