package audit

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// SignedAuditLog extends AuditLog with cryptographic signature for tamper detection.
type SignedAuditLog struct {
	AuditLog

	// ZoneID is the tenant/zone this log belongs to.
	ZoneID string `json:"zone_id"`

	// IPAddress is the client IP.
	IPAddress string `json:"ip_address,omitempty"`

	// UserAgent is the client user agent.
	UserAgent string `json:"user_agent,omitempty"`

	// PreviousHash links to the previous log entry (blockchain-style chaining).
	PreviousHash string `json:"previous_hash,omitempty"`

	// Signature is the HMAC-SHA256 signature of the log entry.
	Signature string `json:"signature"`

	// SignedAt is when the signature was generated.
	SignedAt time.Time `json:"signed_at"`
}

// SecureLogger provides tamper-evident audit logging.
type SecureLogger struct {
	db         *sql.DB
	signingKey []byte
	lastHash   string
}

// NewSecureLogger creates a secure audit logger.
func NewSecureLogger(db *sql.DB, signingKey []byte) (*SecureLogger, error) {
	if len(signingKey) < 32 {
		return nil, fmt.Errorf("signing key must be at least 32 bytes")
	}

	logger := &SecureLogger{
		db:         db,
		signingKey: signingKey,
	}

	// Load the last hash from the database for chaining
	if err := logger.loadLastHash(); err != nil {
		// Not fatal, just means this is the first entry
		logger.lastHash = "genesis"
	}

	return logger, nil
}

// Log creates a signed audit log entry.
func (l *SecureLogger) Log(entry SignedAuditLog) error {
	entry.Timestamp = time.Now()
	entry.SignedAt = entry.Timestamp
	entry.PreviousHash = l.lastHash

	// Generate signature
	sig, err := l.sign(entry)
	if err != nil {
		return fmt.Errorf("failed to sign audit log: %w", err)
	}
	entry.Signature = sig

	// Store in database
	if err := l.store(entry); err != nil {
		return fmt.Errorf("failed to store audit log: %w", err)
	}

	// Update chain
	l.lastHash = sig

	return nil
}

// sign generates an HMAC-SHA256 signature of the log entry.
func (l *SecureLogger) sign(entry SignedAuditLog) (string, error) {
	// Create canonical representation for signing
	canonical := struct {
		Timestamp    int64                  `json:"t"`
		ActorID      string                 `json:"a"`
		ZoneID       string                 `json:"z"`
		Action       string                 `json:"act"`
		ResourceType string                 `json:"rt"`
		ResourceID   string                 `json:"rid"`
		Metadata     map[string]interface{} `json:"m"`
		PreviousHash string                 `json:"ph"`
	}{
		Timestamp:    entry.Timestamp.UnixNano(),
		ActorID:      entry.ActorID,
		ZoneID:       entry.ZoneID,
		Action:       entry.Action,
		ResourceType: entry.ResourceType,
		ResourceID:   entry.ResourceID,
		Metadata:     entry.Metadata,
		PreviousHash: entry.PreviousHash,
	}

	data, err := json.Marshal(canonical)
	if err != nil {
		return "", err
	}

	mac := hmac.New(sha256.New, l.signingKey)
	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil)), nil
}

// Verify checks if an audit log entry has a valid signature.
func (l *SecureLogger) Verify(entry SignedAuditLog) (bool, error) {
	expectedSig, err := l.sign(entry)
	if err != nil {
		return false, err
	}

	return hmac.Equal([]byte(entry.Signature), []byte(expectedSig)), nil
}

// VerifyChain verifies the integrity of the entire audit log chain.
func (l *SecureLogger) VerifyChain(zoneID string, limit int) (bool, []string, error) {
	logs, err := l.getRecentLogs(zoneID, limit)
	if err != nil {
		return false, nil, err
	}

	var errors []string
	previousHash := ""

	for i := len(logs) - 1; i >= 0; i-- {
		log := logs[i]

		// Verify signature
		valid, err := l.Verify(log)
		if err != nil {
			errors = append(errors, fmt.Sprintf("log %s: verification error: %v", log.ID, err))
			continue
		}
		if !valid {
			errors = append(errors, fmt.Sprintf("log %s: invalid signature", log.ID))
		}

		// Verify chain link
		if previousHash != "" && log.PreviousHash != previousHash {
			errors = append(errors, fmt.Sprintf("log %s: chain broken", log.ID))
		}

		previousHash = log.Signature
	}

	return len(errors) == 0, errors, nil
}

func (l *SecureLogger) store(entry SignedAuditLog) error {
	meta, _ := json.Marshal(entry.Metadata)

	query := `
		INSERT INTO audit_logs (
			id, zone_id, timestamp, actor_id, action,
			resource_type, resource_id, metadata,
			ip_address, user_agent, previous_hash, signature
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := l.db.Exec(query,
		entry.ID, entry.ZoneID, entry.Timestamp, entry.ActorID, entry.Action,
		entry.ResourceType, entry.ResourceID, string(meta),
		entry.IPAddress, entry.UserAgent, entry.PreviousHash, entry.Signature,
	)

	return err
}

func (l *SecureLogger) loadLastHash() error {
	var hash string
	err := l.db.QueryRow(`
		SELECT signature FROM audit_logs 
		ORDER BY timestamp DESC LIMIT 1
	`).Scan(&hash)
	if err != nil {
		return err
	}
	l.lastHash = hash
	return nil
}

func (l *SecureLogger) getRecentLogs(zoneID string, limit int) ([]SignedAuditLog, error) {
	rows, err := l.db.Query(`
		SELECT id, zone_id, timestamp, actor_id, action,
			resource_type, resource_id, metadata,
			ip_address, user_agent, previous_hash, signature
		FROM audit_logs
		WHERE zone_id = $1
		ORDER BY timestamp DESC
		LIMIT $2
	`, zoneID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []SignedAuditLog
	for rows.Next() {
		var log SignedAuditLog
		var meta string
		err := rows.Scan(
			&log.ID, &log.ZoneID, &log.Timestamp, &log.ActorID, &log.Action,
			&log.ResourceType, &log.ResourceID, &meta,
			&log.IPAddress, &log.UserAgent, &log.PreviousHash, &log.Signature,
		)
		if err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(meta), &log.Metadata)
		logs = append(logs, log)
	}

	return logs, nil
}

// Common audit actions
const (
	ActionUserLogin      = "user.login"
	ActionUserLogout     = "user.logout"
	ActionUserCreate     = "user.create"
	ActionUserUpdate     = "user.update"
	ActionUserDelete     = "user.delete"
	ActionAPIKeyCreate   = "apikey.create"
	ActionAPIKeyRevoke   = "apikey.revoke"
	ActionFlowCreate     = "flow.create"
	ActionFlowUpdate     = "flow.update"
	ActionFlowDelete     = "flow.delete"
	ActionFlowExecute    = "flow.execute"
	ActionPaymentCreate  = "payment.create"
	ActionPaymentCapture = "payment.capture"
	ActionPaymentRefund  = "payment.refund"
	ActionWebhookCreate  = "webhook.create"
	ActionWebhookUpdate  = "webhook.update"
	ActionWebhookDelete  = "webhook.delete"
	ActionSecretAccess   = "secret.access"
	ActionSecretRotate   = "secret.rotate"
	ActionRoleAssign     = "role.assign"
	ActionRoleRevoke     = "role.revoke"
	ActionExportData     = "data.export"
)
