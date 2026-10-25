package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"go-micro/internal/user"
	"go-micro/pkg/config"
	"go-micro/pkg/logx"
	"go-micro/pkg/middleware"
)

const (
	codeTTL         = 5 * time.Minute // 验证码有效期
	codeLength      = 6
	sendCooldown    = 60 * time.Second // 同一手机号两次发送的最小间隔
	dailySendLimit  = 10               // 同一手机号每日发送上限
	maxVerifyFails  = 5                // 验证码错误尝试上限，超过后作废当次验证码
	tokenTTL        = 2 * time.Hour    // 与账号密码登录的 JWT 有效期保持一致
	minPasswordLen  = 6
	maxPasswordLen  = 64
	maxInviteCodeLn = 32
)

var (
	ErrInvalidPhone      = errors.New("invalid phone number")
	ErrInvalidCodeType   = errors.New("invalid sms code type")
	ErrSendCooldown      = errors.New("sms send cooldown")
	ErrDailyLimit        = errors.New("sms daily send limit exceeded")
	ErrCodeInvalid       = errors.New("sms code invalid")
	ErrCodeExpired       = errors.New("sms code expired")
	ErrTooManyAttempts   = errors.New("too many verify attempts")
	ErrUserNotFound      = errors.New("user not found")
	ErrUserDisabled      = errors.New("user account is disabled")
	ErrUserAlreadyExists = errors.New("phone already registered")
	ErrInvalidPassword   = errors.New("invalid password format")
	ErrInvalidInviteCode = errors.New("invalid invite code")
)

var phonePattern = regexp.MustCompile(`^1[3-9]\d{9}$`)

var validCodeTypes = map[string]bool{"login": true, "register": true, "reset_password": true}

// Service 短信验证码认证：验证码收发校验、短信登录、注册、登出（token 黑名单）。
// 挂载在 gateway 进程内，直连 MySQL 与 Redis，与 social/pricewatch 包同一模式。
type Service struct {
	db     *sqlx.DB
	rdb    *redis.Client
	logger *zap.Logger
}

func NewService(dbx *sqlx.DB, rdb *redis.Client) *Service {
	return &Service{db: dbx, rdb: rdb, logger: logx.L()}
}

func codeKey(typ, phone string) string  { return "sms:code:" + typ + ":" + phone }
func failKey(typ, phone string) string  { return "sms:fail:" + typ + ":" + phone }
func cooldownKey(phone string) string   { return "sms:cooldown:" + phone }
func dailyKey(phone string) string      { return "sms:daily:" + time.Now().Format("20060102") + ":" + phone }

// SendCode 生成验证码并写入 Redis，附带发送冷却与每日上限防刷。
// 未配置 SMS_PROVIDER 时为开发模式：验证码不真实下发，输出到服务日志；
// 可通过 SMS_DEBUG_CODE 固定验证码便于本地联调。
func (s *Service) SendCode(ctx context.Context, phone, typ string) (*SendSmsResponse, error) {
	if !phonePattern.MatchString(phone) {
		return nil, ErrInvalidPhone
	}
	if !validCodeTypes[typ] {
		return nil, ErrInvalidCodeType
	}

	// 60s 发送冷却
	ok, err := s.rdb.SetNX(ctx, cooldownKey(phone), 1, sendCooldown).Result()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrSendCooldown
	}

	// 每日发送上限
	cnt, err := s.rdb.Incr(ctx, dailyKey(phone)).Result()
	if err != nil {
		return nil, err
	}
	if cnt == 1 {
		s.rdb.Expire(ctx, dailyKey(phone), 24*time.Hour)
	}
	if cnt > dailySendLimit {
		return nil, ErrDailyLimit
	}

	code := config.GetEnv("SMS_DEBUG_CODE", "")
	if code == "" {
		code, err = generateCode(codeLength)
		if err != nil {
			return nil, err
		}
	}
	if err := s.rdb.Set(ctx, codeKey(typ, phone), code, codeTTL).Err(); err != nil {
		return nil, err
	}
	s.rdb.Del(ctx, failKey(typ, phone))

	if config.GetEnv("SMS_PROVIDER", "") == "" {
		s.logger.Info("sms code sent (dev mode: no SMS_PROVIDER configured, code logged below)",
			zap.String("phone", maskPhone(phone)), zap.String("type", typ), zap.String("code", code))
	} else {
		// TODO: 接入真实短信服务商（阿里云/腾讯云 SMS），当前仅入库待发送
		s.logger.Warn("SMS_PROVIDER configured but vendor integration not implemented; code not delivered",
			zap.String("phone", maskPhone(phone)), zap.String("type", typ))
	}

	return &SendSmsResponse{ExpiresIn: int(codeTTL.Seconds())}, nil
}

// verifyCode 校验验证码；成功后立即作废，防止重放。
func (s *Service) verifyCode(ctx context.Context, phone, typ, code string) error {
	if fails, _ := s.rdb.Get(ctx, failKey(typ, phone)).Int(); fails >= maxVerifyFails {
		s.rdb.Del(ctx, codeKey(typ, phone))
		return ErrTooManyAttempts
	}
	stored, err := s.rdb.Get(ctx, codeKey(typ, phone)).Result()
	if errors.Is(err, redis.Nil) {
		return ErrCodeExpired
	}
	if err != nil {
		return err
	}
	if stored != code {
		s.rdb.Incr(ctx, failKey(typ, phone))
		s.rdb.Expire(ctx, failKey(typ, phone), codeTTL)
		return ErrCodeInvalid
	}
	s.rdb.Del(ctx, codeKey(typ, phone), failKey(typ, phone))
	return nil
}

// SmsLogin 短信验证码登录：校验验证码后按手机号查找用户并签发 JWT。
func (s *Service) SmsLogin(ctx context.Context, phone, code string) (*LoginResponse, error) {
	if !phonePattern.MatchString(phone) {
		return nil, ErrInvalidPhone
	}
	if err := s.verifyCode(ctx, phone, "login", code); err != nil {
		return nil, err
	}

	u, err := s.getUserByPhone(ctx, phone)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	if u.Status != 1 {
		return nil, ErrUserDisabled
	}

	token, err := issueToken(u.UserID, u.Username, u.Role)
	if err != nil {
		return nil, err
	}
	return &LoginResponse{Token: token, ExpiresIn: int(tokenTTL.Seconds()), User: u}, nil
}

// Register 短信验证码注册：校验验证码、确保手机号未注册后创建用户并直接登录。
func (s *Service) Register(ctx context.Context, phone, code, password, inviteCode string) (*LoginResponse, error) {
	if !phonePattern.MatchString(phone) {
		return nil, ErrInvalidPhone
	}
	if len(password) < minPasswordLen || len(password) > maxPasswordLen {
		return nil, ErrInvalidPassword
	}
	if len(inviteCode) > maxInviteCodeLn {
		return nil, ErrInvalidInviteCode
	}
	if err := s.verifyCode(ctx, phone, "register", code); err != nil {
		return nil, err
	}

	var exists int
	if err := s.db.GetContext(ctx, &exists, `SELECT COUNT(1) FROM users WHERE phone = ?`, phone); err != nil {
		return nil, err
	}
	if exists > 0 {
		return nil, ErrUserAlreadyExists
	}

	pwdHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 用户编号与登录名的生成规则和 006 迁移脚本中的种子用户保持一致
	userID := "USER-" + phone
	username := "user_" + phone
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO users (user_id, username, phone, password_hash, role, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, 'user', 1, NOW(), NOW())`,
		userID, username, phone, string(pwdHash))
	if err != nil {
		return nil, err
	}

	token, err := issueToken(userID, username, "user")
	if err != nil {
		return nil, err
	}
	u := &user.User{UserID: userID, Username: username, Role: "user", Status: 1}
	u.Phone.String = phone
	u.Phone.Valid = true
	return &LoginResponse{Token: token, ExpiresIn: int(tokenTTL.Seconds()), User: u}, nil
}

// Logout 将 token 写入黑名单，剩余有效期内拒绝该 token 的所有请求。
func (s *Service) Logout(ctx context.Context, tokenString string) error {
	claims := jwt.MapClaims{}
	_, _, err := new(jwt.Parser).ParseUnverified(tokenString, claims)
	if err != nil {
		return ErrCodeInvalid
	}
	ttl := time.Duration(0)
	if exp, ok := claims["exp"].(float64); ok {
		if remain := time.Until(time.Unix(int64(exp), 0)); remain > 0 {
			ttl = remain
		}
	}
	if ttl == 0 {
		return nil // 已过期的 token 无需拉黑
	}
	sum := sha256.Sum256([]byte(tokenString))
	key := middleware.TokenBlacklistKey(hex.EncodeToString(sum[:]))
	return s.rdb.Set(ctx, key, 1, ttl).Err()
}

func (s *Service) getUserByPhone(ctx context.Context, phone string) (*user.User, error) {
	var u user.User
	// 列集限定为两套 schema（基础表 + V2）共有的核心列
	err := s.db.GetContext(ctx, &u,
		`SELECT id, user_id, username, password_hash, phone, role, status, created_at, updated_at
		 FROM users WHERE phone = ?`, phone)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// issueToken 签发与账号密码登录完全相同 claims 结构的 JWT。
func issueToken(userID, username, role string) (string, error) {
	secret := []byte(config.GetEnv("JWT_SECRET", "dev-secret"))
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"role":     role,
		"exp":      time.Now().Add(tokenTTL).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
}

func generateCode(n int) (string, error) {
	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n)), nil)
	v, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%0"+strconv.Itoa(n)+"d", v.Int64()), nil
}

// maskPhone 日志脱敏：保留前 3 位与后 4 位
func maskPhone(phone string) string {
	if len(phone) != 11 {
		return "***"
	}
	return phone[:3] + "****" + phone[7:]
}
