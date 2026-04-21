package logx

import (
	"regexp"
	"strings"
)

var (
	// Sensitive field patterns
	sensitivePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(password|passwd|pwd)["\s]*[:=]["\s]*[^,"\s]+`),
		regexp.MustCompile(`(?i)(token|api_key|apikey|secret|jwt)["\s]*[:=]["\s]*[^,"\s]+`),
		regexp.MustCompile(`(?i)(authorization|bearer|auth)["\s]*[:=]["\s]*[^,"\s]+`),
		regexp.MustCompile(`(?i)(credit_card|card_number|card_no|cvv|ccv)["\s]*[:=]["\s]*[^,"\s]+`),
		regexp.MustCompile(`(?i)(phone|mobile|phone_number|tel)["\s]*[:=]["\s]*[^,"\s]+`),
		regexp.MustCompile(`(?i)(ssn|social_security|identity)["\s]*[:=]["\s]*[^,"\s]+`),
		regexp.MustCompile(`(?i)(mysql_dsn|mongodb|redis|connection_string)["\s]*[:=]["\s]*[^,"\s]+`),
		regexp.MustCompile(`(?i)(signature|private_key)["\s]*[:=]["\s]*[^,"\s]+`),
	}

	// Replacement strings for different types of sensitive data
	replacements = map[string]string{
		"password":          "***REDACTED***",
		"passwd":            "***REDACTED***",
		"pwd":               "***REDACTED***",
		"token":             "***TOKEN***",
		"api_key":           "***API_KEY***",
		"apikey":            "***API_KEY***",
		"secret":            "***SECRET***",
		"jwt":               "***JWT***",
		"authorization":     "***AUTH***",
		"bearer":            "***TOKEN***",
		"credit_card":        "***CARD***",
		"card_number":        "***CARD***",
		"card_no":           "***CARD***",
		"cvv":               "***CVV***",
		"ccv":               "***CVV***",
		"phone":              "***PHONE***",
		"mobile":            "***PHONE***",
		"phone_number":       "***PHONE***",
		"ssn":               "***SSN***",
		"social_security":    "***SSN***",
		"identity":          "***ID***",
		"mysql_dsn":         "***DSN***",
		"mongodb":           "***CONN***",
		"redis":             "***CONN***",
		"connection_string": "***CONN***",
		"signature":         "***SIGNATURE***",
		"private_key":       "***KEY***",
	}
)

func Desensitize(input string) string {
	if input == "" {
		return input
	}

	result := input

	for _, pattern := range sensitivePatterns {
		result = pattern.ReplaceAllStringFunc(result, func(match string) string {
			return maskSensitiveField(match)
		})
	}

	return result
}

func maskSensitiveField(match string) string {
	lower := strings.ToLower(match)

	for key, replacement := range replacements {
		if strings.Contains(lower, key) {
			parts := strings.Split(match, ":=")
			if len(parts) < 2 {
				parts = strings.Split(match, ":")
			}
			if len(parts) < 2 {
				parts = strings.Split(match, "=")
			}
			if len(parts) == 2 {
				fieldName := strings.Trim(parts[0], " \"\t\n\r")
				fieldValue := strings.Trim(parts[1], " \"\t\n\r")
				if len(fieldValue) > 4 {
					maskedValue := fieldValue[:2] + "***" + fieldValue[len(fieldValue)-2:]
					return fieldName + ": " + maskedValue
				}
				return fieldName + ": ***"
			}
			return replacement
		}
	}

	return "***REDACTED***"
}

func DesensitizeMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return nil
	}

	result := make(map[string]interface{}, len(m))
	for k, v := range m {
		if isSensitiveKey(k) {
			result[k] = "***REDACTED***"
		} else if sv, ok := v.(string); ok {
			result[k] = Desensitize(sv)
		} else {
			result[k] = v
		}
	}
	return result
}

func isSensitiveKey(key string) bool {
	lower := strings.ToLower(key)
	sensitiveKeys := []string{
		"password", "passwd", "pwd",
		"token", "api_key", "apikey", "secret", "jwt",
		"auth", "authorization", "bearer",
		"credit_card", "card_number", "card_no", "cvv", "ccv",
		"phone", "mobile", "phone_number",
		"ssn", "social_security",
		"mysql_dsn", "mongodb", "redis", "connection_string",
		"signature", "private_key",
	}

	for _, k := range sensitiveKeys {
		if strings.Contains(lower, k) {
			return true
		}
	}
	return false
}

func DesensitizeStruct(s interface{}) interface{} {
	if s == nil {
		return nil
	}
	return s
}
