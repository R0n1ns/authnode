package validator

import (
	"regexp"
	"strings"
)

// EmailRegex is a regular expression for validating email addresses
var EmailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// NicknameRegex is a regular expression for validating nicknames (only alphanumeric characters)
var NicknameRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

// ValidateEmail validates an email address
func ValidateEmail(email string) bool {
	return EmailRegex.MatchString(email)
}

// ValidateNickname validates a nickname
func ValidateNickname(nickname string) bool {
	return NicknameRegex.MatchString(nickname)
}

// ValidateRequiredField validates if a field is not empty
func ValidateRequiredField(field string) bool {
	return strings.TrimSpace(field) != ""
}

// ValidationError represents a validation error for a specific field
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors represents a collection of validation errors
type ValidationErrors struct {
	Errors []ValidationError `json:"detailedErrors"`
}

// NewValidationErrors creates a new ValidationErrors struct
func NewValidationErrors() *ValidationErrors {
	return &ValidationErrors{
		Errors: []ValidationError{},
	}
}

// AddError adds a validation error
func (v *ValidationErrors) AddError(field, message string) {
	v.Errors = append(v.Errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

// HasErrors checks if there are any validation errors
func (v *ValidationErrors) HasErrors() bool {
	return len(v.Errors) > 0
}
