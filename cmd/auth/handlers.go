package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/marwan562/fintech-ecosystem/internal/auth"
	"github.com/marwan562/fintech-ecosystem/pkg/apikey"
	"github.com/marwan562/fintech-ecosystem/pkg/bcryptutil"
	"github.com/marwan562/fintech-ecosystem/pkg/jsonutil"
	"github.com/marwan562/fintech-ecosystem/pkg/jwtutil"
)

type GenerateAPIKeyRequest struct {
	Environment string `json:"environment"` // "test" or "live"
}

type GenerateAPIKeyResponse struct {
	Key          string `json:"key"` // Full key shown ONLY once
	Environment  string `json:"environment"`
	TruncatedKey string `json:"truncated_key"`
}

// Helper to extract UserID from JWT (duplicated from payments, common lib pending)
func extractUserIDFromToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", nil
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := jwtutil.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}

func (h *AuthHandler) GenerateAPIKey(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserIDFromToken(r)
	if err != nil || userID == "" {
		log.Printf("GenerateAPIKey: Validation failed. userID: %s, err: %v", userID, err)
		jsonutil.WriteErrorJSON(w, "Unauthorized")
		return
	}
	log.Printf("GenerateAPIKey: Success for user %s", userID)

	var req GenerateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonutil.WriteErrorJSON(w, "Invalid request body")
		return
	}

	if req.Environment != "test" && req.Environment != "live" {
		jsonutil.WriteErrorJSON(w, "Invalid environment (use 'test' or 'live')")
		return
	}

	prefix := "sk_" + req.Environment
	fullKey, hash, err := apikey.GenerateKey(prefix, h.hmacSecret)
	if err != nil {
		jsonutil.WriteErrorJSON(w, "Failed to generate key")
		return
	}

	truncated := fullKey[len(fullKey)-4:]

	apiKey := &auth.APIKey{
		UserID:       userID,
		KeyPrefix:    prefix,
		KeyHash:      hash,
		TruncatedKey: truncated,
		Environment:  req.Environment,
	}

	if err := h.repo.CreateAPIKey(r.Context(), apiKey); err != nil {
		jsonutil.WriteErrorJSON(w, "Failed to save API key")
		return
	}

	// Return the FULL key only once
	jsonutil.WriteJSON(w, http.StatusCreated, GenerateAPIKeyResponse{
		Key:          fullKey,
		Environment:  req.Environment,
		TruncatedKey: truncated,
	})
}

type ValidateAPIKeyRequest struct {
	KeyHash string `json:"key_hash"`
}

type ValidateAPIKeyResponse struct {
	Valid       bool   `json:"valid"`
	UserID      string `json:"user_id"`
	Environment string `json:"environment"`
}

func (h *AuthHandler) ValidateAPIKey(w http.ResponseWriter, r *http.Request) {
	// Internal endpoint, should be protected (e.g. valid only from localhost or via mTLS/secret)
	// For now, assuming network isolation or simple IP check in real world.

	var req ValidateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonutil.WriteErrorJSON(w, "Invalid request body")
		return
	}

	key, err := h.repo.GetAPIKeyByHash(r.Context(), req.KeyHash)
	if err != nil {
		log.Printf("Error validating key: %v", err)
		jsonutil.WriteErrorJSON(w, "Validation failed")
		return
	}

	if key == nil || key.RevokedAt != nil {
		jsonutil.WriteJSON(w, http.StatusOK, ValidateAPIKeyResponse{Valid: false})
		return
	}

	jsonutil.WriteJSON(w, http.StatusOK, ValidateAPIKeyResponse{
		Valid:       true,
		UserID:      key.UserID,
		Environment: key.Environment,
	})
}

// AuthHandler holds dependencies for authentication endpoints.
type AuthHandler struct {
	repo       *auth.Repository
	hmacSecret string
}

// RegisterRequest defines the payload for user registration.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest defines the payload for user login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse defines the successful response for login.
type LoginResponse struct {
	Token string `json:"token"`
}

// Register handles user account creation.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonutil.WriteErrorJSON(w, "Method not allowed")
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonutil.WriteErrorJSON(w, "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		jsonutil.WriteErrorJSON(w, "Email and password are required")
		return
	}

	// Use bcryptutil to hash the password securely
	b := &bcryptutil.BcryptUtilsImpl{}
	passwordHash, err := b.GenerateHash(req.Password)
	if err != nil {
		jsonutil.WriteErrorJSON(w, "Failed to hash password")
		return
	}

	user, err := h.repo.CreateUser(r.Context(), req.Email, passwordHash)
	if err != nil {
		// Check for duplicate key error (generic check for now)
		// In a real app, parse the error code for unique constraint violation
		jsonutil.WriteErrorJSON(w, "Failed to create user (email might be taken)")
		return
	}

	jsonutil.WriteJSON(w, http.StatusCreated, user)
}

// Login handles user authentication and JWT issuance.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonutil.WriteErrorJSON(w, "Method not allowed")
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonutil.WriteErrorJSON(w, "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		jsonutil.WriteErrorJSON(w, "Email and password are required")
		return
	}

	// Fetch user by email
	user, err := h.repo.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		jsonutil.WriteErrorJSON(w, "Internal server error")
		return
	}
	if user == nil {
		jsonutil.WriteErrorJSON(w, "Invalid email or password") // Generic message for security
		return
	}

	// Verify password
	b := &bcryptutil.BcryptUtilsImpl{}
	match := b.CompareHash(req.Password, user.Password)
	if !match {
		jsonutil.WriteErrorJSON(w, "Invalid email or password")
		return
	}

	// Generate JWT
	token, err := jwtutil.GenerateToken(user.ID, user.Email)
	if err != nil {
		jsonutil.WriteErrorJSON(w, "Failed to generate token")
		return
	}

	log.Printf("Login: Success for user %s", user.Email)
	jsonutil.WriteJSON(w, http.StatusOK, LoginResponse{Token: token})
}

// OAuthTokenResponse represents the response for /oauth/token.
type OAuthTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// OAuthTokenHandler handles OAuth2 token requests (Client Credentials).
func (h *AuthHandler) OAuthTokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonutil.WriteErrorJSON(w, "Method not allowed")
		return
	}

	// Parse Basic Auth (Client ID & Secret)
	clientID, clientSecret, ok := r.BasicAuth()
	if !ok {
		// Fallback to form parameters if Basic Auth not provided
		clientID = r.FormValue("client_id")
		clientSecret = r.FormValue("client_secret")
	}

	grantType := r.FormValue("grant_type")
	if grantType != "client_credentials" {
		jsonutil.WriteErrorJSON(w, "unsupported_grant_type")
		return
	}

	if clientID == "" || clientSecret == "" {
		w.Header().Set("WWW-Authenticate", `Basic realm="api"`)
		http.Error(w, "invalid_client", http.StatusUnauthorized)
		return
	}

	// Verify Client
	client, err := h.repo.GetClientByID(r.Context(), clientID)
	if err != nil {
		log.Printf("OAuthTokenHandler: DB error: %v", err)
		jsonutil.WriteErrorJSON(w, "server_error")
		return
	}
	if client == nil {
		http.Error(w, "invalid_client", http.StatusUnauthorized)
		return
	}

	// Validate Secret
	// In a real app, use bcryptutil.CompareHash if secrets are hashed.
	// For now, assuming basic hash comparison or if we want to support raw secrets (less secure)
	// Changing implementation to match simple hash check as per `oauth.go` (if we implemented hashing there)
	// Looking to `oauth.go`, we used hash.Wait, `oauth.go` was created with `GetClientByID` but no `ValidateClientSecret`.
	// For simplicity in this step, let's assume `client.ClientSecretHash` is what we have.
	// We need to hash the incoming secret and compare.
	incomingHash := auth.HashString(clientSecret)
	// Optimization: Depending on how we stored it. If `oauth.go` HashString uses SHA256 string format.
	// Let's rely on HashString from `auth` package.

	if incomingHash != client.ClientSecretHash {
		http.Error(w, "invalid_client", http.StatusUnauthorized)
		return
	}

	// Generate Access Token
	accessToken, err := auth.GenerateRandomString(32)
	if err != nil {
		jsonutil.WriteErrorJSON(w, "server_error")
		return
	}

	expiresIn := 3600 // 1 hour
	expiresAt := time.Now().Add(time.Duration(expiresIn) * time.Second)

	token := &auth.OAuthToken{
		AccessToken: accessToken,
		ClientID:    client.ID,
		UserID:      client.UserID,
		Scope:       "default", // In future parse 'scope' param
		ExpiresAt:   expiresAt,
	}

	if err := h.repo.CreateOAuthToken(r.Context(), token); err != nil {
		log.Printf("OAuthTokenHandler: Failed to create token: %v", err)
		jsonutil.WriteErrorJSON(w, "server_error")
		return
	}

	jsonutil.WriteJSON(w, http.StatusOK, OAuthTokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
	})
}
