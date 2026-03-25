package user

import "time"

type CreateUserRequest struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Mobile string `json:"mobile"`
}

type User struct {
	ID        int64     `db:"id" json:"-"`
	UserID    string    `db:"user_id" json:"user_id"`
	Name      string    `db:"name" json:"name"`
	Mobile    string    `db:"mobile" json:"mobile"`
	Status    int       `db:"status" json:"status"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
