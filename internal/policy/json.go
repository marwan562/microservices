package policy

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

// JSONPolicyEngine implements PolicyEngine using dynamic JSON configuration
type JSONPolicyEngine struct {
	Roles map[Role]RolePolicy `json:"roles"`
}

type RolePolicy struct {
	Permissions []Action `json:"permissions"`
}

// NewJSONPolicyEngine creates a new JSON policy engine from a file
func NewJSONPolicyEngine(path string) (*JSONPolicyEngine, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read policy file: %w", err)
	}

	var engine JSONPolicyEngine
	if err := json.Unmarshal(data, &engine); err != nil {
		return nil, fmt.Errorf("failed to unmarshal policies: %w", err)
	}

	return &engine, nil
}

// Check evaluates policies loaded from JSON
func (e *JSONPolicyEngine) Check(ctx context.Context, pctx *PolicyContext) (*PolicyResult, error) {
	result := &PolicyResult{
		Allowed: false,
		Rules:   make([]string, 0),
	}

	for _, roleName := range pctx.Roles {
		policy, ok := e.Roles[roleName]
		if !ok {
			continue
		}

		for _, perm := range policy.Permissions {
			// Wildcard check
			if perm == "*" {
				result.Allowed = true
				result.Reason = fmt.Sprintf("allowed by wildcard rule for role: %s", roleName)
				result.Rules = append(result.Rules, fmt.Sprintf("role:%s:*", roleName))
				return result, nil
			}

			// Direct action match
			if perm == pctx.Action {
				result.Allowed = true
				result.Reason = fmt.Sprintf("allowed by explicit rule for role: %s", roleName)
				result.Rules = append(result.Rules, fmt.Sprintf("role:%s:%s", roleName, perm))
				return result, nil
			}
		}
	}

	result.Reason = "no matching policy found in JSON configuration"
	return result, nil
}
