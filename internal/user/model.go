package user

import (
	"database/sql"
	"time"
)

type CreateUserRequest struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type User struct {
	ID                int64          `db:"id" json:"-"`
	UserID            string         `db:"user_no" json:"user_id"`
	Username          string         `db:"username" json:"username"`
	DisplayName       sql.NullString `db:"display_name" json:"display_name,omitempty"`
	PasswordHash      string         `db:"password_hash" json:"-"`
	Phone             sql.NullString `db:"phone" json:"phone,omitempty"`
	Email             sql.NullString `db:"email" json:"email,omitempty"`
	AvatarURL         sql.NullString `db:"avatar_url" json:"avatar_url,omitempty"`
	Status            int            `db:"status" json:"status"`
	RegisterSource    string         `db:"register_source" json:"register_source"`
	LastLoginAt       sql.NullTime   `db:"last_login_at" json:"last_login_at,omitempty"`
	LastLoginIP       sql.NullString `db:"last_login_ip" json:"last_login_ip,omitempty"`
	PasswordUpdatedAt sql.NullTime   `db:"password_updated_at" json:"password_updated_at,omitempty"`
	DeletedAt         sql.NullTime   `db:"deleted_at" json:"deleted_at,omitempty"`
	Role              string         `db:"role" json:"role"`
	CreatedAt         time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time      `db:"updated_at" json:"updated_at"`
}

type CachedUser struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Status   int    `json:"status"`
}
