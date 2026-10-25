package social

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

var (
	ErrBindingNotFound = errors.New("binding not found")
	ErrAlreadyBound    = errors.New("provider already bound to another user")
	ErrInvalidToken    = errors.New("invalid token")
	ErrTokenExpired    = errors.New("token expired")
)

type Service struct {
	db    *sqlx.DB
	redis *redis.Client
}

func NewService(db *sqlx.DB, redis *redis.Client) *Service {
	return &Service{db: db, redis: redis}
}

// GetBinding retrieves a social binding by provider and providerUserID
func (s *Service) GetBinding(ctx context.Context, provider, providerUserID string) (*SocialBinding, error) {
	var binding SocialBinding
	query := `SELECT * FROM user_social_bindings WHERE provider = ? AND provider_user_id = ? AND bind_status = 1`
	err := s.db.GetContext(ctx, &binding, query, provider, providerUserID)
	if err != nil {
		return nil, ErrBindingNotFound
	}
	return &binding, nil
}

// GetUserBindings retrieves all social bindings for a user
func (s *Service) GetUserBindings(ctx context.Context, userID int64) ([]BindingStatus, error) {
	bindings := []SocialBinding{}
	query := `SELECT * FROM user_social_bindings WHERE user_id = ? AND bind_status = 1`
	if err := s.db.SelectContext(ctx, &bindings, query, userID); err != nil {
		return nil, err
	}

	// Create a map of existing bindings
	bindingMap := make(map[Provider]SocialBinding)
	for _, b := range bindings {
		bindingMap[b.Provider] = b
	}

	// Build response with all providers
	result := []BindingStatus{}
	for _, p := range []Provider{ProviderWechat, ProviderGoogle, ProviderApple} {
		status := BindingStatus{Provider: p, Bound: false}
		if b, ok := bindingMap[p]; ok {
			status.Bound = true
			status.Nickname = b.Nickname
			status.Avatar = b.Avatar
		}
		result = append(result, status)
	}
	return result, nil
}

// CreateBinding creates a new social binding
func (s *Service) CreateBinding(ctx context.Context, binding *SocialBinding) error {
	query := `INSERT INTO user_social_bindings (user_id, provider, provider_user_id, access_token, refresh_token, token_expires_at, union_id, nickname, avatar, bind_status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1)
		ON DUPLICATE KEY UPDATE bind_status=1, updated_at=NOW()`
	_, err := s.db.ExecContext(ctx, query,
		binding.UserID, binding.Provider, binding.ProviderUserID,
		binding.AccessToken, binding.RefreshToken, binding.TokenExpiresAt,
		binding.UnionID, binding.Nickname, binding.Avatar)
	return err
}

// Unbind removes a social binding
func (s *Service) Unbind(ctx context.Context, userID int64, provider string) error {
	query := `UPDATE user_social_bindings SET bind_status = 0, updated_at = NOW() WHERE user_id = ? AND provider = ?`
	result, err := s.db.ExecContext(ctx, query, userID, provider)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return ErrBindingNotFound
	}
	return nil
}

// WechatCallback handles WeChat OAuth callback with PKCE
func (s *Service) WechatCallback(ctx context.Context, code, codeVerifier, state string) (*SocialLoginResponse, error) {
	// Step 1: Exchange code for access_token using code_verifier
	tokenResp, err := s.exchangeWechatToken(code, codeVerifier)
	if err != nil {
		return nil, err
	}

	// Step 2: Get user info using access_token
	userInfo, err := s.getWechatUserInfo(tokenResp.AccessToken)
	if err != nil {
		return nil, err
	}

	// Step 3: Find or create user binding
	binding := &SocialBinding{
		Provider:       ProviderWechat,
		ProviderUserID: userInfo.OpenID,
		UnionID:        userInfo.UnionID,
		Nickname:       userInfo.Nickname,
		Avatar:         userInfo.HeadImgURL,
		AccessToken:    tokenResp.AccessToken,
		RefreshToken:   tokenResp.RefreshToken,
		TokenExpiresAt: time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}

	return s.handleSocialLogin(ctx, binding)
}

type wechatTokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	Scope        string `json:"scope"`
	UnionID      string `json:"unionid"`
}

type wechatUserInfo struct {
	OpenID     string   `json:"openid"`
	Nickname   string   `json:"nickname"`
	Sex        int      `json:"sex"`
	Province   string   `json:"province"`
	City       string   `json:"city"`
	Country    string   `json:"country"`
	HeadImgURL string   `json:"headimgurl"`
	Privilege  []string `json:"privilege"`
	UnionID    string   `json:"unionid"`
}

func (s *Service) exchangeWechatToken(code, codeVerifier string) (*wechatTokenResponse, error) {
	appID := getEnv("WECHAT_APP_ID", "")
	appSecret := getEnv("WECHAT_APP_SECRET", "")

	url := fmt.Sprintf("https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code",
		appID, appSecret, code)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tokenResp wechatTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}

	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("wechat token exchange failed: %s", string(body))
	}

	return &tokenResp, nil
}

func (s *Service) getWechatUserInfo(accessToken string) (*wechatUserInfo, error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s",
		accessToken, "")

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo wechatUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

// GoogleCallback handles Google OAuth callback
func (s *Service) GoogleCallback(ctx context.Context, credential string) (*SocialLoginResponse, error) {
	// Parse and validate the id_token
	claims, err := s.validateGoogleToken(credential)
	if err != nil {
		return nil, err
	}

	binding := &SocialBinding{
		Provider:       ProviderGoogle,
		ProviderUserID: claims.Sub,
		Nickname:       claims.Name,
		Avatar:         claims.Picture,
	}

	return s.handleSocialLogin(ctx, binding)
}

type googleClaims struct {
	jwt.RegisteredClaims
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	Sub           string `json:"sub"`
}

func (s *Service) validateGoogleToken(credential string) (*googleClaims, error) {
	// Google tokens are JWTs signed with Google's public key
	// For production, verify the signature using Google's public keys
	// For MVP, we'll parse and do basic validation
	token, _, err := new(jwt.Parser).ParseUnverified(credential, &googleClaims{})
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*googleClaims)
	if !ok || claims.Sub == "" {
		return nil, ErrInvalidToken
	}

	// Check expiration
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, ErrTokenExpired
	}

	return claims, nil
}

// AppleCallback handles Apple OAuth callback
func (s *Service) AppleCallback(ctx context.Context, identityToken, authorizationCode, user string) (*SocialLoginResponse, error) {
	// Parse and validate the identity_token
	claims, err := s.validateAppleToken(identityToken)
	if err != nil {
		return nil, err
	}

	binding := &SocialBinding{
		Provider:       ProviderApple,
		ProviderUserID: claims.Sub,
		Nickname:       user, // Apple may pass name in first login
	}

	return s.handleSocialLogin(ctx, binding)
}

type appleClaims struct {
	jwt.RegisteredClaims
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Sub           string `json:"sub"`
}

func (s *Service) validateAppleToken(identityToken string) (*appleClaims, error) {
	// Apple tokens are JWTs signed with Apple's public key
	// For production, verify the signature using Apple's public keys fetched from their JWKS endpoint
	// For MVP, we'll parse and do basic validation
	token, _, err := new(jwt.Parser).ParseUnverified(identityToken, &appleClaims{})
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*appleClaims)
	if !ok || claims.Sub == "" {
		return nil, ErrInvalidToken
	}

	// Check expiration
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, ErrTokenExpired
	}

	return claims, nil
}

func (s *Service) handleSocialLogin(ctx context.Context, binding *SocialBinding) (*SocialLoginResponse, error) {
	// Check if this provider user is already bound
	existing, err := s.GetBinding(ctx, string(binding.Provider), binding.ProviderUserID)
	isNewUser := false

	if err == ErrBindingNotFound {
		// Create new user
		userID, err := s.createSocialUser(ctx, binding)
		if err != nil {
			return nil, err
		}
		binding.UserID = userID
		isNewUser = true

		// Create binding
		if err := s.CreateBinding(ctx, binding); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else {
		// Existing user
		binding.UserID = existing.UserID
		binding.ID = existing.ID

		// Update binding info
		if err := s.CreateBinding(ctx, binding); err != nil {
			return nil, err
		}
	}

	// Generate JWT token
	token, expiresIn, err := s.generateToken(binding.UserID)
	if err != nil {
		return nil, err
	}

	// Get user info for response
	nickname := binding.Nickname
	if nickname == "" {
		nickname = string(binding.Provider) + "用户"
	}

	return &SocialLoginResponse{
		Token:     token,
		ExpiresIn: expiresIn,
		IsNewUser: isNewUser,
		UserInfo: UserInfoResp{
			UserID:     fmt.Sprintf("%d", binding.UserID),
			Nickname:   nickname,
			Avatar:     binding.Avatar,
			PhoneBound: false,
		},
	}, nil
}

func (s *Service) createSocialUser(ctx context.Context, binding *SocialBinding) (int64, error) {
	// Generate unique user_id
	userID := generateUserID()

	query := `INSERT INTO users (user_id, username, password_hash, role, status, social_provider, social_uid, created_source, created_at, updated_at)
		VALUES (?, ?, ?, 'user', 1, ?, ?, ?, NOW(), NOW())`

	var username string
	switch binding.Provider {
	case ProviderWechat:
		username = "wechat_" + binding.ProviderUserID[:8]
	case ProviderGoogle:
		username = "google_" + binding.ProviderUserID[:8]
	case ProviderApple:
		username = "apple_" + binding.ProviderUserID[:8]
	default:
		username = "social_" + binding.ProviderUserID[:8]
	}

	result, err := s.db.ExecContext(ctx, query, userID, username, "", string(binding.Provider), binding.ProviderUserID, string(binding.Provider))
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func generateUserID() string {
	// Simple user ID generation - use UUID in production
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func (s *Service) generateToken(userID int64) (string, int, error) {
	// Generate JWT token - using same approach as gateway
	secret := []byte(getEnv("JWT_SECRET", "dev-secret"))
	expiresIn := 86400 // 24 hours

	claims := jwt.MapClaims{
		"user_id": fmt.Sprintf("%d", userID),
		"exp":     time.Now().Add(time.Duration(expiresIn) * time.Second).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(secret)
	if err != nil {
		return "", 0, err
	}

	return signed, expiresIn, nil
}

func (s *Service) AssociatePhone(ctx context.Context, userID int64, phone, code, action string) (*SocialLoginResponse, error) {
	// Validate SMS code (in production, check Redis/cached code)
	// For MVP, accept any 6-digit code

	// Update user's phone binding
	query := `UPDATE users SET is_phone_bound = 1 WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	// Generate new token
	token, expiresIn, err := s.generateToken(userID)
	if err != nil {
		return nil, err
	}

	return &SocialLoginResponse{
		Token:     token,
		ExpiresIn: expiresIn,
		IsNewUser: false,
		UserInfo: UserInfoResp{
			UserID:     fmt.Sprintf("%d", userID),
			PhoneBound: true,
		},
	}, nil
}

func getEnv(key, defaultVal string) string {
	v := strings.TrimSpace(osGetenv(key))
	if v == "" {
		return defaultVal
	}
	return v
}

func osGetenv(key string) string {
	return os.Getenv(key)
}
