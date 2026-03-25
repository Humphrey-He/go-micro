package task

import (
	"database/sql"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var ErrNotFound = errors.New("task not found")

const (
	statusPending = "PENDING"
)

type Service struct {
	db *sqlx.DB
}

func NewService(dbx *sqlx.DB) *Service {
	return &Service{db: dbx}
}

func (s *Service) Create(req CreateTaskRequest) (*Task, error) {
	if req.OrderID == "" || req.BizNo == "" || req.Type == "" {
		return nil, errors.New("invalid request")
	}

	taskID := uuid.NewString()
	nextRetry := time.Now().Add(1 * time.Minute)
	_, err := s.db.Exec(`INSERT INTO tasks(task_id,biz_no,order_id,type,status,retry_count,next_retry_at,created_at,updated_at) VALUES(?,?,?,?,?,?,?,NOW(),NOW())`,
		taskID, req.BizNo, req.OrderID, req.Type, statusPending, 0, nextRetry)
	if err != nil {
		if isDuplicate(err) {
			return s.getByOrderAndType(req.OrderID, req.Type)
		}
		return nil, err
	}
	return &Task{TaskID: taskID, BizNo: req.BizNo, OrderID: req.OrderID, Type: req.Type, Status: statusPending, RetryCount: 0, NextRetryAt: nextRetry, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
}

func (s *Service) Get(taskID string) (*Task, error) {
	if taskID == "" {
		return nil, ErrNotFound
	}
	t := Task{}
	if err := s.db.Get(&t, `SELECT * FROM tasks WHERE task_id = ?`, taskID); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (s *Service) Retry(taskID string) (*Task, error) {
	if taskID == "" {
		return nil, ErrNotFound
	}

	_, err := s.db.Exec(`UPDATE tasks SET retry_count = retry_count + 1, next_retry_at = DATE_ADD(NOW(), INTERVAL 1 MINUTE), updated_at = NOW() WHERE task_id = ?`, taskID)
	if err != nil {
		return nil, err
	}
	return s.Get(taskID)
}

func (s *Service) getByOrderAndType(orderID, typ string) (*Task, error) {
	t := Task{}
	if err := s.db.Get(&t, `SELECT * FROM tasks WHERE order_id = ? AND type = ?`, orderID, typ); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}

func isDuplicate(err error) bool {
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		return me.Number == 1062
	}
	return false
}
