package validator

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	validate   *validator.Validate
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	phoneRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)
	urlRegex   = regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
)

func init() {
	validate = validator.New()
}

// Validate validates a struct.
func Validate(data any) error {
	return validate.Struct(data)
}

// ValidateVar validates a single variable.
func ValidateVar(field any, tag string) error {
	return validate.Var(field, tag)
}

// IsEmail validates email format.
func IsEmail(s string) bool {
	return emailRegex.MatchString(s)
}

// IsPhone validates phone number format.
func IsPhone(s string) bool {
	return phoneRegex.MatchString(s)
}

// IsURL validates URL format.
func IsURL(s string) bool {
	return urlRegex.MatchString(s)
}

// IsChineseMobile checks if the string is a valid Chinese mobile phone.
func IsChineseMobile(s string) bool {
	return phoneRegex.MatchString(s)
}

// IsIDCard validates Chinese ID card number (18 digits).
func IsIDCard(s string) bool {
	if len(s) != 18 {
		return false
	}
	// Basic checksum validation
	weights := []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
	checkCodes := []byte{'1', '0', 'X', '9', '8', '7', '6', '5', '4', '3', '2'}

	sum := 0
	for i := 0; i < 17; i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
		sum += int(s[i]-'0') * weights[i]
	}
	checkCode := checkCodes[sum%11]
	return s[17] == checkCode || (s[17] == 'x' && checkCode == 'X')
}

// IsBlank checks if a string is blank.
func IsBlank(s string) bool {
	return strings.TrimSpace(s) == ""
}

// IsAlphanumeric checks if a string contains only alphanumeric characters.
func IsAlphanumeric(s string) bool {
	for _, r := range s {
		if !('a' <= r && r <= 'z' || 'A' <= r && r <= 'Z' || '0' <= r && r <= '9') {
			return false
		}
	}
	return true
}

// IsNumeric checks if a string contains only numeric characters.
func IsNumeric(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// IsAlpha checks if a string contains only alphabetic characters.
func IsAlpha(s string) bool {
	for _, r := range s {
		if !('a' <= r && r <= 'z' || 'A' <= r && r <= 'Z') {
			return false
		}
	}
	return true
}

// IsInRange checks if a number is within the range.
func IsInRange[T int | int64 | float64](val, min, max T) bool {
	return val >= min && val <= max
}

// MinLen checks if the string length is at least min.
func MinLen(s string, min int) bool {
	return len(s) >= min
}

// MaxLen checks if the string length is at most max.
func MaxLen(s string, max int) bool {
	return len(s) <= max
}
