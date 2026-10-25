package user

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"

	"ecomm-micro/proto/userpb"
)

func newUserRouter(svc *Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	NewHandler(svc).Register(r)
	return r
}

func doJSON(r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(method, path, nil)
	} else {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestHandler_CreateUser(t *testing.T) {
	svc, mock, _ := newUserService(t)
	r := newUserRouter(svc)

	mock.ExpectExec(`INSERT INTO users`).
		WithArgs(sqlmock.AnyArg(), "alice", sqlmock.AnyArg(), "user", defaultStatus).
		WillReturnResult(sqlmock.NewResult(1, 1))

	w := doJSON(r, http.MethodPost, "/users", `{"username":"alice","password":"secret123","role":"user"}`)
	if !strings.Contains(w.Body.String(), `"code":0`) || !strings.Contains(w.Body.String(), `"username":"alice"`) {
		t.Fatalf("body = %s", w.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestHandler_CreateUser_MissingFields(t *testing.T) {
	svc, _, _ := newUserService(t)
	r := newUserRouter(svc)

	w := doJSON(r, http.MethodPost, "/users", `{"username":"alice"}`)
	if !strings.Contains(w.Body.String(), `"code":40001`) {
		t.Fatalf("body = %s", w.Body.String())
	}
}

func TestHandler_GetUser_NotFound(t *testing.T) {
	svc, mock, _ := newUserService(t)
	r := newUserRouter(svc)

	mock.ExpectQuery(`SELECT \* FROM users WHERE user_id`).WithArgs("U-404").
		WillReturnError(sql.ErrNoRows)

	w := doJSON(r, http.MethodGet, "/users/U-404", "")
	if !strings.Contains(w.Body.String(), `"code":40401`) {
		t.Fatalf("body = %s", w.Body.String())
	}
}

func userRows(u User) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"user_id", "username", "password_hash", "role", "status", "created_at", "updated_at"}).
		AddRow(u.UserID, u.Username, u.PasswordHash, u.Role, u.Status, u.CreatedAt, u.UpdatedAt)
}

func TestGRPCServer_GetUserByName(t *testing.T) {
	svc, mock, _ := newUserService(t)
	srv := NewGRPCServer(svc)

	want := User{UserID: "U-1", Username: "alice", PasswordHash: "$2a$10$hash", Role: "user", Status: 1,
		CreatedAt: time.Now(), UpdatedAt: time.Now()}
	mock.ExpectQuery(`SELECT \* FROM users WHERE username`).WithArgs("alice").WillReturnRows(userRows(want))

	resp, err := srv.GetUserByName(t.Context(), &userpb.GetUserByNameRequest{Username: "alice"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.UserId != "U-1" || resp.Username != "alice" || resp.Phone != "" || resp.PasswordHash != want.PasswordHash {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestGRPCServer_GetUserByPhone(t *testing.T) {
	svc, mock, _ := newUserService(t)
	srv := NewGRPCServer(svc)

	rows := sqlmock.NewRows([]string{"user_id", "username", "password_hash", "phone", "role", "status", "created_at", "updated_at"}).
		AddRow("U-9", "bob", "$2a$10$hash", "17600000000", "user", 1, time.Now(), time.Now())
	mock.ExpectQuery(`SELECT \* FROM users WHERE phone`).WithArgs("17600000000").WillReturnRows(rows)

	resp, err := srv.GetUserByPhone(t.Context(), &userpb.GetUserByPhoneRequest{Phone: "17600000000"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.UserId != "U-9" || resp.Phone != "17600000000" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestGRPCServer_GetUser_NotFound(t *testing.T) {
	svc, mock, _ := newUserService(t)
	srv := NewGRPCServer(svc)

	mock.ExpectQuery(`SELECT \* FROM users WHERE user_id`).WithArgs("U-404").
		WillReturnError(sql.ErrNoRows)

	if _, err := srv.GetUser(t.Context(), &userpb.GetUserRequest{UserId: "U-404"}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}
