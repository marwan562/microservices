package admin

import (
	"context"
	"sync"
	"time"
)

// MaintenanceManager handles system maintenance state.
type MaintenanceManager struct {
	mu         sync.RWMutex
	enabled    bool
	message    string
	allowedIPs map[string]struct{}
	planedEnd  time.Time
}

// NewMaintenanceManager creates a new maintenance manager.
func NewMaintenanceManager() *MaintenanceManager {
	return &MaintenanceManager{
		allowedIPs: make(map[string]struct{}),
	}
}

// Enable enables maintenance mode.
func (m *MaintenanceManager) Enable(message string, plannedEnd time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = true
	m.message = message
	m.planedEnd = plannedEnd
}

// Disable disables maintenance mode.
func (m *MaintenanceManager) Disable() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = false
	m.message = ""
	m.planedEnd = time.Time{}
}

// IsEnabled checks if maintenance mode is enabled.
func (m *MaintenanceManager) IsEnabled() (bool, string, time.Time) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.enabled, m.message, m.planedEnd
}

// AllowIP adds an IP to the allowlist (bypasses maintenance mode).
func (m *MaintenanceManager) AllowIP(ip string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.allowedIPs[ip] = struct{}{}
}

// RemoveIP removes an IP from the allowlist.
func (m *MaintenanceManager) RemoveIP(ip string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.allowedIPs, ip)
}

// IsAllowed checks if an IP is allowed during maintenance.
func (m *MaintenanceManager) IsAllowed(ip string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, allowed := m.allowedIPs[ip]
	return allowed
}

// Status returns the current maintenance status.
type Status struct {
	Enabled    bool      `json:"enabled"`
	Message    string    `json:"message,omitempty"`
	PlannedEnd time.Time `json:"planned_end,omitempty"`
}

// GetStatus returns the status struct.
func (m *MaintenanceManager) GetStatus() Status {
	enabled, msg, end := m.IsEnabled()
	return Status{
		Enabled:    enabled,
		Message:    msg,
		PlannedEnd: end,
	}
}

// HealthCheck verifies if the system is operational from an admin perspective.
func (m *MaintenanceManager) HealthCheck(ctx context.Context) error {
	// In a real system, this would check critical dependencies
	// For now, it's a simple pass-through
	return nil
}
