package mysqlit

import (
	"testing"

	"go-micro/internal/user"
)

func TestUserService_CreateAndGet(t *testing.T) {
	db, teardown := NewDB(t)
	defer teardown()

	svc := user.NewService(db.DB, db.RDB)

	t.Run("CreateUser", func(t *testing.T) {
		req := user.CreateUserRequest{
			Username: "testusr" + randomID(),
			Password: "password123",
			Role:     "user",
		}
		u, err := svc.Create(req)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
		if u.Username != req.Username {
			t.Errorf("Expected username %s, got %s", req.Username, u.Username)
		}
		if u.Role != req.Role {
			t.Errorf("Expected role %s, got %s", req.Role, u.Role)
		}
		if u.UserID == "" {
			t.Error("Expected non-empty userID")
		}
	})

	t.Run("GetUserByID", func(t *testing.T) {
		req := user.CreateUserRequest{
			Username: "getbyid" + randomID(),
			Password: "password123",
			Role:     "admin",
		}
		created, err := svc.Create(req)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		u, err := svc.Get(created.UserID)
		if err != nil {
			t.Fatalf("Failed to get user: %v", err)
		}
		if u.Username != req.Username {
			t.Errorf("Expected username %s, got %s", req.Username, u.Username)
		}
	})

	t.Run("GetUserByUsername", func(t *testing.T) {
		req := user.CreateUserRequest{
			Username: "getbynam" + randomID(),
			Password: "password123",
			Role:     "user",
		}
		_, err := svc.Create(req)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		u, err := svc.GetByUsername(req.Username)
		if err != nil {
			t.Fatalf("Failed to get user by username: %v", err)
		}
		if u.UserID == "" {
			t.Error("Expected non-empty userID")
		}
	})

	t.Run("GetNonExistentUser", func(t *testing.T) {
		_, err := svc.Get("nonexistent-user-id")
		if err != user.ErrNotFound {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})

	t.Run("CreateUserWithDefaultRole", func(t *testing.T) {
		req := user.CreateUserRequest{
			Username: "defrole" + randomID(),
			Password: "password123",
		}
		u, err := svc.Create(req)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
		if u.Role != "user" {
			t.Errorf("Expected default role 'user', got %s", u.Role)
		}
	})

	t.Run("CreateUserWithEmptyUsername", func(t *testing.T) {
		req := user.CreateUserRequest{
			Username: "",
			Password: "password123",
		}
		_, err := svc.Create(req)
		if err == nil {
			t.Error("Expected error for empty username")
		}
	})

	t.Run("CreateUserWithEmptyPassword", func(t *testing.T) {
		req := user.CreateUserRequest{
			Username: "emptypwd" + randomID(),
			Password: "",
		}
		_, err := svc.Create(req)
		if err == nil {
			t.Error("Expected error for empty password")
		}
	})
}

func TestUserService_VerifyPassword(t *testing.T) {
	db, teardown := NewDB(t)
	defer teardown()

	svc := user.NewService(db.DB, db.RDB)

	username := "verify" + randomID()
	password := "correct-password"
	req := user.CreateUserRequest{
		Username: username,
		Password: password,
		Role:     "user",
	}
	_, err := svc.Create(req)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Query database directly to get user with password hash
	type userWithHash struct {
		user.User
		PasswordHash string `db:"password_hash"`
	}
	var uWithHash userWithHash
	err = db.DB.Get(&uWithHash, "SELECT * FROM users WHERE username = ?", username)
	if err != nil {
		t.Fatalf("Failed to query user: %v", err)
	}
	u := &uWithHash.User
	u.PasswordHash = uWithHash.PasswordHash

	t.Run("CorrectPassword", func(t *testing.T) {
		if !svc.VerifyPassword(u, password) {
			t.Error("Expected password verification to pass")
		}
	})

	t.Run("IncorrectPassword", func(t *testing.T) {
		if svc.VerifyPassword(u, "wrong-password") {
			t.Error("Expected password verification to fail")
		}
	})

	t.Run("NilUser", func(t *testing.T) {
		if svc.VerifyPassword(nil, "any-password") {
			t.Error("Expected password verification to fail for nil user")
		}
	})
}
