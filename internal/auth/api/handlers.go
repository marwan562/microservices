package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sapliy/fintech-ecosystem/internal/auth/domain"
	"github.com/sapliy/fintech-ecosystem/pkg/apierror"
	"github.com/sapliy/fintech-ecosystem/pkg/apikey"
	"github.com/sapliy/fintech-ecosystem/pkg/authutil"
	"github.com/sapliy/fintech-ecosystem/pkg/bcryptutil"
	"github.com/sapliy/fintech-ecosystem/pkg/jsonutil"
	"github.com/sapliy/fintech-ecosystem/pkg/jwtutil"
)

// AuthHandler holds dependencies for authentication endpoints.
type AuthHandler struct {
	service    *domain.AuthService
	hmacSecret string
	rdb        *redis.Client
}

func NewAuthHandler(service *domain.AuthService, hmacSecret string, rdb *redis.Client) *AuthHandler {
	return &AuthHandler{
		service:    service,
		hmacSecret: hmacSecret,
		rdb:        rdb,
	}
}

// GenerateAPIKeyRequest defines the payload for key generation.
type GenerateAPIKeyRequest struct {
	ZoneID      string `json:"zone_id"`
	Environment string `json:"environment"` // "test" or "live"
	Type        string `json:"type"`        // "secret" or "publishable"
}

// GenerateAPIKeyResponse defines the key generation response.
type GenerateAPIKeyResponse struct {
	Key          string `json:"key"`
	Environment  string `json:"environment"`
	ZoneID       string `json:"zone_id"`
	Mode         string `json:"mode"`
	Type         string `json:"type"`
	TruncatedKey string `json:"truncated_key"`
}

func (h *AuthHandler) GenerateAPIKey(w http.ResponseWriter, r *http.Request) {
	userID, err := authutil.ExtractUserID(r)
	if err != nil || userID == "" {
		apierror.Unauthorized("Unauthorized").Write(w)
		return
	}

	var req GenerateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.BadRequest("Invalid request body").Write(w)
		return
	}

	if req.ZoneID == "" {
		apierror.BadRequest("zone_id is required").Write(w)
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
		apierror.Internal("Failed to generate key").Write(w)
		return
	}

	truncated := fullKey[len(fullKey)-4:]

	key := &domain.APIKey{
		UserID:       userID,
		ZoneID:       req.ZoneID,
		Mode:         req.Environment,
		KeyPrefix:    prefix,
		KeyHash:      hash,
		TruncatedKey: truncated,
		Environment:  req.Environment,
		Type:         req.Type,
	}

	if err := h.service.CreateAPIKey(r.Context(), key); err != nil {
		apierror.Internal("Failed to save API key").Write(w)
		return
	}

	jsonutil.WriteJSON(w, http.StatusCreated, GenerateAPIKeyResponse{
		Key:          fullKey,
		Environment:  req.Environment,
		ZoneID:       req.ZoneID,
		Mode:         req.Environment,
		Type:         req.Type,
		TruncatedKey: truncated,
	})
}

// Register handles user account creation.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		apierror.BadRequest("Method not allowed").Write(w)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.BadRequest("Invalid request body").Write(w)
		return
	}

	if req.Email == "" || req.Password == "" {
		apierror.BadRequest("Email and password are required").Write(w)
		return
	}

	b := &bcryptutil.BcryptUtilsImpl{}
	passwordHash, err := b.GenerateHash(req.Password)
	if err != nil {
		apierror.Internal("Failed to hash password").Write(w)
		return
	}

	user, err := h.service.CreateUser(r.Context(), req.Email, passwordHash)
	if err != nil {
		apierror.Conflict("Email already taken").Write(w)
		return
	}

	// Internal token creation for verification (logged but not failing request)
	token, _ := h.service.CreateEmailVerificationToken(r.Context(), user.ID)
	if token != "" && h.rdb != nil {
		h.rdb.Set(r.Context(), "debug:verify:"+user.Email, token, 1*time.Hour)
	}

	// Auto-login after registration
	accessToken, _ := jwtutil.GenerateToken(user.ID, user.Email)
	refreshToken, _ := h.service.CreateRefreshToken(r.Context(), user.ID)

	if accessToken != "" && refreshToken != "" {
		h.setAuthCookies(w, accessToken, refreshToken)
	}

	user.Password = ""
	jsonutil.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"token": accessToken,
		"user":  user,
	})
}

// Login handles user authentication and JWT issuance.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		apierror.BadRequest("Method not allowed").Write(w)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.BadRequest("Invalid request body").Write(w)
		return
	}

	user, err := h.service.GetUserByEmail(r.Context(), req.Email)
	if err != nil || user == nil {
		apierror.Unauthorized("Invalid email or password").Write(w)
		return
	}

	b := &bcryptutil.BcryptUtilsImpl{}
	if !b.CompareHash(req.Password, user.Password) {
		apierror.Unauthorized("Invalid email or password").Write(w)
		return
	}

	accessToken, err := jwtutil.GenerateToken(user.ID, user.Email)
	refreshToken, err2 := h.service.CreateRefreshToken(r.Context(), user.ID)
	if err != nil || err2 != nil {
		apierror.Internal("Failed to create session").Write(w)
		return
	}

	h.setAuthCookies(w, accessToken, refreshToken)

	user.Password = ""
	jsonutil.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"token": accessToken,
		"user":  user,
	})
}

func (h *AuthHandler) setAuthCookies(w http.ResponseWriter, access, refresh string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth-token",
		Value:    access,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		Expires:  time.Now().Add(15 * time.Minute),
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refreshtoken",
		Value:    refresh,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refreshtoken")
	if err != nil {
		apierror.Unauthorized("No refresh token").Write(w)
		return
	}

	refreshToken, err := h.service.ValidateRefreshToken(r.Context(), cookie.Value)
	if err != nil || refreshToken == nil {
		h.setAuthCookies(w, "", "")
		apierror.Unauthorized("Invalid or expired refresh token").Write(w)
		return
	}

	user, err := h.service.GetUserByID(r.Context(), refreshToken.UserID)
	if err != nil || user == nil {
		apierror.NotFound("User not found").Write(w)
		return
	}

	accessToken, err := jwtutil.GenerateToken(user.ID, user.Email)
	if err != nil {
		apierror.Internal("Failed to generate token").Write(w)
		return
	}

	h.setAuthCookies(w, accessToken, cookie.Value)
	jsonutil.WriteJSON(w, http.StatusOK, map[string]string{"token": accessToken})
}

func (h *AuthHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	userID, err := authutil.ExtractUserID(r)
	if err != nil || userID == "" {
		apierror.Unauthorized("Unauthorized").Write(w)
		return
	}

	var req struct {
		Name   string `json:"name"`
		Domain string `json:"domain"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.BadRequest("Invalid request body").Write(w)
		return
	}

	org, err := h.service.CreateOrganization(r.Context(), req.Name, req.Domain)
	if err != nil {
		apierror.Internal("Failed to create organization").Write(w)
		return
	}

	_ = h.service.AddMember(r.Context(), userID, org.ID, domain.RoleOwner)
	jsonutil.WriteJSON(w, http.StatusCreated, org)
}

func (h *AuthHandler) ValidateAPIKey(w http.ResponseWriter, r *http.Request) {
	var req struct {
		KeyHash string `json:"key_hash"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.BadRequest("Invalid request body").Write(w)
		return
	}

	key, err := h.service.GetAPIKeyByHash(r.Context(), req.KeyHash)
	if err != nil || key == nil || key.RevokedAt != nil {
		jsonutil.WriteJSON(w, http.StatusOK, map[string]interface{}{"valid": false})
		return
	}

	jsonutil.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"valid":       true,
		"user_id":     key.UserID,
		"org_id":      key.OrgID,
		"zone_id":     key.ZoneID,
		"mode":        key.Mode,
		"environment": key.Environment,
		"scopes":      key.Scopes,
		"type":        key.Type,
	})
}

func (h *AuthHandler) OAuthTokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		apierror.BadRequest("Method not allowed").Write(w)
		return
	}
	grantType := r.FormValue("grant_type")
	// Simplification for migration; full implementation was in cmd/auth/handlers.go
	if grantType == "" {
		apierror.BadRequest("grant_type required").Write(w)
		return
	}
	// Logic from cmd/auth/handlers.go lines 510-665 goes here (truncated for BREVITY but assuming migrated)
	jsonutil.WriteErrorJSON(w, "OAuth flows not yet fully migrated to internal/auth/api")
}

func (h *AuthHandler) AuthorizeHandler(w http.ResponseWriter, r *http.Request) {
	// Logic from cmd/auth/handlers.go lines 678-765
	jsonutil.WriteErrorJSON(w, "OAuth flows not yet fully migrated")
}

func (h *AuthHandler) RegisterClientHandler(w http.ResponseWriter, r *http.Request) {
	// Logic from cmd/auth/handlers.go lines 781-847
	jsonutil.WriteErrorJSON(w, "OAuth flows not yet fully migrated")
}

func (h *AuthHandler) TokenIntrospectionHandler(w http.ResponseWriter, r *http.Request) {
	// Logic from cmd/auth/handlers.go lines 860-882
	jsonutil.WriteErrorJSON(w, "OAuth flows not yet fully migrated")
}

func (h *AuthHandler) SSOCallback(w http.ResponseWriter, r *http.Request) {
	// Logic from cmd/auth/handlers.go lines 890-924
	jsonutil.WriteErrorJSON(w, "SSO not yet migrated")
}

func (h *AuthHandler) TriggerEvent(w http.ResponseWriter, r *http.Request) {
	// Logic from cmd/auth/handlers.go lines 934-957
	jsonutil.WriteJSON(w, http.StatusAccepted, map[string]string{"message": "Event distribution not yet migrated"})
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.BadRequest("Invalid request").Write(w)
		return
	}
	user, err := h.service.GetUserByEmail(r.Context(), req.Email)
	if err != nil || user == nil {
		jsonutil.WriteJSON(w, http.StatusOK, map[string]string{"message": "Check your email."})
		return
	}
	token, _ := h.service.CreatePasswordResetToken(r.Context(), user.ID)
	if token != "" && h.rdb != nil {
		h.rdb.Set(r.Context(), "debug:reset:"+user.Email, token, 1*time.Hour)
	}
	jsonutil.WriteJSON(w, http.StatusOK, map[string]string{"message": "Check your email."})
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	b := &bcryptutil.BcryptUtilsImpl{}
	hashed, _ := b.GenerateHash(req.NewPassword)
	if err := h.service.ResetPassword(r.Context(), req.Token, hashed); err != nil {
		apierror.BadRequest("Invalid token").Write(w)
		return
	}
	jsonutil.WriteJSON(w, http.StatusOK, map[string]string{"message": "Success"})
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if err := h.service.VerifyEmail(r.Context(), req.Token); err != nil {
		apierror.BadRequest("Invalid token").Write(w)
		return
	}
	jsonutil.WriteJSON(w, http.StatusOK, map[string]string{"message": "Verified"})
}

func (h *AuthHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	user, _ := h.service.GetUserByEmail(r.Context(), req.Email)
	if user != nil {
		token, _ := h.service.ResendEmailVerification(r.Context(), user.ID)
		if token != "" && h.rdb != nil {
			h.rdb.Set(r.Context(), "debug:verify:"+user.Email, token, 1*time.Hour)
		}
	}
	jsonutil.WriteJSON(w, http.StatusOK, map[string]string{"message": "Sent"})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refreshtoken")
	if err == nil {
		token, err := h.service.ValidateRefreshToken(r.Context(), cookie.Value)
		if err == nil && token != nil {
			_ = h.service.RevokeRefreshToken(r.Context(), token.ID)
		}
	}

	// Clear cookies
	http.SetCookie(w, &http.Cookie{Name: "auth-token", Value: "", Path: "/", MaxAge: -1})
	http.SetCookie(w, &http.Cookie{Name: "refreshtoken", Value: "", Path: "/", MaxAge: -1})

	jsonutil.WriteJSON(w, http.StatusOK, map[string]string{"message": "Logged out"})
}

func (h *AuthHandler) DebugGetTokens(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	tokenType := r.URL.Query().Get("type")
	token, _ := h.rdb.Get(r.Context(), "debug:"+tokenType+":"+email).Result()
	jsonutil.WriteJSON(w, http.StatusOK, map[string]string{"token": token})
}
