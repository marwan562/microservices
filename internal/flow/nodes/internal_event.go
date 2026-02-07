package nodes

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// InternalEventNode emits an event to Redis Streams to trigger other flows
type InternalEventNode struct {
	NodeID    string            `json:"id"`
	ZoneID    string            `json:"zone_id"`
	EventType string            `json:"event_type"`
	Payload   map[string]string `json:"payload,omitempty"` // Template mappings
	NextNode  string            `json:"next,omitempty"`
	rdb       *redis.Client     `json:"-"`
}

// InternalEventConfig is used to create a new internal event node
type InternalEventConfig struct {
	ID        string
	ZoneID    string
	EventType string
	Payload   map[string]string
	NextNode  string
	Redis     *redis.Client
}

// NewInternalEventNode creates a new internal event action node
func NewInternalEventNode(config InternalEventConfig) *InternalEventNode {
	return &InternalEventNode{
		NodeID:    config.ID,
		ZoneID:    config.ZoneID,
		EventType: config.EventType,
		Payload:   config.Payload,
		NextNode:  config.NextNode,
		rdb:       config.Redis,
	}
}

// ID returns the node ID
func (n *InternalEventNode) ID() string {
	return n.NodeID
}

// Type returns the node type
func (n *InternalEventNode) Type() string {
	return "internal_event"
}

// Execute emits an event to Redis Streams
func (n *InternalEventNode) Execute(ctx context.Context, input map[string]interface{}) (*NodeResult, error) {
	if n.rdb == nil {
		return &NodeResult{
			Success: false,
			Error:   "Redis client not configured",
		}, fmt.Errorf("redis client not configured")
	}

	// Resolve zone ID from input if not set
	zoneID := n.ZoneID
	if zoneID == "" {
		if z, ok := input["zone_id"].(string); ok {
			zoneID = z
		}
	}

	if zoneID == "" {
		return &NodeResult{
			Success: false,
			Error:   "zone_id not specified",
		}, fmt.Errorf("zone_id not specified")
	}

	// Build payload from input using mappings
	payload := make(map[string]interface{})
	for key, path := range n.Payload {
		val, err := extractValue(input, path)
		if err == nil {
			payload[key] = val
		}
	}

	// If no mappings, use entire input
	if len(n.Payload) == 0 {
		payload = input
	}

	// Construct stream name
	topic := fmt.Sprintf("zone.%s.event.%s", zoneID, n.EventType)
	payloadBytes, _ := json.Marshal(payload)

	// Publish to Redis Stream
	err := n.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: topic,
		Values: map[string]interface{}{
			"data": payloadBytes,
			"ts":   time.Now().Unix(),
		},
	}).Err()

	if err != nil {
		return &NodeResult{
			Success: false,
			Error:   fmt.Sprintf("failed to emit event: %v", err),
		}, err
	}

	return &NodeResult{
		Success: true,
		Output: map[string]interface{}{
			"topic":      topic,
			"event_type": n.EventType,
			"payload":    payload,
		},
		Next: n.NextNode,
	}, nil
}

// InternalEventBuilder provides a fluent interface for building internal event nodes
type InternalEventBuilder struct {
	config InternalEventConfig
}

// NewInternalEvent starts building an internal event node
func NewInternalEvent(id string) *InternalEventBuilder {
	return &InternalEventBuilder{
		config: InternalEventConfig{
			ID:      id,
			Payload: make(map[string]string),
		},
	}
}

// Zone sets the target zone ID
func (b *InternalEventBuilder) Zone(zoneID string) *InternalEventBuilder {
	b.config.ZoneID = zoneID
	return b
}

// EventType sets the event type to emit
func (b *InternalEventBuilder) EventType(eventType string) *InternalEventBuilder {
	b.config.EventType = eventType
	return b
}

// MapField maps an output field from input path
func (b *InternalEventBuilder) MapField(outputKey, inputPath string) *InternalEventBuilder {
	b.config.Payload[outputKey] = inputPath
	return b
}

// WithRedis sets the Redis client
func (b *InternalEventBuilder) WithRedis(rdb *redis.Client) *InternalEventBuilder {
	b.config.Redis = rdb
	return b
}

// Then sets the next node on success
func (b *InternalEventBuilder) Then(nodeID string) *InternalEventBuilder {
	b.config.NextNode = nodeID
	return b
}

// Build creates the internal event node
func (b *InternalEventBuilder) Build() *InternalEventNode {
	return NewInternalEventNode(b.config)
}
