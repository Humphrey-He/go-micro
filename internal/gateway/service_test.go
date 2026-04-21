package gateway

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestComputeViewStatus(t *testing.T) {
	tests := []struct {
		name        string
		orderStatus string
		taskStatus  string
		taskType    string
		resvStatus  string
		wantView    string
		wantReason  string
	}{
		{name: "canceled", orderStatus: "CANCELED", taskStatus: "FAILED", taskType: "FULFILL", wantView: "CANCELED"},
		{name: "timeout", orderStatus: "CANCELED", taskStatus: "DEAD", taskType: "TIMEOUT_CANCEL", wantView: "TIMEOUT", wantReason: "timeout"},
		{name: "dead", orderStatus: "RESERVED", taskStatus: "DEAD", taskType: "FULFILL", wantView: "DEAD"},
		{name: "failed", orderStatus: "RESERVED", taskStatus: "FAILED", taskType: "FULFILL", wantView: "FAILED"},
		{name: "success", orderStatus: "SUCCESS", taskStatus: "SUCCESS", taskType: "FULFILL", wantView: "SUCCESS"},
		{name: "processing", orderStatus: "RESERVED", taskStatus: "RUNNING", taskType: "FULFILL", wantView: "PROCESSING"},
		{name: "pending", orderStatus: "RESERVED", taskStatus: "PENDING", taskType: "FULFILL", wantView: "PENDING"},
		{name: "pending_not_found", orderStatus: "RESERVED", taskStatus: "NOT_FOUND", taskType: "FULFILL", wantView: "PENDING"},
		{name: "unknown", orderStatus: "", taskStatus: "", taskType: "", resvStatus: "", wantView: "UNKNOWN"},
		{name: "order_status_fallback", orderStatus: "CREATED", taskStatus: "", taskType: "", wantView: "CREATED"},
		{name: "canceled_no_task", orderStatus: "CANCELED", taskStatus: "", taskType: "", wantView: "CANCELED"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotView, gotReason := computeViewStatus(tt.orderStatus, tt.taskStatus, tt.taskType, tt.resvStatus)
			if gotView != tt.wantView {
				t.Fatalf("view_status expected %s, got %s", tt.wantView, gotView)
			}
			if gotReason != tt.wantReason {
				t.Fatalf("cancel_reason expected %s, got %s", tt.wantReason, gotReason)
			}
		})
	}
}

func TestGetInt(t *testing.T) {
	cases := []struct {
		key string
		def int
		want int
	}{
		{"", 0, 0},
		{"", 10, 10},
	}

	for _, c := range cases {
		t.Run(c.key, func(t *testing.T) {
			got := getInt(c.key, c.def)
			if got != c.want {
				t.Fatalf("getInt(%s, %d): got %d, want %d", c.key, c.def, got, c.want)
			}
		})
	}
}

func TestNewBreakerFromEnv(t *testing.T) {
	breaker := newBreakerFromEnv()
	if breaker == nil {
		t.Fatal("expected non-nil circuit breaker")
	}
}

func TestJWTTokenGeneration(t *testing.T) {
	secret := []byte("test-secret")
	claims := jwt.MapClaims{
		"user_id":  "U-1",
		"username": "testuser",
		"role":     "user",
		"exp":      9999999999,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(secret)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	if signed == "" {
		t.Fatal("expected non-empty token")
	}

	parsed, err := jwt.Parse(signed, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}
	if !parsed.Valid {
		t.Fatal("expected valid token")
	}
}
