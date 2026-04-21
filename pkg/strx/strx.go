package strx

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"regexp"
	"strings"
	"unicode"
)

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	phoneRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)
	uuidRegex  = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
)

// IsEmail validates if the string is a valid email address.
func IsEmail(s string) bool {
	return emailRegex.MatchString(s)
}

// IsPhone validates if the string is a valid Chinese phone number.
func IsPhone(s string) bool {
	return phoneRegex.MatchString(s)
}

// IsUUID validates if the string is a valid UUID.
func IsUUID(s string) bool {
	return uuidRegex.MatchString(s)
}

// MaskPhone masks a phone number like 138****1234.
func MaskPhone(phone string) string {
	if len(phone) != 11 {
		return phone
	}
	return phone[:3] + "****" + phone[7:]
}

// MaskEmail masks an email like t***@example.com.
func MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}
	username := parts[0]
	if len(username) <= 1 {
		return email
	}
	return string(username[0]) + strings.Repeat("*", min(3, len(username)-1)) + "@" + parts[1]
}

// MaskIDCard masks an ID card number like 110***********1234.
func MaskIDCard(id string) string {
	if len(id) < 8 {
		return strings.Repeat("*", len(id))
	}
	return id[:4] + strings.Repeat("*", len(id)-8) + id[len(id)-4:]
}

// Truncate truncates a string to the specified length with ellipsis.
func Truncate(s string, maxLen int) string {
	if maxLen <= 0 || len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}

// SnakeToCamel converts snake_case to camelCase.
func SnakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	for i := range parts {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

// CamelToSnake converts camelCase to snake_case.
func CamelToSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteByte('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// MD5 returns the MD5 hash of the string.
func MD5(s string) string {
	h := md5.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}

// SHA256 returns the SHA256 hash of the string.
func SHA256(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// SHA1 returns the SHA1 hash of the string.
func SHA1(s string) string {
	h := sha1.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}

// Base64Encode encodes a string to base64.
func Base64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// Base64Decode decodes a base64 string.
func Base64Decode(s string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FirstUpper returns the string with first character uppercased.
func FirstUpper(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// FirstLower returns the string with first character lowercased.
func FirstLower(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

// Contains checks if the slice contains the substring.
func Contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// JoinNonEmpty joins non-empty strings with the separator.
func JoinNonEmpty(sep string, parts ...string) string {
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			result = append(result, p)
		}
	}
	return strings.Join(result, sep)
}
