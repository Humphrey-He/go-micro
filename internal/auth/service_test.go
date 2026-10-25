package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

	"ecomm-micro/pkg/middleware"
)

const testPhone = "17600000000"

func newAuthService(t *testing.T) (*Service, sqlmock.Sqlmock, *miniredis.Miniredis) {
	t.Helper()
	t.Setenv("SMS_DEBUG_CODE", "")
	t.Setenv("SMS_PROVIDER", "")

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	sqlxDB := sqlx.NewDb(db, "mysql")

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() {
		sqlxDB.Close()
		rdb.Close()
	})

	return NewService(sqlxDB, rdb), mock, mr
}

func userRows(phone, status string) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "user_id", "username", "password_hash", "phone", "role", "status", "created_at", "updated_at"}).
		AddRow(1, "USER-"+phone, "user_"+phone, "$2a$10$hash", phone, "user", status, time.Now(), time.Now())
}

func TestGenerateCode(t *testing.T) {
	for i := 0; i < 20; i++ {
		code, err := generateCode(codeLength)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(code) != codeLength {
			t.Fatalf("code length = %d, want %d", len(code), codeLength)
		}
		for _, ch := range code {
			if ch < '0' || ch > '9' {
				t.Fatalf("code %q contains non-digit", code)
			}
		}
	}
}

func TestMaskPhone(t *testing.T) {
	if got := maskPhone("17612345678"); got != "176****5678" {
		t.Fatalf("maskPhone = %q, want 176****5678", got)
	}
	if got := maskPhone("123"); got != "***" {
		t.Fatalf("maskPhone short = %q, want ***", got)
	}
}

func TestSendCode_Success(t *testing.T) {
	svc, _, mr := newAuthService(t)
	ctx := context.Background()

	resp, err := svc.SendCode(ctx, testPhone, "login")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ExpiresIn != int(codeTTL.Seconds()) {
		t.Fatalf("expires_in = %d, want %d", resp.ExpiresIn, int(codeTTL.Seconds()))
	}
	code, err := mr.Get(codeKey("login", testPhone))
	if err != nil {
		t.Fatalf("code not stored in redis: %v", err)
	}
	if len(code) != codeLength {
		t.Fatalf("stored code = %q, want 6 digits", code)
	}
	if mr.TTL(codeKey("login", testPhone)) <= 0 {
		t.Fatal("code key should have positive TTL")
	}
}

func TestSendCode_InvalidInput(t *testing.T) {
	svc, _, _ := newAuthService(t)
	ctx := context.Background()

	if _, err := svc.SendCode(ctx, "12345", "login"); !errors.Is(err, ErrInvalidPhone) {
		t.Fatalf("err = %v, want ErrInvalidPhone", err)
	}
	if _, err := svc.SendCode(ctx, testPhone, "signup"); !errors.Is(err, ErrInvalidCodeType) {
		t.Fatalf("err = %v, want ErrInvalidCodeType", err)
	}
}

func TestSendCode_Cooldown(t *testing.T) {
	svc, _, _ := newAuthService(t)
	ctx := context.Background()

	if _, err := svc.SendCode(ctx, testPhone, "login"); err != nil {
		t.Fatalf("first send failed: %v", err)
	}
	_, err := svc.SendCode(ctx, testPhone, "login")
	if !errors.Is(err, ErrSendCooldown) {
		t.Fatalf("err = %v, want ErrSendCooldown", err)
	}
}

func TestSendCode_DailyLimit(t *testing.T) {
	svc, _, mr := newAuthService(t)
	ctx := context.Background()

	for i := 0; i < dailySendLimit; i++ {
		// 绕过冷却只验证每日上限
		mr.Del(cooldownKey(testPhone))
		if _, err := svc.SendCode(ctx, testPhone, "login"); err != nil {
			t.Fatalf("send #%d failed: %v", i+1, err)
		}
	}
	mr.Del(cooldownKey(testPhone))
	if _, err := svc.SendCode(ctx, testPhone, "login"); !errors.Is(err, ErrDailyLimit) {
		t.Fatalf("err = %v, want ErrDailyLimit", err)
	}
}

func TestVerifyCode_ReplayProtection(t *testing.T) {
	svc, _, mr := newAuthService(t)
	ctx := context.Background()

	mr.Set(codeKey("login", testPhone), "123456")
	if err := svc.verifyCode(ctx, testPhone, "login", "123456"); err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	// 成功后立即作废
	if err := svc.verifyCode(ctx, testPhone, "login", "123456"); !errors.Is(err, ErrCodeExpired) {
		t.Fatalf("err = %v, want ErrCodeExpired", err)
	}
}

func TestVerifyCode_TooManyAttempts(t *testing.T) {
	svc, _, mr := newAuthService(t)
	ctx := context.Background()

	mr.Set(codeKey("login", testPhone), "123456")
	for i := 0; i < maxVerifyFails; i++ {
		err := svc.verifyCode(ctx, testPhone, "login", "000000")
		if !errors.Is(err, ErrCodeInvalid) {
			t.Fatalf("attempt #%d err = %v, want ErrCodeInvalid", i+1, err)
		}
	}
	err := svc.verifyCode(ctx, testPhone, "login", "123456")
	if !errors.Is(err, ErrTooManyAttempts) {
		t.Fatalf("err = %v, want ErrTooManyAttempts", err)
	}
}

func TestSmsLogin_Success(t *testing.T) {
	svc, mock, mr := newAuthService(t)
	ctx := context.Background()

	mr.Set(codeKey("login", testPhone), "123456")
	mock.ExpectQuery("SELECT id, user_id, username").
		WithArgs(testPhone).
		WillReturnRows(userRows(testPhone, "1"))

	resp, err := svc.SmsLogin(ctx, testPhone, "123456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Token == "" || resp.User == nil || resp.User.UserID != "USER-"+testPhone {
		t.Fatalf("bad response: %+v", resp)
	}
	if resp.ExpiresIn != int(tokenTTL.Seconds()) {
		t.Fatalf("expires_in = %d, want %d", resp.ExpiresIn, int(tokenTTL.Seconds()))
	}
	// token 的 claims 应与账号登录结构一致
	claims := jwt.MapClaims{}
	if _, _, err := new(jwt.Parser).ParseUnverified(resp.Token, claims); err != nil {
		t.Fatalf("token unparseable: %v", err)
	}
	if claims["user_id"] != "USER-"+testPhone || claims["role"] != "user" {
		t.Fatalf("unexpected claims: %v", claims)
	}
}

func TestSmsLogin_UserNotFound(t *testing.T) {
	svc, mock, mr := newAuthService(t)
	ctx := context.Background()

	mr.Set(codeKey("login", testPhone), "123456")
	mock.ExpectQuery("SELECT id, user_id, username").
		WithArgs(testPhone).
		WillReturnError(sql.ErrNoRows)

	if _, err := svc.SmsLogin(ctx, testPhone, "123456"); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("err = %v, want ErrUserNotFound", err)
	}
}

func TestSmsLogin_Disabled(t *testing.T) {
	svc, mock, mr := newAuthService(t)
	ctx := context.Background()

	mr.Set(codeKey("login", testPhone), "123456")
	mock.ExpectQuery("SELECT id, user_id, username").
		WithArgs(testPhone).
		WillReturnRows(userRows(testPhone, "2"))

	if _, err := svc.SmsLogin(ctx, testPhone, "123456"); !errors.Is(err, ErrUserDisabled) {
		t.Fatalf("err = %v, want ErrUserDisabled", err)
	}
}

func TestRegister_Success(t *testing.T) {
	svc, mock, mr := newAuthService(t)
	ctx := context.Background()

	mr.Set(codeKey("register", testPhone), "123456")
	mock.ExpectQuery("SELECT COUNT").WithArgs(testPhone).
		WillReturnRows(sqlmock.NewRows([]string{"cnt"}).AddRow(0))
	mock.ExpectExec("INSERT INTO users").
		WithArgs("USER-"+testPhone, "user_"+testPhone, testPhone, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := svc.Register(ctx, testPhone, "123456", "secret123", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Token == "" || resp.User.UserID != "USER-"+testPhone {
		t.Fatalf("bad response: %+v", resp)
	}
	// 注册码一次性作废
	if mr.Exists(codeKey("register", testPhone)) {
		t.Fatal("register code should be consumed")
	}
}

func TestRegister_AlreadyExists(t *testing.T) {
	svc, mock, mr := newAuthService(t)
	ctx := context.Background()

	mr.Set(codeKey("register", testPhone), "123456")
	mock.ExpectQuery("SELECT COUNT").WithArgs(testPhone).
		WillReturnRows(sqlmock.NewRows([]string{"cnt"}).AddRow(1))

	if _, err := svc.Register(ctx, testPhone, "123456", "secret123", ""); !errors.Is(err, ErrUserAlreadyExists) {
		t.Fatalf("err = %v, want ErrUserAlreadyExists", err)
	}
}

func TestRegister_WeakPassword(t *testing.T) {
	svc, _, _ := newAuthService(t)
	if _, err := svc.Register(context.Background(), testPhone, "123456", "123", ""); !errors.Is(err, ErrInvalidPassword) {
		t.Fatalf("err = %v, want ErrInvalidPassword", err)
	}
}

// TestLogout_BlacklistFullChain 验证登出后 JWTAuth 中间件会拒绝该 token。
func TestLogout_BlacklistFullChain(t *testing.T) {
	svc, _, mr := newAuthService(t)
	ctx := context.Background()

	token, err := issueToken("USER-"+testPhone, "user_"+testPhone, "user")
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Logout(ctx, token); err != nil {
		t.Fatalf("logout failed: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() {
		middleware.InitJWTBlacklist(nil)
		rdb.Close()
	})
	middleware.InitJWTBlacklist(rdb)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/v1/ping", middleware.JWTAuth(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401 after logout", w.Code)
	}
	var body map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	if !strings.Contains(body["message"].(string), "revoked") {
		t.Fatalf("message = %v, want revoked message", body["message"])
	}
}

func TestHandler_ValidationError(t *testing.T) {
	svc, _, _ := newAuthService(t)
	h := NewHandler(svc)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h.RegisterPublic(r.Group("/api/v1"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/sms/send", strings.NewReader(`{"phone":"","type":"login"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK { // 项目约定：错误也返回 HTTP 200，code 在 body 中
		t.Fatalf("http status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"code":40001`) {
		t.Fatalf("body = %s, want code 40001", w.Body.String())
	}
}

func TestHandler_SendCooldownMapped(t *testing.T) {
	svc, _, mr := newAuthService(t)
	h := NewHandler(svc)

	mr.Set(cooldownKey(testPhone), "1") // 模拟刚发过

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h.RegisterPublic(r.Group("/api/v1"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/sms/send",
		strings.NewReader(`{"phone":"`+testPhone+`","type":"login"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if !strings.Contains(w.Body.String(), `"code":40111`) {
		t.Fatalf("body = %s, want code 40111", w.Body.String())
	}
}
