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
	statusPending   = "PENDING"
	statusRunning   = "RUNNING"
	statusSuccess   = "SUCCESS"
	statusFailed    = "FAILED"
	statusDead      = "DEAD"
	taskTypeFulfill = "FULFILL"
	taskTypeTimeout = "TIMEOUT_CANCEL"
)

const (
	sagaTypeOrder   = "ORDER_FULFILL"
	sagaStatusStart = "STARTED"
	sagaStatusDone  = "COMPLETED"
	sagaStatusComp  = "COMPENSATED"
)

const (
	stepPending         = "PENDING"
	stepRunning         = "RUNNING"
	stepDone            = "COMPLETED"
	stepFailed          = "FAILED"
	stepOrderFail       = "ORDER_FAIL"
	stepOrderCancel     = "ORDER_CANCEL"
	stepInvRelease      = "INVENTORY_RELEASE"
	stepInvReleaseOrder = "INVENTORY_RELEASE_BY_ORDER"
)

const (
	maxRetryCount = 3
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

	nextRetry := time.Now().Add(1 * time.Minute)
	return s.createWithNextRetry(req, nextRetry)
}

func (s *Service) CreateTimeoutTask(orderID, bizNo string, delay time.Duration) (*Task, error) {
	if delay <= 0 {
		delay = 15 * time.Minute
	}
	req := CreateTaskRequest{OrderID: orderID, BizNo: bizNo, Type: taskTypeTimeout}
	nextRetry := time.Now().Add(delay)
	return s.createWithNextRetry(req, nextRetry)
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

func (s *Service) GetByOrder(orderID string) (*Task, error) {
	if orderID == "" {
		return nil, ErrNotFound
	}
	return s.getByOrderAndType(orderID, taskTypeFulfill)
}

func (s *Service) MarkRunning(taskID string) error {
	_, err := s.db.Exec(`UPDATE tasks SET status=?, updated_at=NOW() WHERE task_id = ?`, statusRunning, taskID)
	return err
}

func (s *Service) MarkRunningIfStatus(taskID, status string) (bool, error) {
	res, err := s.db.Exec(`UPDATE tasks SET status=?, updated_at=NOW() WHERE task_id = ? AND status = ?`, statusRunning, taskID, status)
	if err != nil {
		return false, err
	}
	affected, _ := res.RowsAffected()
	return affected > 0, nil
}

func (s *Service) MarkRunningIfFailed(taskID string) (bool, error) {
	res, err := s.db.Exec(`UPDATE tasks SET status=?, updated_at=NOW() WHERE task_id = ? AND status = ?`, statusRunning, taskID, statusFailed)
	if err != nil {
		return false, err
	}
	affected, _ := res.RowsAffected()
	return affected > 0, nil
}

func (s *Service) MarkSuccess(taskID string) error {
	_, err := s.db.Exec(`UPDATE tasks SET status=?, updated_at=NOW() WHERE task_id = ?`, statusSuccess, taskID)
	return err
}

func (s *Service) MarkFailed(taskID string, retryCount int, nextRetryAt time.Time) error {
	_, err := s.db.Exec(`UPDATE tasks SET status=?, retry_count=?, next_retry_at=?, updated_at=NOW() WHERE task_id = ?`, statusFailed, retryCount, nextRetryAt, taskID)
	return err
}

func (s *Service) MarkDead(taskID string, retryCount int) error {
	_, err := s.db.Exec(`UPDATE tasks SET status=?, retry_count=?, updated_at=NOW() WHERE task_id = ?`, statusDead, retryCount, taskID)
	return err
}

func (s *Service) ListRetryTasks(limit int) ([]Task, error) {
	if limit <= 0 {
		limit = 20
	}
	var tasks []Task
	if err := s.db.Select(&tasks, `SELECT * FROM tasks WHERE status = ? AND retry_count < ? AND next_retry_at <= NOW() ORDER BY next_retry_at ASC LIMIT ?`, statusFailed, maxRetryCount, limit); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s *Service) ListTimeoutTasks(limit int) ([]Task, error) {
	if limit <= 0 {
		limit = 20
	}
	var tasks []Task
	if err := s.db.Select(&tasks, `SELECT * FROM tasks WHERE status = ? AND type = ? AND next_retry_at <= NOW() ORDER BY next_retry_at ASC LIMIT ?`, statusPending, taskTypeTimeout, limit); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s *Service) CreateSaga(sagaID, bizNo, typ string) error {
	if sagaID == "" || bizNo == "" {
		return errors.New("invalid saga")
	}
	if typ == "" {
		typ = sagaTypeOrder
	}
	_, err := s.db.Exec(`INSERT INTO sagas(saga_id,biz_no,type,status,reason,created_at,updated_at) VALUES(?,?,?,?,?,NOW(),NOW())`,
		sagaID, bizNo, typ, sagaStatusStart, "")
	if err != nil {
		if isDuplicate(err) {
			return nil
		}
		return err
	}
	return nil
}

func (s *Service) MarkSagaCompleted(sagaID string) error {
	if sagaID == "" {
		return nil
	}
	_, err := s.db.Exec(`UPDATE sagas SET status=?, updated_at=NOW() WHERE saga_id = ? AND status = ?`, sagaStatusDone, sagaID, sagaStatusStart)
	return err
}

func (s *Service) MarkSagaCompensated(sagaID, reason string) error {
	if sagaID == "" {
		return nil
	}
	_, err := s.db.Exec(`UPDATE sagas SET status=?, reason=?, updated_at=NOW() WHERE saga_id = ? AND status <> ?`, sagaStatusComp, reason, sagaID, sagaStatusDone)
	return err
}

func (s *Service) CreateSagaStep(sagaID, step, nextStep, reason, payload string) error {
	if sagaID == "" || step == "" {
		return errors.New("invalid saga step")
	}
	if payload == "" {
		payload = "{}"
	}
	_, err := s.db.Exec(`INSERT INTO saga_steps(saga_id,step,status,reason,payload,next_step,created_at,updated_at) VALUES(?,?,?,?,?,?,NOW(),NOW())`,
		sagaID, step, stepPending, reason, payload, nextStep)
	if err != nil {
		if isDuplicate(err) {
			return nil
		}
		return err
	}
	return nil
}

func (s *Service) ListPendingSagaSteps(limit int) ([]SagaStep, error) {
	if limit <= 0 {
		limit = 20
	}
	var steps []SagaStep
	if err := s.db.Select(&steps, `SELECT * FROM saga_steps WHERE status = ? ORDER BY created_at ASC LIMIT ?`, stepPending, limit); err != nil {
		return nil, err
	}
	return steps, nil
}

func (s *Service) MarkSagaStepRunning(sagaID, step string) (bool, error) {
	res, err := s.db.Exec(`UPDATE saga_steps SET status=?, updated_at=NOW() WHERE saga_id = ? AND step = ? AND status = ?`,
		stepRunning, sagaID, step, stepPending)
	if err != nil {
		return false, err
	}
	affected, _ := res.RowsAffected()
	return affected > 0, nil
}

func (s *Service) MarkSagaStepCompleted(sagaID, step string) error {
	_, err := s.db.Exec(`UPDATE saga_steps SET status=?, updated_at=NOW() WHERE saga_id = ? AND step = ?`, stepDone, sagaID, step)
	return err
}

func (s *Service) MarkSagaStepFailed(sagaID, step, reason string) error {
	_, err := s.db.Exec(`UPDATE saga_steps SET status=?, reason=?, updated_at=NOW() WHERE saga_id = ? AND step = ?`, stepFailed, reason, sagaID, step)
	return err
}

func (s *Service) DelayTask(taskID string, retryCount int, nextRetryAt time.Time) error {
	_, err := s.db.Exec(`UPDATE tasks SET retry_count = ?, next_retry_at = ?, updated_at = NOW() WHERE task_id = ?`, retryCount, nextRetryAt, taskID)
	return err
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

func (s *Service) createWithNextRetry(req CreateTaskRequest, nextRetry time.Time) (*Task, error) {
	taskID := uuid.NewString()
	_, err := s.db.Exec(`INSERT INTO tasks(task_id,biz_no,order_id,type,status,retry_count,next_retry_at,created_at,updated_at) VALUES(?,?,?,?,?,?,?,NOW(),NOW())`,
		taskID, req.BizNo, req.OrderID, req.Type, statusPending, 0, nextRetry)
	if err != nil {
		if isDuplicate(err) {
			return s.getByOrderAndType(req.OrderID, req.Type)
		}
		return nil, err
	}
	return &Task{
		TaskID:      taskID,
		BizNo:       req.BizNo,
		OrderID:     req.OrderID,
		Type:        req.Type,
		Status:      statusPending,
		RetryCount:  0,
		NextRetryAt: nextRetry,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

func isDuplicate(err error) bool {
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		return me.Number == 1062
	}
	return false
}
