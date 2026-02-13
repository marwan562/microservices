package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/sapliy/fintech-ecosystem/pkg/apikey"
	"github.com/sapliy/fintech-ecosystem/pkg/jsonutil"
	"github.com/sapliy/fintech-ecosystem/pkg/observability"
	"github.com/sapliy/fintech-ecosystem/pkg/scopes"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	pb "github.com/sapliy/fintech-ecosystem/proto/auth"
	walletpb "github.com/sapliy/fintech-ecosystem/proto/wallet"

	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/redis/go-redis/v9"
	"github.com/sapliy/fintech-ecosystem/pkg/monitoring"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	GatewayRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gateway_requests_total",
		Help: "Total number of requests handled by the gateway.",
	}, []string{"method", "path", "status"})
)

// GatewayHandler holds the configuration for upstream service URLs and Redis.
type GatewayHandler struct {
	authServiceURL         string
	paymentServiceURL      string
	ledgerServiceURL       string
	walletServiceURL       string
	billingServiceURL      string
	eventsServiceURL       string
	flowServiceURL         string
	notificationServiceURL string
	rdb                    *redis.Client
	upgrader               websocket.Upgrader
	authClient             pb.AuthServiceClient
	walletClient           walletpb.WalletServiceClient
	hmacSecret             string
	logger                 *observability.Logger
}

// NewGatewayHandler creates a new instance of GatewayHandler.
func NewGatewayHandler(auth, payment, ledger, wallet, billing, events, flow, notification string, rdb *redis.Client, authClient pb.AuthServiceClient, walletClient walletpb.WalletServiceClient, hmacSecret string, logger *observability.Logger) *GatewayHandler {
	return &GatewayHandler{
		authServiceURL:         auth,
		paymentServiceURL:      payment,
		ledgerServiceURL:       ledger,
		walletServiceURL:       wallet,
		billingServiceURL:      billing,
		eventsServiceURL:       events,
		flowServiceURL:         flow,
		notificationServiceURL: notification,
		rdb:                    rdb,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		authClient:   authClient,
		walletClient: walletClient,
		hmacSecret:   hmacSecret,
		logger:       logger,
	}
}

// validateKeyWithAuthService calls the Auth service to validate the API key hash.
func (h *GatewayHandler) validateKeyWithAuthService(ctx context.Context, keyHash string) (string, string, string, string, string, int32, string, string, string, bool) {
	res, err := h.authClient.ValidateKey(ctx, &pb.ValidateKeyRequest{KeyHash: keyHash})
	if err != nil {
		h.logger.Error("Auth service gRPC validation call failed", "error", err)
		return "", "", "", "", "", 0, "", "", "", false
	}

	return res.UserId, res.Environment, res.Scopes, res.OrgId, res.Role, res.RateLimitQuota, res.ZoneId, res.Mode, res.KeyType, res.Valid
}

// checkRateLimit checks if the key has exceeded its quota.
func (h *GatewayHandler) checkRateLimit(ctx context.Context, keyHash string, quota int32) (bool, error) {
	if quota <= 0 {
		quota = 100 // Default fallback
	}
	window := time.Now().Unix() / 60
	redisKey := fmt.Sprintf("rate_limit:%s:%d", keyHash, window)

	count, err := h.rdb.Incr(ctx, redisKey).Result()
	if err != nil {
		return false, err
	}

	if count == 1 {
		h.rdb.Expire(ctx, redisKey, 60*time.Second)
	}

	return count <= int64(quota), nil
}

// proxyRequest creates a reverse proxy to the target URL and serves the request.
func (h *GatewayHandler) proxyRequest(target string, w http.ResponseWriter, r *http.Request) {
	targetURL, err := url.Parse(target)
	if err != nil {
		h.logger.Error("Error parsing target URL", "target", target, "error", err)
		jsonutil.WriteErrorJSON(w, "Internal Server Error; Invalid Target")
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = targetURL.Host
		// Ensure headers injected by middleware persist
		if userID := r.Header.Get("X-User-ID"); userID != "" {
			req.Header.Set("X-User-ID", userID)
		}
		if env := r.Header.Get("X-Environment"); env != "" {
			req.Header.Set("X-Environment", env)
		}
		if orgID := r.Header.Get("X-Org-ID"); orgID != "" {
			req.Header.Set("X-Org-ID", orgID)
		}
		if zoneID := r.Header.Get("X-Zone-ID"); zoneID != "" {
			req.Header.Set("X-Zone-ID", zoneID)
		}
		if mode := r.Header.Get("X-Zone-Mode"); mode != "" {
			req.Header.Set("X-Zone-Mode", mode)
		}
	}

	proxy.ServeHTTP(w, r)
}

// ServeHTTP implements the http.Handler interface with Middleware.
func (h *GatewayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	h.logger.Info("Incoming request", "method", r.Method, "path", path)

	if strings.HasPrefix(path, "/auth") || path == "/health" {
		h.logger.Debug("Routing public path", "path", path)
		h.routePublic(w, r)
		return
	}

	// Protected Endpoints (API Key Required)
	// Extract Secret Key
	authHeader := r.Header.Get("Authorization")
	apiKey := ""
	if strings.HasPrefix(authHeader, "Bearer ") {
		apiKey = strings.TrimPrefix(authHeader, "Bearer ")
	} else {
		// Fallback to query parameter for WebSockets or simple GETs
		apiKey = r.URL.Query().Get("api_key")
	}

	if apiKey == "" || (!strings.HasPrefix(apiKey, "sk_") && !strings.HasPrefix(apiKey, "pk_")) {
		jsonutil.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "Missing or invalid API Key"})
		return
	}

	// Hash Key
	keyHash := apikey.HashKey(apiKey, h.hmacSecret)

	// Validate with Auth Service
	userID, env, keyScopes, orgID, role, quota, zoneID, mode, keyType, valid := h.validateKeyWithAuthService(r.Context(), keyHash)
	if !valid {
		jsonutil.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid or revoked API Key"})
		return
	}

	// Key Type Enforcement (Example: pk_ keys can only emit events)
	if keyType == "publishable" && !strings.HasPrefix(path, "/v1/events/emit") {
		jsonutil.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "Publishable keys only allowed for event emission"})
		return
	}

	// Scope Enforcement
	requiredScope := scopes.GetRequiredScope(path, r.Method)
	if requiredScope != "" && !scopes.HasScope(keyScopes, requiredScope) {
		jsonutil.WriteJSON(w, http.StatusForbidden, map[string]string{
			"error":          "Insufficient scope",
			"required_scope": requiredScope,
		})
		return
	}

	// Rate Limiting
	allowed, err := h.checkRateLimit(r.Context(), keyHash, quota)
	if err != nil {
		h.logger.Error("Redis error in rate limiter", "error", err)
		// Fail open or closed? Closed for security.
		jsonutil.WriteErrorJSON(w, "Internal Server Error")
		return
	}
	if !allowed {
		w.Header().Set("Retry-After", "60")
		jsonutil.WriteJSON(w, http.StatusTooManyRequests, map[string]string{"error": "Rate limit exceeded"})
		return
	}

	// Inject Context
	r.Header.Set("X-User-ID", userID)
	r.Header.Set("X-Environment", env)
	r.Header.Set("X-Org-ID", orgID)
	r.Header.Set("X-Role", role)
	r.Header.Set("X-Zone-ID", zoneID)
	r.Header.Set("X-Zone-Mode", mode)

	// Route to Service
	// Handle /v1 prefix by optional stripping
	p := path
	if after, ok := strings.CutPrefix(p, "/v1"); ok {
		p = after
	}

	switch {
	case strings.HasPrefix(p, "/payments"):
		http.StripPrefix(path[:len(path)-len(p)]+"/payments", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.proxyRequest(h.paymentServiceURL, w, r)
		})).ServeHTTP(w, r)

	case strings.HasPrefix(p, "/ledger"):
		http.StripPrefix(path[:len(path)-len(p)]+"/ledger", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.proxyRequest(h.ledgerServiceURL, w, r)
		})).ServeHTTP(w, r)

	case strings.HasPrefix(p, "/wallets"):
		h.proxyRequest(h.walletServiceURL, w, r)

	case strings.HasPrefix(p, "/billing"):
		http.StripPrefix(path[:len(path)-len(p)]+"/billing", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.proxyRequest(h.billingServiceURL, w, r)
		})).ServeHTTP(w, r)

	case strings.HasPrefix(p, "/webhooks") || strings.HasPrefix(p, "/notifications"):
		h.proxyRequest(h.notificationServiceURL, w, r)

	case strings.HasPrefix(p, "/events"):
		// Use internal WS handler for events stream, or proxy for others
		if p == "/events/stream" && websocket.IsWebSocketUpgrade(r) {
			h.handleWebSocket(w, r)
			return
		}
		if p == "/events/emit" && r.Method == http.MethodPost {
			h.handleEventEmit(w, r)
			return
		}
		h.proxyRequest(h.eventsServiceURL, w, r)

	case strings.HasPrefix(p, "/flows") || strings.HasPrefix(p, "/executions"):
		h.proxyRequest(h.flowServiceURL, w, r)

	case p == "/ws": // Legacy or alternative WS path
		if websocket.IsWebSocketUpgrade(r) {
			h.handleWebSocket(w, r)
			return
		}
		jsonutil.WriteErrorJSON(w, "WebSocket upgrade required")

	default:
		// Fallback for root path if it's a WebSocket upgrade
		if (p == "/" || p == "") && websocket.IsWebSocketUpgrade(r) {
			h.handleWebSocket(w, r)
			return
		}
		h.logger.Warn("Route not found", "path", path)
		jsonutil.WriteErrorJSON(w, "Not Found")
	}
}

func (h *GatewayHandler) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("WS upgrade failed", "error", err)
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			h.logger.Warn("Failed to close WS connection", "error", err)
		}
	}()

	pubsub := h.rdb.Subscribe(r.Context(), "webhook_events")
	defer func() {
		if err := pubsub.Close(); err != nil {
			h.logger.Warn("Failed to close Redis PubSub", "error", err)
		}
	}()

	ch := pubsub.Channel()
	for msg := range ch {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
			h.logger.Error("WS write failed", "error", err)
			break
		}
	}
}

// EventEnvelope is the full event structure following the spec
type EventEnvelope struct {
	ID             string                 `json:"id"`
	Type           string                 `json:"type"`
	ZoneID         string                 `json:"zone_id"`
	OrgID          string                 `json:"org_id"`
	Timestamp      string                 `json:"timestamp"`
	IdempotencyKey string                 `json:"idempotency_key"`
	Payload        map[string]interface{} `json:"payload"`
	Meta           map[string]string      `json:"meta"`
}

type EventEmitRequest struct {
	Type           string                 `json:"type"`
	IdempotencyKey string                 `json:"idempotency_key,omitempty"`
	Data           map[string]interface{} `json:"data"`
	Meta           map[string]string      `json:"meta,omitempty"`
}

type EventEmitResponse struct {
	Status  string `json:"status"`
	EventID string `json:"event_id,omitempty"`
	Topic   string `json:"topic,omitempty"`
	Message string `json:"message,omitempty"`
}

// generateEventID creates a unique event ID
func generateEventID() string {
	return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}

// hashPayload creates a SHA256 hash of the payload for dedup verification
func hashPayload(data map[string]interface{}) string {
	bytes, _ := json.Marshal(data)
	hash := sha256.Sum256(bytes)
	return fmt.Sprintf("%x", hash[:8])
}

func (h *GatewayHandler) handleEventEmit(w http.ResponseWriter, r *http.Request) {
	zoneID := r.Header.Get("X-Zone-ID")
	orgID := r.Header.Get("X-Org-ID")
	if zoneID == "" {
		jsonutil.WriteErrorJSON(w, "Zone context missing")
		return
	}

	var req EventEmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonutil.WriteErrorJSON(w, "Invalid request body")
		return
	}

	if req.Type == "" {
		jsonutil.WriteErrorJSON(w, "Event type required")
		return
	}

	// Generate or use provided idempotency key
	idempotencyKey := req.IdempotencyKey
	if idempotencyKey == "" {
		idempotencyKey = r.Header.Get("Idempotency-Key")
	}
	if idempotencyKey == "" {
		// Auto-generate but warn (non-idempotent)
		idempotencyKey = fmt.Sprintf("auto_%d", time.Now().UnixNano())
	}

	// Build dedup key: zone_id:idempotency_key
	dedupKey := fmt.Sprintf("dedup:%s:%s", zoneID, idempotencyKey)
	payloadHash := hashPayload(req.Data)

	// Check for existing idempotency record
	cachedResult, err := h.rdb.Get(r.Context(), dedupKey).Result()
	if err == nil && cachedResult != "" {
		// Found cached result - parse and return
		var cached EventEmitResponse
		if json.Unmarshal([]byte(cachedResult), &cached) == nil {
			cached.Status = "duplicate"
			cached.Message = "Event already processed"
			jsonutil.WriteJSON(w, http.StatusAccepted, cached)
			return
		}
	}

	// Generate event envelope
	eventID := generateEventID()
	envelope := EventEnvelope{
		ID:             eventID,
		Type:           req.Type,
		ZoneID:         zoneID,
		OrgID:          orgID,
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		IdempotencyKey: idempotencyKey,
		Payload:        req.Data,
		Meta: map[string]string{
			"source":       req.Meta["source"],
			"env":          req.Meta["env"],
			"payload_hash": payloadHash,
		},
	}

	// Set default meta values
	if envelope.Meta["source"] == "" {
		envelope.Meta["source"] = "gateway"
	}

	topic := fmt.Sprintf("zone.%s.event.%s", zoneID, req.Type)
	envelopeBytes, _ := json.Marshal(envelope)

	// Publish to Redis Stream
	err = h.rdb.XAdd(r.Context(), &redis.XAddArgs{
		Stream: topic,
		Values: map[string]interface{}{
			"envelope": envelopeBytes,
			"data":     envelopeBytes, // For backward compatibility
			"ts":       time.Now().Unix(),
		},
	}).Err()

	if err != nil {
		h.logger.Error("Failed to publish to Redis Stream", "error", err)
		jsonutil.WriteErrorJSON(w, "Failed to ingest event")
		return
	}

	// Build response
	response := EventEmitResponse{
		Status:  "ingested",
		EventID: eventID,
		Topic:   topic,
	}

	// Cache result for dedup (30 days TTL)
	resultBytes, _ := json.Marshal(response)
	h.rdb.Set(r.Context(), dedupKey, resultBytes, 30*24*time.Hour)

	h.logger.Info("Event ingested", "event_id", eventID, "topic", topic, "zone", zoneID)
	jsonutil.WriteJSON(w, http.StatusAccepted, response)
}

func (h *GatewayHandler) routePublic(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if strings.HasPrefix(path, "/auth") {
		http.StripPrefix("/auth", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.proxyRequest(h.authServiceURL, w, r)
		})).ServeHTTP(w, r)
		return
	}
	// Health
	jsonutil.WriteJSON(w, http.StatusOK, map[string]string{
		"status":  "active",
		"service": "gateway",
		"date":    time.Now().Format(time.DateTime),
	})
}

// CORSMiddleware adds Cross-Origin Resource Sharing headers so the
// Next.js frontend (and other allowed origins) can call the gateway.
func CORSMiddleware(allowedOrigins string, next http.Handler) http.Handler {
	origins := strings.Split(allowedOrigins, ",")
	allowed := make(map[string]bool, len(origins))
	for _, o := range origins {
		allowed[strings.TrimSpace(o)] = true
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Set CORS headers for matching origins
		if origin != "" && allowed[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else if allowed["*"] {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else if origin != "" {
			// For preflight from any origin, still set the header so the
			// browser sees a valid CORS response. In production, tighten this.
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Idempotency-Key, X-Zone-ID")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Handle preflight — must return before reaching auth/routing logic
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	logger := observability.NewLogger("gateway-service")

	// configuration
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	authURL := os.Getenv("AUTH_SERVICE_URL")
	if authURL == "" {
		authURL = "http://127.0.0.1:8081"
	}

	paymentURL := os.Getenv("PAYMENT_SERVICE_URL")
	if paymentURL == "" {
		paymentURL = "http://127.0.0.1:8082"
	}

	ledgerURL := os.Getenv("LEDGER_SERVICE_URL")
	if ledgerURL == "" {
		ledgerURL = "http://127.0.0.1:8083"
	}

	walletURL := os.Getenv("WALLET_SERVICE_URL")
	if walletURL == "" {
		walletURL = "http://127.0.0.1:8085"
	}

	notificationURL := os.Getenv("NOTIFICATION_SERVICE_URL")
	if notificationURL == "" {
		notificationURL = "http://127.0.0.1:8084"
	}

	eventsURL := os.Getenv("EVENTS_SERVICE_URL")
	if eventsURL == "" {
		eventsURL = "http://127.0.0.1:8089"
	}

	flowURL := os.Getenv("FLOW_SERVICE_URL")
	if flowURL == "" {
		flowURL = "http://127.0.0.1:8088"
	}

	billingURL := os.Getenv("BILLING_SERVICE_URL")
	if billingURL == "" {
		billingURL = "http://127.0.0.1:8089" // Assuming a port for billing REST if added, or proxying gRPC?
		// Note: Billing seems gRPC only currently, but Gateway proxies HTTP.
		// For now, I'll point it to where billing might listen or keep it for future REST expansion.
		// If Billing has no REST handlers, this will 404/502 which is expected for now.
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	// Ping Redis
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		logger.Warn("Redis connection failed", "error", err)
	} else {
		logger.Info("Redis connection established")
	}

	// Setup Auth Service gRPC Client
	authGRPCAddr := os.Getenv("AUTH_GRPC_ADDR")
	if authGRPCAddr == "" {
		authGRPCAddr = "localhost:50051"
	}
	conn, err := grpc.NewClient(authGRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(monitoring.UnaryClientInterceptor("gateway")),
	)
	if err != nil {
		logger.Error("did not connect to auth gRPC", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			logger.Error("Failed to close gRPC connection", "error", err)
		}
	}()
	authClient := pb.NewAuthServiceClient(conn)

	// Setup Wallet Service gRPC Client
	walletGRPCAddr := os.Getenv("WALLET_GRPC_ADDR")
	if walletGRPCAddr == "" {
		walletGRPCAddr = "localhost:50053"
	}
	connWallet, err := grpc.NewClient(walletGRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(monitoring.UnaryClientInterceptor("gateway")),
	)
	if err != nil {
		logger.Error("did not connect to wallet gRPC", "error", err)
		os.Exit(1)
	}
	defer func() { _ = connWallet.Close() }()
	walletClient := walletpb.NewWalletServiceClient(connWallet)

	// Initialize Tracer
	shutdown, err := observability.InitTracer(context.Background(), observability.Config{
		ServiceName:    "gateway",
		ServiceVersion: "0.1.0",
		Endpoint:       os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		Environment:    "production",
	})
	if err != nil {
		logger.Error("Failed to init tracer", "error", err)
	} else {
		defer func() {
			if err := shutdown(context.Background()); err != nil {
				logger.Error("Failed to shutdown tracer", "error", err)
			}
		}()
	}

	// Start Metrics Server
	monitoring.StartMetricsServer(":8087")

	// HMAC Secret
	hmacSecret := os.Getenv("API_KEY_HMAC_SECRET")
	if hmacSecret == "" {
		hmacSecret = "local-dev-secret-do-not-use-in-prod"
		logger.Warn("API_KEY_HMAC_SECRET not set, using default for dev")
	}

	gateway := NewGatewayHandler(authURL, paymentURL, ledgerURL, walletURL, billingURL, eventsURL, flowURL, notificationURL, rdb, authClient, walletClient, hmacSecret, logger)

	// CORS configuration
	corsOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if corsOrigins == "" {
		corsOrigins = "http://localhost:3000"
		logger.Info("CORS_ALLOWED_ORIGINS not set, defaulting to localhost:3000")
	}

	// Wrap handler with CORS, OpenTelemetry and Prometheus
	corsHandler := CORSMiddleware(corsOrigins, gateway)
	otelHandler := otelhttp.NewHandler(corsHandler, "gateway-request")
	promHandler := monitoring.PrometheusMiddleware(otelHandler)

	server := &http.Server{
		Addr:    ":8080",
		Handler: promHandler,
	}

	logger.Info("Gateway service starting", "port", ":8080")
	if err := server.ListenAndServe(); err != nil {
		logger.Error("Gateway server failed", "error", err)
		os.Exit(1)
	}
}
