package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	EnvDevelopment = "development"
	EnvStaging     = "staging"
	EnvProduction  = "production"
)

var productionRequired = map[string][]string{
	"activity-service": {
		"MYSQL_DSN",
		"REDIS_ADDR",
	},
	"gateway-api": {
		"MYSQL_DSN",
		"JWT_SECRET",
		"ORDER_GRPC_TARGET",
		"USER_GRPC_TARGET",
		"INVENTORY_GRPC_TARGET",
		"TASK_GRPC_TARGET",
		"REFUND_GRPC_TARGET",
		"ACTIVITY_GRPC_TARGET",
		"PRICE_GRPC_TARGET",
	},
	"inventory-service": {
		"MYSQL_DSN",
		"REDIS_ADDR",
	},
	"order-service": {
		"MYSQL_DSN",
		"REDIS_ADDR",
		"MQ_URL",
		"INVENTORY_GRPC_TARGET",
	},
	"payment-service": {
		"MYSQL_DSN",
		"ORDER_GRPC_TARGET",
		"INVENTORY_GRPC_TARGET",
	},
	"price-service": {
		"MYSQL_DSN",
	},
	"refund-service": {
		"MYSQL_DSN",
		"MQ_URL",
		"ORDER_GRPC_TARGET",
		"INVENTORY_GRPC_TARGET",
	},
	"task-service": {
		"MYSQL_DSN",
		"MQ_URL",
		"ORDER_GRPC_TARGET",
		"INVENTORY_GRPC_TARGET",
	},
	"user-service": {
		"MYSQL_DSN",
		"REDIS_ADDR",
	},
}

func GetEnv(key, def string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	return v
}

func GetInt(key string, def int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return def
	}
	return v
}

func AppEnv() string {
	return strings.ToLower(GetEnv("APP_ENV", EnvDevelopment))
}

func IsProduction() bool {
	return AppEnv() == EnvProduction
}

func ValidateService(service string) error {
	if !IsProduction() {
		return nil
	}
	required, ok := productionRequired[service]
	if !ok {
		return fmt.Errorf("unknown service %q for production config validation", service)
	}
	var missing []string
	for _, key := range required {
		if strings.TrimSpace(os.Getenv(key)) == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required production config for %s: %s", service, strings.Join(missing, ", "))
	}
	if err := validateSensitiveDefaults(); err != nil {
		return err
	}
	return nil
}

func validateSensitiveDefaults() error {
	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if secret != "" && (secret == "dev-secret" || secret == "secret" || len(secret) < 32 || looksLikePlaceholder(secret)) {
		return errors.New("JWT_SECRET must be at least 32 characters and must not use a development default in production")
	}
	mysqlDSN := strings.TrimSpace(os.Getenv("MYSQL_DSN"))
	if mysqlDSN != "" && (strings.Contains(mysqlDSN, "root:root") || looksLikePlaceholder(mysqlDSN)) {
		return errors.New("MYSQL_DSN must not use root:root or placeholder credentials in production")
	}
	mqURL := strings.TrimSpace(os.Getenv("MQ_URL"))
	if mqURL != "" && (strings.Contains(mqURL, "guest:guest") || looksLikePlaceholder(mqURL)) {
		return errors.New("MQ_URL must not use guest:guest or placeholder credentials in production")
	}
	return nil
}

func looksLikePlaceholder(v string) bool {
	lower := strings.ToLower(v)
	return strings.Contains(lower, "change-me") || strings.Contains(lower, "replace-me") || strings.Contains(lower, "placeholder")
}
