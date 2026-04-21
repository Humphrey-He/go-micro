package config

import (
	"strings"
	"testing"
)

func TestValidateService_SkipsStrictChecksOutsideProduction(t *testing.T) {
	t.Setenv("APP_ENV", EnvDevelopment)
	t.Setenv("MYSQL_DSN", "")

	if err := ValidateService("order-service"); err != nil {
		t.Fatalf("expected development validation to pass, got %v", err)
	}
}

func TestValidateService_RequiresProductionConfig(t *testing.T) {
	t.Setenv("APP_ENV", EnvProduction)
	t.Setenv("MYSQL_DSN", "app:strong-password@tcp(mysql:3306)/go_micro")

	err := ValidateService("order-service")
	if err == nil {
		t.Fatal("expected missing production config error")
	}
	msg := err.Error()
	for _, want := range []string{"REDIS_ADDR", "MQ_URL", "INVENTORY_GRPC_TARGET"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("expected error to mention %s, got %q", want, msg)
		}
	}
}

func TestValidateService_RejectsWeakJWTSecretInProduction(t *testing.T) {
	t.Setenv("APP_ENV", EnvProduction)
	t.Setenv("MYSQL_DSN", "app:strong-password@tcp(mysql:3306)/go_micro")
	t.Setenv("JWT_SECRET", "dev-secret")
	t.Setenv("ORDER_GRPC_TARGET", "order-service:9081")
	t.Setenv("USER_GRPC_TARGET", "user-service:9083")
	t.Setenv("INVENTORY_GRPC_TARGET", "inventory-service:9082")
	t.Setenv("TASK_GRPC_TARGET", "task-service:9084")
	t.Setenv("REFUND_GRPC_TARGET", "refund-service:9086")
	t.Setenv("ACTIVITY_GRPC_TARGET", "activity-service:9087")
	t.Setenv("PRICE_GRPC_TARGET", "price-service:9088")

	err := ValidateService("gateway-api")
	if err == nil {
		t.Fatal("expected weak jwt secret error")
	}
	if !strings.Contains(err.Error(), "JWT_SECRET") {
		t.Fatalf("expected error to mention JWT_SECRET, got %q", err.Error())
	}
}

func TestValidateService_AcceptsProductionGatewayConfig(t *testing.T) {
	t.Setenv("APP_ENV", EnvProduction)
	t.Setenv("MYSQL_DSN", "app:strong-password@tcp(mysql:3306)/go_micro")
	t.Setenv("JWT_SECRET", "0123456789abcdef0123456789abcdef")
	t.Setenv("ORDER_GRPC_TARGET", "order-service:9081")
	t.Setenv("USER_GRPC_TARGET", "user-service:9083")
	t.Setenv("INVENTORY_GRPC_TARGET", "inventory-service:9082")
	t.Setenv("TASK_GRPC_TARGET", "task-service:9084")
	t.Setenv("REFUND_GRPC_TARGET", "refund-service:9086")
	t.Setenv("ACTIVITY_GRPC_TARGET", "activity-service:9087")
	t.Setenv("PRICE_GRPC_TARGET", "price-service:9088")

	if err := ValidateService("gateway-api"); err != nil {
		t.Fatalf("expected production gateway config to pass, got %v", err)
	}
}

func TestGetInt(t *testing.T) {
	t.Setenv("POSITIVE_INT", "42")
	t.Setenv("BAD_INT", "oops")
	t.Setenv("ZERO_INT", "0")

	if got := GetInt("POSITIVE_INT", 7); got != 42 {
		t.Fatalf("expected 42, got %d", got)
	}
	if got := GetInt("BAD_INT", 7); got != 7 {
		t.Fatalf("expected default for invalid int, got %d", got)
	}
	if got := GetInt("ZERO_INT", 7); got != 7 {
		t.Fatalf("expected default for zero int, got %d", got)
	}
}
