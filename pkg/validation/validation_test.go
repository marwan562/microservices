package validation

import (
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name       string
		validators []Validator
		wantErr    bool
		errMatch   string
	}{
		{
			name:       "All pass",
			validators: []Validator{Email("test@example.com"), NotEmpty("hello", "field")},
			wantErr:    false,
		},
		{
			name:       "Email fail",
			validators: []Validator{Email("invalid"), NotEmpty("hello", "field")},
			wantErr:    true,
			errMatch:   "invalid email format",
		},
		{
			name:       "NotEmpty fail",
			validators: []Validator{NotEmpty("  ", "field")},
			wantErr:    true,
			errMatch:   "field cannot be empty",
		},
		{
			name:       "UUID fail",
			validators: []Validator{UUID("123", "id")},
			wantErr:    true,
			errMatch:   "invalid id format",
		},
		{
			name:       "MinLength fail",
			validators: []Validator{MinLength("abc", 5, "field")},
			wantErr:    true,
			errMatch:   "field must be at least 5 characters",
		},
		{
			name:       "InList fail",
			validators: []Validator{InList("c", []string{"a", "b"}, "field")},
			wantErr:    true,
			errMatch:   "invalid field: must be one of [a, b]",
		},
		{
			name:       "PositiveAmount fail",
			validators: []Validator{PositiveAmount(0, "amount")},
			wantErr:    true,
			errMatch:   "amount must be greater than zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.validators...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && !strings.Contains(err.Error(), tt.errMatch) {
				t.Errorf("Validate() error = %v, wantMatch %v", err, tt.errMatch)
			}
		})
	}
}
