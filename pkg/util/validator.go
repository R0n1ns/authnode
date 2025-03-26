package util

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator defines a custom validator
type Validator struct {
	validator *validator.Validate
}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	v := validator.New()

	// Register custom validations
	err := v.RegisterValidation("nickname", validateNickname)
	if err != nil {
		panic(err)
	}

	return &Validator{
		validator: v,
	}
}

// Validate validates a struct
func (v *Validator) Validate(i interface{}) error {
	return v.validator.Struct(i)
}

// validateNickname validates a nickname
// Must contain only alphanumeric characters (a-zA-Z0-9)
func validateNickname(fl validator.FieldLevel) bool {
	nickname := fl.Field().String()

	// Check if nickname is empty
	if nickname == "" {
		return false
	}

	// Check if nickname contains only alphanumeric characters
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9]+$`, nickname)
	return matched
}

// ValidateEmail checks if an email is valid
func ValidateEmail(email string) bool {
	// Basic email validation
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

// ValidateNickname checks if a nickname is valid
func ValidateNickname(nickname string) bool {
	// Check if nickname is empty
	if nickname == "" {
		return false
	}

	// Check if nickname contains only alphanumeric characters
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9]+$`, nickname)
	return matched
}

// CheckPasswordStrength checks if a password meets the strength requirements
func CheckPasswordStrength(password string) (bool, string) {
	if len(password) < 8 {
		return false, "Password must be at least 8 characters long"
	}

	// Check for uppercase letters
	if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		return false, "Password must contain at least one uppercase letter"
	}

	// Check for lowercase letters
	if !regexp.MustCompile(`[a-z]`).MatchString(password) {
		return false, "Password must contain at least one lowercase letter"
	}

	// Check for numbers
	if !regexp.MustCompile(`[0-9]`).MatchString(password) {
		return false, "Password must contain at least one number"
	}

	// Check for special characters
	if !regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password) {
		return false, "Password must contain at least one special character"
	}

	return true, ""
}

// SanitizeString removes unwanted characters from a string
func SanitizeString(s string) string {
	// Remove leading and trailing spaces
	s = strings.TrimSpace(s)

	// Replace multiple spaces with a single space
	re := regexp.MustCompile(`\s+`)
	s = re.ReplaceAllString(s, " ")

	return s
}
