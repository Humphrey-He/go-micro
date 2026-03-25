package user

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"go-micro/pkg/cache"
	"golang.org/x/crypto/bcrypt"
)

var ErrNotFound = errors.New("user not found")

const defaultStatus = 1

type Service struct {
	ctx   context.Context
	db    *sqlx.DB
	cache *cacheClient
}

func NewService(dbx *sqlx.DB, rdb *redis.Client) *Service {
	return &Service{
		ctx:   context.Background(),
		db:    dbx,
		cache: &cacheClient{rdb: rdb},
	}
}

func (s *Service) Create(req CreateUserRequest) (*User, error) {
	if req.UserID == "" {
		req.UserID = uuid.NewString()
	}
	if req.Username == "" || req.Password == "" {
		return nil, errors.New("invalid request")
	}
	if req.Role == "" {
		req.Role = "user"
	}
	pwdHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	_, err = s.db.Exec(`INSERT INTO users(user_id,username,password_hash,role,status,created_at,updated_at) VALUES(?,?,?,?,?,NOW(),NOW())`,
		req.UserID, req.Username, string(pwdHash), req.Role, defaultStatus)
	if err != nil {
		return nil, err
	}
	user := &User{UserID: req.UserID, Username: req.Username, Role: req.Role, Status: defaultStatus, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	_ = s.cache.setUser(s.ctx, req.UserID, user)
	_ = s.cache.setUserByName(s.ctx, req.Username, user)
	return user, nil
}

func (s *Service) Get(userID string) (*User, error) {
	if userID == "" {
		return nil, ErrNotFound
	}
	if u, ok := s.cache.getUser(s.ctx, userID); ok {
		return u, nil
	}
	lockKey := "lock:user:" + userID
	locked, _ := cache.TryLock(s.ctx, s.cache.rdb, lockKey, 5*time.Second)
	if locked {
		defer func() { _ = cache.Unlock(s.ctx, s.cache.rdb, lockKey) }()
		if u, ok := s.cache.getUser(s.ctx, userID); ok {
			return u, nil
		}
	}
	user := User{}
	if err := s.db.Get(&user, `SELECT * FROM users WHERE user_id = ?`, userID); err != nil {
		if err == sql.ErrNoRows {
			_ = s.cache.setNil(s.ctx, userID)
			return nil, ErrNotFound
		}
		return nil, err
	}
	_ = s.cache.setUser(s.ctx, userID, &user)
	_ = s.cache.setUserByName(s.ctx, user.Username, &user)
	return &user, nil
}

func (s *Service) GetByUsername(username string) (*User, error) {
	if username == "" {
		return nil, ErrNotFound
	}
	if u, ok := s.cache.getUserByName(s.ctx, username); ok {
		return u, nil
	}
	lockKey := "lock:user:uname:" + username
	locked, _ := cache.TryLock(s.ctx, s.cache.rdb, lockKey, 5*time.Second)
	if locked {
		defer func() { _ = cache.Unlock(s.ctx, s.cache.rdb, lockKey) }()
		if u, ok := s.cache.getUserByName(s.ctx, username); ok {
			return u, nil
		}
	}
	user := User{}
	if err := s.db.Get(&user, `SELECT * FROM users WHERE username = ?`, username); err != nil {
		if err == sql.ErrNoRows {
			_ = s.cache.setNilByName(s.ctx, username)
			return nil, ErrNotFound
		}
		return nil, err
	}
	_ = s.cache.setUserByName(s.ctx, username, &user)
	return &user, nil
}

func (s *Service) VerifyPassword(user *User, password string) bool {
	if user == nil {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) == nil
}

type cacheClient struct {
	rdb *redis.Client
}

func (c *cacheClient) getUser(ctx context.Context, userID string) (*User, bool) {
	key := "user:" + userID
	var u User
	hit, isNil, err := cache.GetJSON(ctx, c.rdb, key, &u)
	if err != nil || isNil || !hit {
		return nil, false
	}
	if u.UserID == "" {
		return nil, false
	}
	return &u, true
}

func (c *cacheClient) setUser(ctx context.Context, userID string, u *User) error {
	key := "user:" + userID
	return cache.SetJSON(ctx, c.rdb, key, u, ttlWithJitter(5*time.Minute, 30*time.Second))
}

func (c *cacheClient) setNil(ctx context.Context, userID string) error {
	key := "user:" + userID
	return cache.SetNil(ctx, c.rdb, key, ttlWithJitter(30*time.Second, 10*time.Second))
}

func (c *cacheClient) getUserByName(ctx context.Context, username string) (*User, bool) {
	key := "user:uname:" + username
	var u User
	hit, isNil, err := cache.GetJSON(ctx, c.rdb, key, &u)
	if err != nil || isNil || !hit {
		return nil, false
	}
	if u.UserID == "" {
		return nil, false
	}
	return &u, true
}

func (c *cacheClient) setUserByName(ctx context.Context, username string, u *User) error {
	key := "user:uname:" + username
	return cache.SetJSON(ctx, c.rdb, key, u, ttlWithJitter(5*time.Minute, 30*time.Second))
}

func (c *cacheClient) setNilByName(ctx context.Context, username string) error {
	key := "user:uname:" + username
	return cache.SetNil(ctx, c.rdb, key, ttlWithJitter(30*time.Second, 10*time.Second))
}

func ttlWithJitter(base, jitter time.Duration) time.Duration {
	return base + time.Duration(time.Now().UnixNano()%int64(jitter))
}
