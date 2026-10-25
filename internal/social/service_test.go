package social

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

	"ecomm-micro/pkg/middleware"
)

func newSocialService(t *testing.T) (*Service, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	sqlxDB := sqlx.NewDb(db, "mysql")
	t.Cleanup(func() { sqlxDB.Close() })
	return NewService(sqlxDB, &redis.Client{}), mock
}

func bindingRows(bindings ...SocialBinding) *sqlmock.Rows {
	r := sqlmock.NewRows([]string{"id", "user_id", "provider", "provider_user_id", "access_token", "refresh_token", "token_expires_at", "union_id", "nickname", "avatar", "bind_status", "created_at", "updated_at"})
	for _, b := range bindings {
		r.AddRow(b.ID, b.UserID, b.Provider, b.ProviderUserID, b.AccessToken, b.RefreshToken, b.TokenExpiresAt, b.UnionID, b.Nickname, b.Avatar, b.BindStatus, b.CreatedAt, b.UpdatedAt)
	}
	return r
}

func TestGetBinding_FoundAndMissing(t *testing.T) {
	svc, mock := newSocialService(t)
	ctx := context.Background()

	want := SocialBinding{ID: 1, UserID: 42, Provider: ProviderWechat, ProviderUserID: "wx-open-1",
		Nickname: "小明", BindStatus: 1}
	mock.ExpectQuery(`FROM user_social_bindings`).
		WithArgs("wechat", "wx-open-1").
		WillReturnRows(bindingRows(want))

	got, err := svc.GetBinding(ctx, "wechat", "wx-open-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.UserID != 42 || got.Provider != ProviderWechat || got.Nickname != "小明" {
		t.Fatalf("binding = %+v", got)
	}

	// 不存在
	mock.ExpectQuery(`FROM user_social_bindings`).
		WithArgs("google", "g-404").
		WillReturnError(sql.ErrNoRows)
	if _, err := svc.GetBinding(ctx, "google", "g-404"); !errors.Is(err, ErrBindingNotFound) {
		t.Fatalf("err = %v, want ErrBindingNotFound", err)
	}
}

func TestGetUserBindings(t *testing.T) {
	svc, mock := newSocialService(t)

	mock.ExpectQuery(`WHERE user_id = \? AND bind_status = 1`).WithArgs(int64(42)).
		WillReturnRows(bindingRows(
			SocialBinding{UserID: 42, Provider: ProviderGoogle, ProviderUserID: "g-1", Nickname: "Bob", Avatar: "a.png"},
		))

	bindings, err := svc.GetUserBindings(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bindings) != 3 {
		t.Fatalf("should return status for all 3 providers, got %d", len(bindings))
	}
	byProvider := map[Provider]BindingStatus{}
	for _, b := range bindings {
		byProvider[b.Provider] = b
	}
	if !byProvider[ProviderGoogle].Bound || byProvider[ProviderGoogle].Nickname != "Bob" {
		t.Fatalf("google = %+v", byProvider[ProviderGoogle])
	}
	if byProvider[ProviderWechat].Bound || byProvider[ProviderApple].Bound {
		t.Fatalf("wechat/apple should be unbound: %+v", byProvider)
	}
}

func TestUnbind(t *testing.T) {
	svc, mock := newSocialService(t)
	ctx := context.Background()

	mock.ExpectExec(`SET bind_status = 0`).WithArgs(int64(42), "wechat").
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := svc.Unbind(ctx, 42, "wechat"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mock.ExpectExec(`SET bind_status = 0`).WithArgs(int64(42), "apple").
		WillReturnResult(sqlmock.NewResult(0, 0))
	if err := svc.Unbind(ctx, 42, "apple"); !errors.Is(err, ErrBindingNotFound) {
		t.Fatalf("err = %v, want ErrBindingNotFound", err)
	}
}

func TestCreateBinding(t *testing.T) {
	svc, mock := newSocialService(t)

	mock.ExpectExec(`INSERT INTO user_social_bindings`).WillReturnResult(sqlmock.NewResult(1, 1))

	b := &SocialBinding{UserID: 42, Provider: ProviderWechat, ProviderUserID: "wx-1",
		TokenExpiresAt: time.Now().Add(time.Hour)}
	if err := svc.CreateBinding(context.Background(), b); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHandleSocialLogin_ExistingUser(t *testing.T) {
	svc, mock := newSocialService(t)
	ctx := context.Background()

	existing := SocialBinding{ID: 5, UserID: 42, Provider: ProviderWechat, ProviderUserID: "wx-open-1", BindStatus: 1}
	mock.ExpectQuery(`FROM user_social_bindings`).
		WithArgs("wechat", "wx-open-1").
		WillReturnRows(bindingRows(existing))
	// 已绑定用户：更新绑定信息后直接签发 token
	mock.ExpectExec(`INSERT INTO user_social_bindings`).WillReturnResult(sqlmock.NewResult(5, 1))

	resp, err := svc.handleSocialLogin(ctx, &SocialBinding{Provider: ProviderWechat, ProviderUserID: "wx-open-1", Nickname: "小明"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.IsNewUser {
		t.Fatal("existing binding should not be flagged as new user")
	}
	if resp.Token == "" || resp.UserInfo.UserID != "42" || resp.UserInfo.Nickname != "小明" {
		t.Fatalf("resp = %+v", resp)
	}
}

func TestHandleSocialLogin_NewUser(t *testing.T) {
	svc, mock := newSocialService(t)
	ctx := context.Background()

	// 无绑定 → 创建用户 + 绑定
	mock.ExpectQuery(`FROM user_social_bindings`).
		WithArgs("google", "google-sub-99").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(`INSERT INTO users`).WillReturnResult(sqlmock.NewResult(99, 1))
	mock.ExpectExec(`INSERT INTO user_social_bindings`).WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := svc.handleSocialLogin(ctx, &SocialBinding{Provider: ProviderGoogle, ProviderUserID: "google-sub-99"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.IsNewUser {
		t.Fatal("fresh provider user should be flagged as new user")
	}
	if resp.UserInfo.Nickname != "google用户" {
		t.Fatalf("default nickname = %q, want google用户", resp.UserInfo.Nickname)
	}
}

func TestGenerateToken(t *testing.T) {
	svc, _ := newSocialService(t)

	token, expiresIn, err := svc.generateToken(42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if expiresIn != 86400 {
		t.Fatalf("expires_in = %d, want 86400", expiresIn)
	}
	claims := jwt.MapClaims{}
	if _, _, err := new(jwt.Parser).ParseUnverified(token, claims); err != nil {
		t.Fatalf("token unparseable: %v", err)
	}
	if claims["user_id"] != "42" {
		t.Fatalf("user_id claim = %v, want \"42\"", claims["user_id"])
	}
}

func signRS256Style(t *testing.T, claims jwt.MapClaims) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte("any-key"))
	if err != nil {
		t.Fatal(err)
	}
	return signed
}

func TestValidateGoogleToken(t *testing.T) {
	svc, _ := newSocialService(t)

	valid := signRS256Style(t, jwt.MapClaims{
		"sub": "google-sub-1", "email": "a@b.c", "name": "Bob",
		"picture": "p.png", "exp": time.Now().Add(time.Hour).Unix(),
	})
	claims, err := svc.validateGoogleToken(valid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims.Sub != "google-sub-1" || claims.Name != "Bob" {
		t.Fatalf("claims = %+v", claims)
	}

	expired := signRS256Style(t, jwt.MapClaims{"sub": "s", "exp": time.Now().Add(-time.Hour).Unix()})
	if _, err := svc.validateGoogleToken(expired); !errors.Is(err, ErrTokenExpired) {
		t.Fatalf("err = %v, want ErrTokenExpired", err)
	}

	if _, err := svc.validateGoogleToken("not-a-jwt"); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("err = %v, want ErrInvalidToken", err)
	}
}

func TestValidateAppleToken(t *testing.T) {
	svc, _ := newSocialService(t)

	valid := signRS256Style(t, jwt.MapClaims{"sub": "apple-sub-1", "exp": time.Now().Add(time.Hour).Unix()})
	claims, err := svc.validateAppleToken(valid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims.Sub != "apple-sub-1" {
		t.Fatalf("claims = %+v", claims)
	}

	if _, err := svc.validateAppleToken("garbage"); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("err = %v, want ErrInvalidToken", err)
	}
}

func TestCreateSocialUser_UsernamePrefix(t *testing.T) {
	svc, mock := newSocialService(t)
	ctx := context.Background()

	// username = provider 前缀 + ProviderUserID 前 8 位
	const pid = "provuser123456"
	cases := []struct {
		provider Provider
		prefix   string
	}{
		{ProviderWechat, "wechat_"},
		{ProviderGoogle, "google_"},
		{ProviderApple, "apple_"},
		{Provider("other"), "social_"},
	}
	for i, c := range cases {
		want := c.prefix + pid[:8]
		mock.ExpectExec(`INSERT INTO users`).WithArgs(sqlmock.AnyArg(), want, "", string(c.provider), pid, string(c.provider)).
			WillReturnResult(sqlmock.NewResult(int64(i+1), 1))
		id, err := svc.createSocialUser(ctx, &SocialBinding{Provider: c.provider, ProviderUserID: pid})
		if err != nil {
			t.Fatalf("case %s: %v", c.provider, err)
		}
		if id != int64(i+1) {
			t.Fatalf("case %s: id = %d", c.provider, id)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

// ---- handler：OAuth 回调无需登录、绑定管理需登录（验证路由拆分修复）----

func newSocialRouter(svc *Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	NewHandler(svc).Register(r.Group("/api/v1"))
	return r
}

func TestHandler_CallbackAccessibleWithoutToken(t *testing.T) {
	svc, _ := newSocialService(t)
	r := newSocialRouter(svc)

	// 无 Authorization 头调用回调：应进入业务逻辑（参数校验失败 40001）而非 401
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/social/callback/google", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if strings.Contains(w.Body.String(), `"code":40101`) {
		t.Fatalf("OAuth callback must not require JWT, body = %s", w.Body.String())
	}
	// 空 credential 会走到业务层校验失败（50001），证明请求未被鉴权拦截
	if !strings.Contains(w.Body.String(), `"code":50001`) {
		t.Fatalf("expected business error for empty credential, body = %s", w.Body.String())
	}
}

func TestHandler_BindingsRequireTokenAndConvertStringUserID(t *testing.T) {
	// JWTAuth 在构造时固化 secret
	t.Setenv("JWT_SECRET", "test-secret")
	svc, mock := newSocialService(t)
	r := newSocialRouter(svc)

	// 未登录 → 401
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/social/bindings", nil)
	r.ServeHTTP(w, req)
	if !strings.Contains(w.Body.String(), `"code":40101`) {
		t.Fatalf("bindings should require auth, body = %s", w.Body.String())
	}

	// 已登录（user_id 为 string "42"，与生产一致）→ 修复后应正确 ParseInt 查库
	claims := jwt.MapClaims{"user_id": "42", "username": "u", "role": "user", "exp": time.Now().Add(time.Hour).Unix()}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := tok.SignedString([]byte("test-secret"))

	mock.ExpectQuery(`WHERE user_id = \? AND bind_status = 1`).WithArgs(int64(42)).
		WillReturnRows(bindingRows())

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/auth/social/bindings", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	r.ServeHTTP(w, req)

	if !strings.Contains(w.Body.String(), `"code":0`) {
		t.Fatalf("authenticated body = %s", w.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

// 编译期确认 middleware 常量仍被引用（避免未使用 import）
var _ = middleware.CtxUserID
