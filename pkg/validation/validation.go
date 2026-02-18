package validation

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

var (
	// Basic regex for demonstration, can be refined.
	phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
)

// Validator is a function that returns an error if validation fails.
type Validator func() error

// Validate runs all validators and returns the first error encountered.
func Validate(validators ...Validator) error {
	for _, v := range validators {
		if err := v(); err != nil {
			return err
		}
	}
	return nil
}

// Email validates email format.
func Email(email string) Validator {
	return func() error {
		_, err := mail.ParseAddress(email)
		if err != nil {
			return fmt.Errorf("invalid email format")
		}
		return nil
	}
}

// UUID validates that a string is a valid UUID.
func UUID(id string, fieldName string) Validator {
	return func() error {
		if _, err := uuid.Parse(id); err != nil {
			return fmt.Errorf("invalid %s format: must be a valid UUID", fieldName)
		}
		return nil
	}
}

// MinLength validates minimum string length after trimming whitespace.
func MinLength(str string, min int, fieldName string) Validator {
	return func() error {
		if len(strings.TrimSpace(str)) < min {
			return fmt.Errorf("%s must be at least %d characters", fieldName, min)
		}
		return nil
	}
}

// MaxLength validates maximum string length.
func MaxLength(str string, max int, fieldName string) Validator {
	return func() error {
		if len(str) > max {
			return fmt.Errorf("%s cannot exceed %d characters", fieldName, max)
		}
		return nil
	}
}

// InList validates that a value is in a list of allowed strings.
func InList(val string, list []string, fieldName string) Validator {
	return func() error {
		for _, item := range list {
			if val == item {
				return nil
			}
		}
		return fmt.Errorf("invalid %s: must be one of [%s]", fieldName, strings.Join(list, ", "))
	}
}

// PositiveAmount validates that a number is greater than zero.
func PositiveAmount(amount int64, fieldName string) Validator {
	return func() error {
		if amount <= 0 {
			return fmt.Errorf("%s must be greater than zero", fieldName)
		}
		return nil
	}
}

// NotEmpty validates that a string is not empty or whitespace.
func NotEmpty(val string, fieldName string) Validator {
	return func() error {
		if strings.TrimSpace(val) == "" {
			return fmt.Errorf("%s cannot be empty", fieldName)
		}
		return nil
	}
}
