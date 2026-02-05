package policy

import (
	"context"
	"os"
	"testing"
)

func TestJSONPolicyEngine_Check(t *testing.T) {
	// Create temporary policy file
	content := `{
		"roles": {
			"admin": {
				"permissions": ["*"]
			},
			"finance": {
				"permissions": ["payment.create", "refund.create"]
			},
			"developer": {
				"permissions": ["zone.create", "flow.deploy"]
			}
		}
	}`

	tmpfile, err := os.CreateTemp("", "policies.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	engine, err := NewJSONPolicyEngine(tmpfile.Name())
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	tests := []struct {
		name     string
		roles    []Role
		action   Action
		expected bool
	}{
		{
			name:     "Admin can do anything",
			roles:    []Role{RoleAdmin},
			action:   "any.action",
			expected: true,
		},
		{
			name:     "Finance can create payment",
			roles:    []Role{RoleFinance},
			action:   ActionPaymentCreate,
			expected: true,
		},
		{
			name:     "Finance cannot create zone",
			roles:    []Role{RoleFinance},
			action:   ActionZoneCreate,
			expected: false,
		},
		{
			name:     "Developer can deploy flow",
			roles:    []Role{RoleDeveloper},
			action:   ActionFlowDeploy,
			expected: true,
		},
		{
			name:     "Unknown role is denied",
			roles:    []Role{"intruder"},
			action:   ActionPaymentCreate,
			expected: false,
		},
		{
			name:     "Multiple roles (Finance + Developer)",
			roles:    []Role{RoleFinance, RoleDeveloper},
			action:   ActionZoneCreate,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pctx := &PolicyContext{
				Roles:  tt.roles,
				Action: tt.action,
			}
			result, err := engine.Check(context.Background(), pctx)
			if err != nil {
				t.Errorf("Check() error = %v", err)
				return
			}
			if result.Allowed != tt.expected {
				t.Errorf("Check() allowed = %v, expected %v (Reason: %s)", result.Allowed, tt.expected, result.Reason)
			}
		})
	}
}
