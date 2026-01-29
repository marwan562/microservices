package audit

import (
	"context"
	"encoding/json"
	"log"
	"time"
)

type AuditLog struct {
	ID           string                 `json:"id"`
	Timestamp    time.Time              `json:"timestamp"`
	ActorID      string                 `json:"actor_id"`
	OrgID        string                 `json:"org_id,omitempty"`
	Action       string                 `json:"action"`        // e.g. "payment.refund", "user.login"
	ResourceType string                 `json:"resource_type"` // e.g. "payment_intent"
	ResourceID   string                 `json:"resource_id"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// Log records an audit event.
// In a real system, this would write to a dedicated DB or message bus.
func Log(ctx context.Context, entry AuditLog) {
	entry.Timestamp = time.Now()

	// For now, we log to stdout.
	// Future: Write to 'audit_logs' table or Kafka.
	meta, _ := json.Marshal(entry.Metadata)
	log.Printf("[AUDIT] %s | Actor:%s | Org:%s | Action:%s | Resource:%s:%s | Meta:%s",
		entry.Timestamp.Format(time.RFC3339),
		entry.ActorID,
		entry.OrgID,
		entry.Action,
		entry.ResourceType,
		entry.ResourceID,
		string(meta),
	)
}
