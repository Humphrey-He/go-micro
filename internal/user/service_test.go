package user

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

func newUserService(t *testing.T) (*Service, sqlmock.Sqlmock, *miniredis.Miniredis) {
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

	return NewService(sqlxDB, rdb), mock, mr
}

func TestCreate_Success(t *testing.T) {
	svc, mock, mr := newUserService(t)
	defer mr.Close()

	req := CreateUserRequest{
		UserID:   "U-1",
		Username: "testuser",
		Password: "password123",
		Role:     "user",
	}

	mock.ExpectExec(`INSERT INTO users`).
		WithArgs(req.UserID, req.Username, sqlmock.AnyArg(), req.Role, defaultStatus).
		WillReturnResult(sqlmock.NewResult(1, 1))

	user, err := svc.Create(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.UserID != req.UserID {
		t.Fatalf("expected user_id %s, got %s", req.UserID, user.UserID)
	}
	if user.Username != req.Username {
		t.Fatalf("expected username %s, got %s", req.Username, user.Username)
	}
	if user.Role != req.Role {
		t.Fatalf("expected role %s, got %s", req.Role, user.Role)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCreate_DefaultRole(t *testing.T) {
	svc, mock, mr := newUserService(t)
	defer mr.Close()

	req := CreateUserRequest{
		Username: "testuser",
		Password: "password123",
	}

	mock.ExpectExec(`INSERT INTO users`).
		WithArgs(sqlmock.AnyArg(), req.Username, sqlmock.AnyArg(), "user", defaultStatus).
		WillReturnResult(sqlmock.NewResult(1, 1))

	user, err := svc.Create(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Role != "user" {
		t.Fatalf("expected default role 'user', got %s", user.Role)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCreate_InvalidRequest(t *testing.T) {
	svc, _, mr := newUserService(t)
	defer mr.Close()

	cases := []struct {
		name string
		req  CreateUserRequest
	}{
		{"empty username", CreateUserRequest{Username: "", Password: "pass"}},
		{"empty password", CreateUserRequest{Username: "user", Password: ""}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := svc.Create(c.req)
			if err == nil {
				t.Fatalf("expected error for %s", c.name)
			}
		})
	}
}

func TestGet_Success(t *testing.T) {
	svc, mock, mr := newUserService(t)
	defer mr.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"user_id", "username", "password_hash", "role", "status", "created_at", "updated_at"}).
		AddRow("U-1", "testuser", "$2a$10$hash", "user", defaultStatus, now, now)
	mock.ExpectQuery(`SELECT \* FROM users WHERE user_id`).
		WithArgs("U-1").
		WillReturnRows(rows)

	user, err := svc.Get("U-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.UserID != "U-1" {
		t.Fatalf("expected user_id U-1, got %s", user.UserID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGet_NotFound(t *testing.T) {
	svc, mock, mr := newUserService(t)
	defer mr.Close()

	mock.ExpectQuery(`SELECT \* FROM users WHERE user_id`).
		WithArgs("U-404").
		WillReturnError(sql.ErrNoRows)

	_, err := svc.Get("U-404")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGet_EmptyUserID(t *testing.T) {
	svc, _, mr := newUserService(t)
	defer mr.Close()

	_, err := svc.Get("")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestGetByUsername_Success(t *testing.T) {
	svc, mock, mr := newUserService(t)
	defer mr.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"user_id", "username", "password_hash", "role", "status", "created_at", "updated_at"}).
		AddRow("U-1", "testuser", "$2a$10$hash", "user", defaultStatus, now, now)
	mock.ExpectQuery(`SELECT \* FROM users WHERE username`).
		WithArgs("testuser").
		WillReturnRows(rows)

	user, err := svc.GetByUsername("testuser")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Username != "testuser" {
		t.Fatalf("expected username 'testuser', got %s", user.Username)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetByUsername_NotFound(t *testing.T) {
	svc, mock, mr := newUserService(t)
	defer mr.Close()

	mock.ExpectQuery(`SELECT \* FROM users WHERE username`).
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	_, err := svc.GetByUsername("nonexistent")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetByUsername_EmptyUsername(t *testing.T) {
	svc, _, mr := newUserService(t)
	defer mr.Close()

	_, err := svc.GetByUsername("")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestVerifyPassword_Correct(t *testing.T) {
	password := "password123"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	user := &User{
		UserID:       "U-1",
		Username:     "testuser",
		PasswordHash: string(hash),
		Role:         "user",
		Status:       defaultStatus,
	}

	svc := &Service{}
	if !svc.VerifyPassword(user, password) {
		t.Fatalf("expected password to be verified")
	}
}

func TestVerifyPassword_Incorrect(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)

	user := &User{
		UserID:       "U-1",
		Username:     "testuser",
		PasswordHash: string(hash),
		Role:         "user",
		Status:       defaultStatus,
	}

	svc := &Service{}
	if svc.VerifyPassword(user, "wrongpassword") {
		t.Fatalf("expected password to be rejected")
	}
}

func TestVerifyPassword_NilUser(t *testing.T) {
	svc := &Service{}
	if svc.VerifyPassword(nil, "password") {
		t.Fatalf("expected nil user to fail verification")
	}
}

func TestVerifyPassword_EmptyHash(t *testing.T) {
	user := &User{
		UserID:       "U-1",
		Username:     "testuser",
		PasswordHash: "",
		Role:         "user",
		Status:       defaultStatus,
	}

	svc := &Service{}
	if svc.VerifyPassword(user, "password") {
		t.Fatalf("expected empty hash to fail verification")
	}
}

func TestCacheHit(t *testing.T) {
	svc, mock, mr := newUserService(t)
	defer mr.Close()

	// Pre-populate cache with user data
	user := &User{
		UserID:       "U-CACHED",
		Username:     "cacheduser",
		PasswordHash: "$2a$10$hash",
		Role:         "user",
		Status:       defaultStatus,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	ctx := context.Background()
	_ = svc.cache.setUser(ctx, "U-CACHED", user)

	// Get should return cached user without DB query
	got, err := svc.Get("U-CACHED")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Username != "cacheduser" {
		t.Fatalf("expected username 'cacheduser', got %s", got.Username)
	}

	// Verify no DB call was made
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB call: %v", err)
	}
}
