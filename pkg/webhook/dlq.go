package webhook

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

// DLQEntry represents a dead letter queue entry for failed webhooks.
type DLQEntry struct {
	ID              string            `json:"id"`
	WebhookID       string            `json:"webhook_id"`
	ZoneID          string            `json:"zone_id"`
	URL             string            `json:"url"`
	Payload         json.RawMessage   `json:"payload"`
	Headers         map[string]string `json:"headers"`
	FailureReason   string            `json:"failure_reason"`
	LastStatusCode  int               `json:"last_status_code"`
	AttemptCount    int               `json:"attempt_count"`
	MaxAttempts     int               `json:"max_attempts"`
	FirstFailedAt   time.Time         `json:"first_failed_at"`
	LastAttemptedAt time.Time         `json:"last_attempted_at"`
	NextRetryAt     *time.Time        `json:"next_retry_at"`
	ExpiresAt       time.Time         `json:"expires_at"`
	Status          DLQStatus         `json:"status"`
}

// DLQStatus represents the status of a DLQ entry.
type DLQStatus string

const (
	// DLQStatusPending - awaiting retry.
	DLQStatusPending DLQStatus = "pending"

	// DLQStatusRetrying - currently being retried.
	DLQStatusRetrying DLQStatus = "retrying"

	// DLQStatusSucceeded - retry succeeded.
	DLQStatusSucceeded DLQStatus = "succeeded"

	// DLQStatusExhausted - all retries exhausted.
	DLQStatusExhausted DLQStatus = "exhausted"

	// DLQStatusExpired - entry expired before completion.
	DLQStatusExpired DLQStatus = "expired"

	// DLQStatusAbandoned - manually abandoned.
	DLQStatusAbandoned DLQStatus = "abandoned"
)

// DeadLetterQueue manages failed webhook deliveries.
type DeadLetterQueue struct {
	db          *sql.DB
	maxRetries  int
	retryDelays []time.Duration
	ttl         time.Duration
}

// DLQConfig configures the dead letter queue.
type DLQConfig struct {
	// MaxRetries is the maximum number of retry attempts (default: 5).
	MaxRetries int

	// RetryDelays defines the backoff schedule (default: exponential).
	RetryDelays []time.Duration

	// TTL is how long entries stay in the DLQ before expiring (default: 7 days).
	TTL time.Duration
}

// NewDeadLetterQueue creates a new dead letter queue.
func NewDeadLetterQueue(db *sql.DB, cfg DLQConfig) *DeadLetterQueue {
	dlq := &DeadLetterQueue{
		db:          db,
		maxRetries:  cfg.MaxRetries,
		retryDelays: cfg.RetryDelays,
		ttl:         cfg.TTL,
	}

	if dlq.maxRetries == 0 {
		dlq.maxRetries = 5
	}

	if len(dlq.retryDelays) == 0 {
		// Default exponential backoff: 1m, 5m, 30m, 2h, 6h
		dlq.retryDelays = []time.Duration{
			1 * time.Minute,
			5 * time.Minute,
			30 * time.Minute,
			2 * time.Hour,
			6 * time.Hour,
		}
	}

	if dlq.ttl == 0 {
		dlq.ttl = 7 * 24 * time.Hour // 7 days
	}

	return dlq
}

// Add adds a failed webhook to the dead letter queue.
func (q *DeadLetterQueue) Add(ctx context.Context, entry DLQEntry) error {
	entry.Status = DLQStatusPending
	entry.FirstFailedAt = time.Now()
	entry.LastAttemptedAt = entry.FirstFailedAt
	entry.AttemptCount = 1
	entry.MaxAttempts = q.maxRetries
	entry.ExpiresAt = time.Now().Add(q.ttl)

	// Calculate next retry
	nextRetry := time.Now().Add(q.getRetryDelay(entry.AttemptCount))
	entry.NextRetryAt = &nextRetry

	headers, _ := json.Marshal(entry.Headers)

	query := `
		INSERT INTO webhook_dlq (
			id, webhook_id, zone_id, url, payload, headers,
			failure_reason, last_status_code, attempt_count, max_attempts,
			first_failed_at, last_attempted_at, next_retry_at, expires_at, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err := q.db.ExecContext(ctx, query,
		entry.ID, entry.WebhookID, entry.ZoneID, entry.URL, entry.Payload, headers,
		entry.FailureReason, entry.LastStatusCode, entry.AttemptCount, entry.MaxAttempts,
		entry.FirstFailedAt, entry.LastAttemptedAt, entry.NextRetryAt, entry.ExpiresAt, entry.Status,
	)

	return err
}

// GetPendingRetries returns entries ready for retry.
func (q *DeadLetterQueue) GetPendingRetries(ctx context.Context, limit int) ([]DLQEntry, error) {
	query := `
		SELECT id, webhook_id, zone_id, url, payload, headers,
			failure_reason, last_status_code, attempt_count, max_attempts,
			first_failed_at, last_attempted_at, next_retry_at, expires_at, status
		FROM webhook_dlq
		WHERE status = $1 
			AND next_retry_at <= NOW()
			AND expires_at > NOW()
		ORDER BY next_retry_at
		LIMIT $2
	`

	rows, err := q.db.QueryContext(ctx, query, DLQStatusPending, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return q.scanEntries(rows)
}

// MarkRetrying marks an entry as currently being retried.
func (q *DeadLetterQueue) MarkRetrying(ctx context.Context, id string) error {
	query := `UPDATE webhook_dlq SET status = $1 WHERE id = $2`
	_, err := q.db.ExecContext(ctx, query, DLQStatusRetrying, id)
	return err
}

// RecordSuccess marks a retry as successful.
func (q *DeadLetterQueue) RecordSuccess(ctx context.Context, id string) error {
	query := `
		UPDATE webhook_dlq 
		SET status = $1, last_attempted_at = NOW(), next_retry_at = NULL
		WHERE id = $2
	`
	_, err := q.db.ExecContext(ctx, query, DLQStatusSucceeded, id)
	return err
}

// RecordFailure records another failed attempt.
func (q *DeadLetterQueue) RecordFailure(ctx context.Context, id string, reason string, statusCode int) error {
	// Get current attempt count
	var attemptCount, maxAttempts int
	err := q.db.QueryRowContext(ctx,
		`SELECT attempt_count, max_attempts FROM webhook_dlq WHERE id = $1`, id,
	).Scan(&attemptCount, &maxAttempts)
	if err != nil {
		return err
	}

	attemptCount++
	var status DLQStatus
	var nextRetry *time.Time

	if attemptCount >= maxAttempts {
		status = DLQStatusExhausted
	} else {
		status = DLQStatusPending
		retry := time.Now().Add(q.getRetryDelay(attemptCount))
		nextRetry = &retry
	}

	query := `
		UPDATE webhook_dlq 
		SET status = $1, attempt_count = $2, last_attempted_at = NOW(),
			next_retry_at = $3, failure_reason = $4, last_status_code = $5
		WHERE id = $6
	`
	_, err = q.db.ExecContext(ctx, query, status, attemptCount, nextRetry, reason, statusCode, id)
	return err
}

// Abandon manually abandons a DLQ entry (stop retrying).
func (q *DeadLetterQueue) Abandon(ctx context.Context, id string) error {
	query := `UPDATE webhook_dlq SET status = $1 WHERE id = $2`
	_, err := q.db.ExecContext(ctx, query, DLQStatusAbandoned, id)
	return err
}

// Replay manually triggers a retry for a DLQ entry.
func (q *DeadLetterQueue) Replay(ctx context.Context, id string) error {
	now := time.Now()
	query := `UPDATE webhook_dlq SET status = $1, next_retry_at = $2 WHERE id = $3`
	_, err := q.db.ExecContext(ctx, query, DLQStatusPending, now, id)
	return err
}

// GetStats returns DLQ statistics for a zone.
func (q *DeadLetterQueue) GetStats(ctx context.Context, zoneID string) (*DLQStats, error) {
	query := `
		SELECT 
			COUNT(*) FILTER (WHERE status = 'pending') as pending,
			COUNT(*) FILTER (WHERE status = 'retrying') as retrying,
			COUNT(*) FILTER (WHERE status = 'succeeded') as succeeded,
			COUNT(*) FILTER (WHERE status = 'exhausted') as exhausted,
			COUNT(*) FILTER (WHERE status = 'expired') as expired,
			COUNT(*) FILTER (WHERE status = 'abandoned') as abandoned
		FROM webhook_dlq
		WHERE zone_id = $1
	`

	var stats DLQStats
	err := q.db.QueryRowContext(ctx, query, zoneID).Scan(
		&stats.Pending, &stats.Retrying, &stats.Succeeded,
		&stats.Exhausted, &stats.Expired, &stats.Abandoned,
	)
	return &stats, err
}

// DLQStats contains DLQ statistics.
type DLQStats struct {
	Pending   int `json:"pending"`
	Retrying  int `json:"retrying"`
	Succeeded int `json:"succeeded"`
	Exhausted int `json:"exhausted"`
	Expired   int `json:"expired"`
	Abandoned int `json:"abandoned"`
}

// CleanupExpired removes expired entries.
func (q *DeadLetterQueue) CleanupExpired(ctx context.Context) (int64, error) {
	result, err := q.db.ExecContext(ctx,
		`UPDATE webhook_dlq SET status = $1 WHERE status = $2 AND expires_at <= NOW()`,
		DLQStatusExpired, DLQStatusPending,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (q *DeadLetterQueue) getRetryDelay(attemptCount int) time.Duration {
	if attemptCount <= 0 {
		attemptCount = 1
	}
	idx := attemptCount - 1
	if idx >= len(q.retryDelays) {
		idx = len(q.retryDelays) - 1
	}
	return q.retryDelays[idx]
}

func (q *DeadLetterQueue) scanEntries(rows *sql.Rows) ([]DLQEntry, error) {
	var entries []DLQEntry
	for rows.Next() {
		var e DLQEntry
		var headers string
		err := rows.Scan(
			&e.ID, &e.WebhookID, &e.ZoneID, &e.URL, &e.Payload, &headers,
			&e.FailureReason, &e.LastStatusCode, &e.AttemptCount, &e.MaxAttempts,
			&e.FirstFailedAt, &e.LastAttemptedAt, &e.NextRetryAt, &e.ExpiresAt, &e.Status,
		)
		if err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(headers), &e.Headers)
		entries = append(entries, e)
	}
	return entries, nil
}
