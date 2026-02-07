package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sapliy/fintech-ecosystem/internal/auth/domain"
	"github.com/sapliy/fintech-ecosystem/pkg/apikey"
	"github.com/sapliy/fintech-ecosystem/pkg/bcryptutil"
	"github.com/sapliy/fintech-ecosystem/pkg/jsonutil"
	"github.com/sapliy/fintech-ecosystem/pkg/jwtutil"
)

type GenerateAPIKeyRequest struct {
	ZoneID      string `json:"zone_id"`
	Environment string `json:"environment"` // "test" or "live"
	Type        string `json:"type"`        // "secret" or "publishable"
}

type GenerateAPIKeyResponse struct {
	Key          string `json:"key"` // Full key shown ONLY once
	Environment  string `json:"environment"`
	ZoneID       string `json:"zone_id"`
	Mode         string `json:"mode"`
	Type         string `json:"type"`
	TruncatedKey string `json:"truncated_key"`
}

// Helper to extract UserID from JWT
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

	if req.ZoneID == "" {
		jsonutil.WriteErrorJSON(w, "zone_id is required")
		return
	}

	if req.Type == "" {
		req.Type = "secret"
	}

	prefix := "sk_"
	if req.Type == "publishable" {
		prefix = "pk_"
	}
	prefix += req.Environment

	fullKey, hash, err := apikey.GenerateKey(prefix, h.hmacSecret)
	if err != nil {
		jsonutil.WriteErrorJSON(w, "Failed to generate key")
		return
	}

	truncated := fullKey[len(fullKey)-4:]

	key := &domain.APIKey{
		UserID:       userID,
		OrgID:        "", // Optional, will be linked via zone or user
		ZoneID:       req.ZoneID,
		Mode:         req.Environment,
		KeyPrefix:    prefix,
		KeyHash:      hash,
		TruncatedKey: truncated,
		Environment:  req.Environment,
		Type:         req.Type,
	}

	if err := h.service.CreateAPIKey(r.Context(), key); err != nil {
		log.Printf("Failed to save API key: %v", err)
		jsonutil.WriteErrorJSON(w, "Failed to save API key (verify zone_id)")
		return
	}

	// Return the FULL key only once
	jsonutil.WriteJSON(w, http.StatusCreated, GenerateAPIKeyResponse{
		Key:          fullKey,
		Environment:  req.Environment,
		ZoneID:       req.ZoneID,
		Mode:         req.Environment,
		Type:         req.Type,
		TruncatedKey: truncated,
	})
}

type ValidateAPIKeyRequest struct {
	KeyHash string `json:"key_hash"`
}

type ValidateAPIKeyResponse struct {
	Valid       bool   `json:"valid"`
	UserID      string `json:"user_id"`
	OrgID       string `json:"org_id"`
	ZoneID      string `json:"zone_id"`
	Mode        string `json:"mode"`
	Environment string `json:"environment"`
	Scopes      string `json:"scopes"`
	Type        string `json:"type"`
}

func (h *AuthHandler) ValidateAPIKey(w http.ResponseWriter, r *http.Request) {
	var req ValidateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonutil.WriteErrorJSON(w, "Invalid request body")
		return
	}

	key, err := h.service.GetAPIKeyByHash(r.Context(), req.KeyHash)
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
		OrgID:       key.OrgID,
		ZoneID:      key.ZoneID,
		Mode:        key.Mode,
		Environment: key.Environment,
		Scopes:      key.Scopes,
		Type:        key.Type,
	})
}

type CreateOrganizationRequest struct {
	Name   string `json:"name"`
	Domain string `json:"domain"`
}

func (h *AuthHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserIDFromToken(r)
	if err != nil || userID == "" {
		jsonutil.WriteErrorJSON(w, "Unauthorized")
		return
	}

	var req CreateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonutil.WriteErrorJSON(w, "Invalid request body")
		return
	}

	if req.Name == "" {
		jsonutil.WriteErrorJSON(w, "Name is required")
		return
	}

	org, err := h.service.CreateOrganization(r.Context(), req.Name, req.Domain)
	if err != nil {
		log.Printf("Failed to create organization: %v", err)
		jsonutil.WriteErrorJSON(w, "Failed to create organization")
		return
	}

	// Add creator as owner
	if err := h.service.AddMember(r.Context(), userID, org.ID, domain.RoleOwner); err != nil {
		log.Printf("Failed to add owner to organization: %v", err)
	}

	jsonutil.WriteJSON(w, http.StatusCreated, org)
}

// AuthHandler holds dependencies for authentication endpoints.
type AuthHandler struct {
	service    *domain.AuthService
	hmacSecret string
	rdb        *redis.Client
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
	Token string       `json:"token"`
	User  *domain.User `json:"user"`
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

	b := &bcryptutil.BcryptUtilsImpl{}
	passwordHash, err := b.GenerateHash(req.Password)
	if err != nil {
		jsonutil.WriteErrorJSON(w, "Failed to hash password")
		return
	}

	user, err := h.service.CreateUser(r.Context(), req.Email, passwordHash)
	if err != nil {
		jsonutil.WriteErrorJSON(w, "Failed to create user (email might be taken)")
		return
	}

	// Generate verification token
	token, err := h.service.CreateEmailVerificationToken(r.Context(), user.ID)
	if err != nil {
		// Log error but don't fail registration
		log.Printf("Register: Failed to create verification token: %v", err)
	} else {
		log.Printf("Email verification token for %s: %s", user.Email, token)

		// Store in Redis for debug/testing
		if h.rdb != nil {
			h.rdb.Set(r.Context(), "debug:verify:"+user.Email, token, 1*time.Hour)
		}
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

	user, err := h.service.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		jsonutil.WriteErrorJSON(w, "Internal server error")
		return
	}
	if user == nil {
		jsonutil.WriteErrorJSON(w, "Invalid email or password")
		return
	}

	b := &bcryptutil.BcryptUtilsImpl{}
	match := b.CompareHash(req.Password, user.Password)
	if !match {
		jsonutil.WriteErrorJSON(w, "Invalid email or password")
		return
	}

	token, err := jwtutil.GenerateToken(user.ID, user.Email)
	if err != nil {
		jsonutil.WriteErrorJSON(w, "Failed to generate token")
		return
	}

	// Hide password hash in response
	user.Password = ""

	log.Printf("Login: Success for user %s", user.Email)
	jsonutil.WriteJSON(w, http.StatusOK, LoginResponse{
		Token: token,
		User:  user,
	})
}

// OAuthTokenResponse represents the response for /oauth/token.
type OAuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// OAuthTokenHandler handles OAuth2 token requests.
func (h *AuthHandler) OAuthTokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonutil.WriteErrorJSON(w, "Method not allowed")
		return
	}

	grantType := r.FormValue("grant_type")

	switch grantType {
	case "client_credentials":
		h.handleClientCredentials(w, r)
	case "authorization_code":
		h.handleAuthorizationCode(w, r)
	default:
		writeOAuthError(w, "unsupported_grant_type", "Grant type not supported", http.StatusBadRequest)
	}
}

func (h *AuthHandler) handleClientCredentials(w http.ResponseWriter, r *http.Request) {
	clientID, clientSecret, ok := r.BasicAuth()
	if !ok {
		clientID = r.FormValue("client_id")
		clientSecret = r.FormValue("client_secret")
	}

	if clientID == "" || clientSecret == "" {
		w.Header().Set("WWW-Authenticate", `Basic realm="api"`)
		writeOAuthError(w, "invalid_client", "Client credentials required", http.StatusUnauthorized)
		return
	}

	client, err := h.service.GetClientByID(r.Context(), clientID)
	if err != nil {
		log.Printf("handleClientCredentials: DB error: %v", err)
		writeOAuthError(w, "server_error", "Internal error", http.StatusInternalServerError)
		return
	}
	if client == nil {
		writeOAuthError(w, "invalid_client", "Unknown client", http.StatusUnauthorized)
		return
	}

	incomingHash := h.service.HashString(clientSecret)
	if incomingHash != client.ClientSecretHash {
		writeOAuthError(w, "invalid_client", "Invalid credentials", http.StatusUnauthorized)
		return
	}

	scope := r.FormValue("scope")
	if scope == "" {
		scope = "default"
	}

	h.issueTokens(w, r, client.ID, client.UserID, scope)
}

func (h *AuthHandler) handleAuthorizationCode(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	redirectURI := r.FormValue("redirect_uri")
	codeVerifier := r.FormValue("code_verifier")
	clientID := r.FormValue("client_id")

	if code == "" {
		writeOAuthError(w, "invalid_request", "Authorization code required", http.StatusBadRequest)
		return
	}

	authCode, err := h.service.GetAuthorizationCode(r.Context(), code)
	if err != nil {
		log.Printf("handleAuthorizationCode: DB error: %v", err)
		writeOAuthError(w, "server_error", "Internal error", http.StatusInternalServerError)
		return
	}
	if authCode == nil {
		writeOAuthError(w, "invalid_grant", "Invalid authorization code", http.StatusBadRequest)
		return
	}

	if authCode.Used {
		writeOAuthError(w, "invalid_grant", "Authorization code already used", http.StatusBadRequest)
		return
	}

	if time.Now().After(authCode.ExpiresAt) {
		writeOAuthError(w, "invalid_grant", "Authorization code expired", http.StatusBadRequest)
		return
	}

	if clientID != "" && clientID != authCode.ClientID {
		writeOAuthError(w, "invalid_grant", "Client mismatch", http.StatusBadRequest)
		return
	}

	if redirectURI != "" && redirectURI != authCode.RedirectURI {
		writeOAuthError(w, "invalid_grant", "Redirect URI mismatch", http.StatusBadRequest)
		return
	}

	if authCode.CodeChallenge != "" {
		if codeVerifier == "" {
			writeOAuthError(w, "invalid_request", "Code verifier required", http.StatusBadRequest)
			return
		}
		if !h.service.VerifyCodeChallenge(codeVerifier, authCode.CodeChallenge, authCode.CodeChallengeMethod) {
			writeOAuthError(w, "invalid_grant", "Invalid code verifier", http.StatusBadRequest)
			return
		}
	}

	if err := h.service.MarkAuthorizationCodeUsed(r.Context(), code); err != nil {
		log.Printf("handleAuthorizationCode: Failed to mark code used: %v", err)
		writeOAuthError(w, "server_error", "Internal error", http.StatusInternalServerError)
		return
	}

	h.issueTokens(w, r, authCode.ClientID, authCode.UserID, authCode.Scope)
}

func (h *AuthHandler) issueTokens(w http.ResponseWriter, r *http.Request, clientID, userID, scope string) {
	accessToken, err := h.service.GenerateRandomString(32)
	if err != nil {
		writeOAuthError(w, "server_error", "Failed to generate token", http.StatusInternalServerError)
		return
	}

	refreshToken, err := h.service.GenerateRandomString(32)
	if err != nil {
		writeOAuthError(w, "server_error", "Failed to generate token", http.StatusInternalServerError)
		return
	}

	expiresIn := 3600 // 1 hour
	expiresAt := time.Now().Add(time.Duration(expiresIn) * time.Second)

	token := &domain.OAuthToken{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ClientID:     clientID,
		UserID:       userID,
		Scope:        scope,
		ExpiresAt:    expiresAt,
	}

	if err := h.service.CreateOAuthToken(r.Context(), token); err != nil {
		log.Printf("issueTokens: Failed to create token: %v", err)
		writeOAuthError(w, "server_error", "Failed to store token", http.StatusInternalServerError)
		return
	}

	jsonutil.WriteJSON(w, http.StatusOK, OAuthTokenResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
		RefreshToken: refreshToken,
		Scope:        scope,
	})
}

// AuthorizeRequest holds the parsed authorize endpoint parameters.
type AuthorizeRequest struct {
	ResponseType        string
	ClientID            string
	RedirectURI         string
	Scope               string
	State               string
	CodeChallenge       string
	CodeChallengeMethod string
}

func (h *AuthHandler) AuthorizeHandler(w http.ResponseWriter, r *http.Request) {
	req := AuthorizeRequest{
		ResponseType:        r.FormValue("response_type"),
		ClientID:            r.FormValue("client_id"),
		RedirectURI:         r.FormValue("redirect_uri"),
		Scope:               r.FormValue("scope"),
		State:               r.FormValue("state"),
		CodeChallenge:       r.FormValue("code_challenge"),
		CodeChallengeMethod: r.FormValue("code_challenge_method"),
	}

	if req.ResponseType != "code" {
		writeOAuthError(w, "unsupported_response_type", "Only 'code' response type supported", http.StatusBadRequest)
		return
	}

	client, err := h.service.GetClientByID(r.Context(), req.ClientID)
	if err != nil {
		log.Printf("AuthorizeHandler: DB error: %v", err)
		writeOAuthError(w, "server_error", "Internal error", http.StatusInternalServerError)
		return
	}
	if client == nil {
		writeOAuthError(w, "invalid_request", "Unknown client", http.StatusBadRequest)
		return
	}

	if req.RedirectURI == "" {
		writeOAuthError(w, "invalid_request", "Redirect URI required", http.StatusBadRequest)
		return
	}

	validURI, err := h.service.ValidateRedirectURI(r.Context(), req.ClientID, req.RedirectURI)
	if err != nil {
		log.Printf("AuthorizeHandler: Failed to validate redirect URI: %v", err)
		writeOAuthError(w, "server_error", "Internal error", http.StatusInternalServerError)
		return
	}
	if !validURI {
		writeOAuthError(w, "invalid_request", "Invalid redirect URI", http.StatusBadRequest)
		return
	}

	if client.IsPublic && req.CodeChallenge == "" {
		writeOAuthError(w, "invalid_request", "PKCE required for public clients", http.StatusBadRequest)
		return
	}

	if req.CodeChallenge != "" && req.CodeChallengeMethod == "" {
		req.CodeChallengeMethod = "S256"
	}

	userID, err := extractUserIDFromToken(r)
	if err != nil || userID == "" {
		writeOAuthError(w, "access_denied", "User not authenticated", http.StatusUnauthorized)
		return
	}

	codeValue, err := h.service.GenerateRandomString(32)
	if err != nil {
		writeOAuthError(w, "server_error", "Failed to generate code", http.StatusInternalServerError)
		return
	}

	authCode := &domain.AuthorizationCode{
		Code:                codeValue,
		ClientID:            req.ClientID,
		UserID:              userID,
		RedirectURI:         req.RedirectURI,
		Scope:               req.Scope,
		CodeChallenge:       req.CodeChallenge,
		CodeChallengeMethod: req.CodeChallengeMethod,
		ExpiresAt:           time.Now().Add(10 * time.Minute),
	}

	if err := h.service.CreateAuthorizationCode(r.Context(), authCode); err != nil {
		log.Printf("AuthorizeHandler: Failed to create auth code: %v", err)
		writeOAuthError(w, "server_error", "Failed to create authorization code", http.StatusInternalServerError)
		return
	}

	redirectURL := req.RedirectURI + "?code=" + codeValue
	if req.State != "" {
		redirectURL += "&state=" + req.State
	}

	http.Redirect(w, r, redirectURL, http.StatusFound)
}

type RegisterClientRequest struct {
	Name         string   `json:"name"`
	RedirectURIs []string `json:"redirect_uris"`
	IsPublic     bool     `json:"is_public"`
}

type RegisterClientResponse struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret,omitempty"`
	Name         string   `json:"name"`
	RedirectURIs []string `json:"redirect_uris"`
	IsPublic     bool     `json:"is_public"`
}

func (h *AuthHandler) RegisterClientHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserIDFromToken(r)
	if err != nil || userID == "" {
		jsonutil.WriteErrorJSON(w, "Unauthorized")
		return
	}

	var req RegisterClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonutil.WriteErrorJSON(w, "Invalid request body")
		return
	}

	if req.Name == "" {
		jsonutil.WriteErrorJSON(w, "Client name required")
		return
	}

	if len(req.RedirectURIs) == 0 {
		jsonutil.WriteErrorJSON(w, "At least one redirect URI required")
		return
	}

	clientID, err := h.service.GenerateRandomString(16)
	if err != nil {
		jsonutil.WriteErrorJSON(w, "Failed to generate client ID")
		return
	}

	var clientSecret, clientSecretHash string
	if !req.IsPublic {
		clientSecret, err = h.service.GenerateRandomString(32)
		if err != nil {
			jsonutil.WriteErrorJSON(w, "Failed to generate client secret")
			return
		}
		clientSecretHash = h.service.HashString(clientSecret)
	}

	client := &domain.OAuthClient{
		ID:               clientID,
		ClientSecretHash: clientSecretHash,
		UserID:           userID,
		Name:             req.Name,
		IsPublic:         req.IsPublic,
	}

	if err := h.service.CreateOAuthClient(r.Context(), client); err != nil {
		log.Printf("RegisterClientHandler: Failed to create client: %v", err)
		jsonutil.WriteErrorJSON(w, "Failed to create client")
		return
	}

	for _, uri := range req.RedirectURIs {
		if err := h.service.AddRedirectURI(r.Context(), clientID, uri); err != nil {
			log.Printf("RegisterClientHandler: Failed to add redirect URI: %v", err)
		}
	}

	jsonutil.WriteJSON(w, http.StatusCreated, RegisterClientResponse{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Name:         req.Name,
		RedirectURIs: req.RedirectURIs,
		IsPublic:     req.IsPublic,
	})
}

type TokenIntrospectionResponse struct {
	Active    bool   `json:"active"`
	Scope     string `json:"scope,omitempty"`
	ClientID  string `json:"client_id,omitempty"`
	Username  string `json:"username,omitempty"`
	TokenType string `json:"token_type,omitempty"`
	Exp       int64  `json:"exp,omitempty"`
	Iat       int64  `json:"iat,omitempty"`
	Sub       string `json:"sub,omitempty"`
}

func (h *AuthHandler) TokenIntrospectionHandler(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("token")
	if token == "" {
		jsonutil.WriteJSON(w, http.StatusOK, TokenIntrospectionResponse{Active: false})
		return
	}

	oauthToken, err := h.service.ValidateOAuthToken(r.Context(), token)
	if err != nil || oauthToken == nil {
		jsonutil.WriteJSON(w, http.StatusOK, TokenIntrospectionResponse{Active: false})
		return
	}

	jsonutil.WriteJSON(w, http.StatusOK, TokenIntrospectionResponse{
		Active:    true,
		Scope:     oauthToken.Scope,
		ClientID:  oauthToken.ClientID,
		Sub:       oauthToken.UserID,
		TokenType: "Bearer",
		Exp:       oauthToken.ExpiresAt.Unix(),
		Iat:       oauthToken.CreatedAt.Unix(),
	})
}

type SSOCallbackRequest struct {
	Provider       string `json:"provider"`
	ProviderUserID string `json:"provider_user_id"`
	Email          string `json:"email"`
}

func (h *AuthHandler) SSOCallback(w http.ResponseWriter, r *http.Request) {
	var req SSOCallbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonutil.WriteErrorJSON(w, "Invalid request body")
		return
	}

	user, err := h.service.GetUserByExternalID(r.Context(), req.Provider, req.ProviderUserID)
	if err != nil {
		log.Printf("SSOCallback: DB error: %v", err)
		jsonutil.WriteErrorJSON(w, "Internal server error")
		return
	}

	if user == nil {
		user, err = h.service.CreateUser(r.Context(), req.Email, "SSO_MANAGED")
		if err != nil {
			log.Printf("SSOCallback: Failed to create user: %v", err)
			jsonutil.WriteErrorJSON(w, "Failed to provision user")
			return
		}

		if err := h.service.LinkExternalIdentity(r.Context(), user.ID, req.Provider, req.ProviderUserID); err != nil {
			log.Printf("SSOCallback: Failed to link identity: %v", err)
		}
	}

	token, err := jwtutil.GenerateToken(user.ID, user.Email)
	if err != nil {
		jsonutil.WriteErrorJSON(w, "Failed to generate token")
		return
	}

	jsonutil.WriteJSON(w, http.StatusOK, LoginResponse{Token: token})
}

func writeOAuthError(w http.ResponseWriter, errorCode, description string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error":             errorCode,
		"error_description": description,
	})
}
func (h *AuthHandler) TriggerEvent(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Type   string                 `json:"type"`
		ZoneID string                 `json:"zone_id"`
		Data   map[string]interface{} `json:"data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonutil.WriteErrorJSON(w, "Invalid request body")
		return
	}

	if req.Type == "" || req.ZoneID == "" {
		jsonutil.WriteErrorJSON(w, "Type and ZoneID are required")
		return
	}

	// Publish to Kafka (this assumes KafkaProducer is available in the handler)
	// For simplicity, we'll hit the event-sourcing bus directly
	log.Printf("Manually triggering event %s for zone %s", req.Type, req.ZoneID)

	// Implementation placeholder: in a real setup, we'd use h.kafkaProducer.Publish
	// For now, return success to indicate the API is wired
	jsonutil.WriteJSON(w, http.StatusAccepted, map[string]string{"status": "event_queued"})
}

// Password Reset Handlers

// ForgotPasswordRequest defines the payload for forgot password requests.
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// ForgotPassword handles password reset requests.
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonutil.WriteErrorJSON(w, "Method not allowed")
		return
	}

	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonutil.WriteErrorJSON(w, "Invalid request body")
		return
	}

	if req.Email == "" {
		jsonutil.WriteErrorJSON(w, "Email is required")
		return
	}

	// Get user by email
	user, err := h.service.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		log.Printf("ForgotPassword: Error getting user: %v", err)
		// Don't reveal if user exists
		jsonutil.WriteJSON(w, http.StatusOK, map[string]string{
			"message": "If an account with that email exists, a password reset link has been sent.",
		})
		return
	}

	if user == nil {
		// Don't reveal if user exists - return same message
		jsonutil.WriteJSON(w, http.StatusOK, map[string]string{
			"message": "If an account with that email exists, a password reset link has been sent.",
		})
		return
	}

	// Generate reset token
	token, err := h.service.CreatePasswordResetToken(r.Context(), user.ID)
	if err != nil {
		log.Printf("ForgotPassword: Error creating reset token: %v", err)
		jsonutil.WriteErrorJSON(w, "Failed to process request")
		return
	}

	// Log the token for testing purposes (in production, send via email)
	// Log the token for testing purposes (in production, send via email)
	log.Printf("Password reset token for %s: %s", user.Email, token)

	// Store in Redis for debug/testing
	if h.rdb != nil {
		h.rdb.Set(r.Context(), "debug:reset:"+user.Email, token, 1*time.Hour)
	}

	jsonutil.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "If an account with that email exists, a password reset link has been sent.",
	})
}

// ResetPasswordRequest defines the payload for password reset.
type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// ResetPassword handles password reset with token validation.
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonutil.WriteErrorJSON(w, "Method not allowed")
		return
	}

	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonutil.WriteErrorJSON(w, "Invalid request body")
		return
	}

	if req.Token == "" || req.NewPassword == "" {
		jsonutil.WriteErrorJSON(w, "Token and new password are required")
		return
	}

	if len(req.NewPassword) < 8 {
		jsonutil.WriteErrorJSON(w, "Password must be at least 8 characters")
		return
	}

	// Hash the new password
	b := &bcryptutil.BcryptUtilsImpl{}
	hashedPassword, err := b.GenerateHash(req.NewPassword)
	if err != nil {
		log.Printf("ResetPassword: Error hashing password: %v", err)
		jsonutil.WriteErrorJSON(w, "Failed to process request")
		return
	}

	// Reset the password
	if err := h.service.ResetPassword(r.Context(), req.Token, hashedPassword); err != nil {
		log.Printf("ResetPassword: Error resetting password: %v", err)
		jsonutil.WriteErrorJSON(w, "Invalid or expired reset token")
		return
	}

	jsonutil.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Password has been reset successfully.",
	})
}

// Email Verification Handlers

// VerifyEmailRequest defines the payload for email verification.
type VerifyEmailRequest struct {
	Token string `json:"token"`
}

// VerifyEmail handles email verification with token.
func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonutil.WriteErrorJSON(w, "Method not allowed")
		return
	}

	var req VerifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonutil.WriteErrorJSON(w, "Invalid request body")
		return
	}

	if req.Token == "" {
		jsonutil.WriteErrorJSON(w, "Verification token is required")
		return
	}

	if err := h.service.VerifyEmail(r.Context(), req.Token); err != nil {
		log.Printf("VerifyEmail: Error verifying email: %v", err)
		jsonutil.WriteErrorJSON(w, "Invalid or expired verification token")
		return
	}

	jsonutil.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Email has been verified successfully.",
	})
}

// ResendVerificationRequest defines the payload for resending verification.
type ResendVerificationRequest struct {
	Email string `json:"email"`
}

// ResendVerification handles resending email verification.
func (h *AuthHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonutil.WriteErrorJSON(w, "Method not allowed")
		return
	}

	var req ResendVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonutil.WriteErrorJSON(w, "Invalid request body")
		return
	}

	if req.Email == "" {
		jsonutil.WriteErrorJSON(w, "Email is required")
		return
	}

	user, err := h.service.GetUserByEmail(r.Context(), req.Email)
	if err != nil || user == nil {
		// Don't reveal if user exists
		jsonutil.WriteJSON(w, http.StatusOK, map[string]string{
			"message": "If an unverified account with that email exists, a verification email has been sent.",
		})
		return
	}

	if user.EmailVerified {
		jsonutil.WriteJSON(w, http.StatusOK, map[string]string{
			"message": "Email is already verified.",
		})
		return
	}

	// Generate verification token
	token, err := h.service.CreateEmailVerificationToken(r.Context(), user.ID)
	if err != nil {
		log.Printf("ResendVerification: Error creating verification token: %v", err)
		jsonutil.WriteErrorJSON(w, "Failed to process request")
		return
	}

	// Log the token for testing purposes (in production, send via email)
	log.Printf("Email verification token for %s: %s", user.Email, token)

	// Store in Redis for debug/testing
	if h.rdb != nil {
		h.rdb.Set(r.Context(), "debug:verify:"+user.Email, token, 1*time.Hour)
	}

	jsonutil.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "If an unverified account with that email exists, a verification email has been sent.",
	})
}

// DebugGetTokensRequest defines payload for debug token retrieval
type DebugGetTokensRequest struct {
	Email string `json:"email"`
	Type  string `json:"type"` // "verify" or "reset"
}

// DebugGetTokens returns the latest token for an email (Debug only)
func (h *AuthHandler) DebugGetTokens(w http.ResponseWriter, r *http.Request) {
	if h.rdb == nil {
		jsonutil.WriteErrorJSON(w, "Debug store not available")
		return
	}

	// Simple query param or body
	email := r.URL.Query().Get("email")
	tokenType := r.URL.Query().Get("type")

	if email == "" || tokenType == "" {
		jsonutil.WriteErrorJSON(w, "email and type required")
		return
	}

	key := "debug:" + tokenType + ":" + email
	token, err := h.rdb.Get(r.Context(), key).Result()
	if err == redis.Nil {
		jsonutil.WriteErrorJSON(w, "Token not found")
		return
	} else if err != nil {
		jsonutil.WriteErrorJSON(w, "Error fetching token")
		return
	}

	jsonutil.WriteJSON(w, http.StatusOK, map[string]string{
		"token": token,
	})
}
