package task

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func newTaskService(t *testing.T) (*Service, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	sqlxDB := sqlx.NewDb(db, "mysql")
	return NewService(sqlxDB), mock
}

func TestCreateTask_Success(t *testing.T) {
	svc, mock := newTaskService(t)

	req := CreateTaskRequest{
		BizNo:   "BIZ-1",
		OrderID: "O-1",
		Type:    taskTypeFulfill,
	}
	mock.ExpectExec(`INSERT INTO tasks`).
		WithArgs(sqlmock.AnyArg(), req.BizNo, req.OrderID, req.Type, statusPending, 0, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	task, err := svc.Create(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.OrderID != req.OrderID {
		t.Fatalf("expected order_id %s, got %s", req.OrderID, task.OrderID)
	}
	if task.BizNo != req.BizNo {
		t.Fatalf("expected biz_no %s, got %s", req.BizNo, task.BizNo)
	}
	if task.Type != taskTypeFulfill {
		t.Fatalf("expected type %s, got %s", taskTypeFulfill, task.Type)
	}
	if task.Status != statusPending {
		t.Fatalf("expected status %s, got %s", statusPending, task.Status)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCreateTask_InvalidRequest(t *testing.T) {
	svc, _ := newTaskService(t)

	cases := []struct {
		name string
		req  CreateTaskRequest
	}{
		{"empty order_id", CreateTaskRequest{BizNo: "BIZ-1", OrderID: "", Type: taskTypeFulfill}},
		{"empty biz_no", CreateTaskRequest{BizNo: "", OrderID: "O-1", Type: taskTypeFulfill}},
		{"empty type", CreateTaskRequest{BizNo: "BIZ-1", OrderID: "O-1", Type: ""}},
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

func TestCreateTimeoutTask_DefaultDelay(t *testing.T) {
	svc, mock := newTaskService(t)

	now := time.Now()
	mock.ExpectExec(`INSERT INTO tasks`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), taskTypeTimeout, statusPending, 0, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	task, err := svc.CreateTimeoutTask("O-1", "BIZ-1", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Type != taskTypeTimeout {
		t.Fatalf("expected type %s, got %s", taskTypeTimeout, task.Type)
	}
	if task.NextRetryAt.Before(now.Add(14 * time.Minute)) {
		t.Fatalf("expected next_retry_at at least 15min from now, got %v", task.NextRetryAt)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCreateTimeoutTask_CustomDelay(t *testing.T) {
	svc, mock := newTaskService(t)

	delay := 10 * time.Minute
	mock.ExpectExec(`INSERT INTO tasks`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), taskTypeTimeout, statusPending, 0, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	task, err := svc.CreateTimeoutTask("O-1", "BIZ-1", delay)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.OrderID != "O-1" {
		t.Fatalf("expected order_id O-1, got %s", task.OrderID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGet_NotFound(t *testing.T) {
	svc, mock := newTaskService(t)

	mock.ExpectQuery(`SELECT \* FROM tasks WHERE task_id`).
		WithArgs("T-404").
		WillReturnError(sql.ErrNoRows)

	_, err := svc.Get("T-404")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGet_Success(t *testing.T) {
	svc, mock := newTaskService(t)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"task_id", "biz_no", "order_id", "type", "status", "retry_count", "next_retry_at", "created_at", "updated_at"}).
		AddRow("T-1", "BIZ-1", "O-1", taskTypeFulfill, statusPending, 0, now, now, now)
	mock.ExpectQuery(`SELECT \* FROM tasks WHERE task_id`).
		WithArgs("T-1").
		WillReturnRows(rows)

	task, err := svc.Get("T-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.TaskID != "T-1" {
		t.Fatalf("expected task_id T-1, got %s", task.TaskID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetByOrder_Success(t *testing.T) {
	svc, mock := newTaskService(t)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"task_id", "biz_no", "order_id", "type", "status", "retry_count", "next_retry_at", "created_at", "updated_at"}).
		AddRow("T-1", "BIZ-1", "O-1", taskTypeFulfill, statusSuccess, 0, now, now, now)
	mock.ExpectQuery(`SELECT \* FROM tasks WHERE order_id = \? AND type = \?`).
		WithArgs("O-1", taskTypeFulfill).
		WillReturnRows(rows)

	task, err := svc.GetByOrder("O-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.OrderID != "O-1" {
		t.Fatalf("expected order_id O-1, got %s", task.OrderID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetByOrder_NotFound(t *testing.T) {
	svc, mock := newTaskService(t)

	mock.ExpectQuery(`SELECT \* FROM tasks WHERE order_id = \? AND type = \?`).
		WithArgs("O-404", taskTypeFulfill).
		WillReturnError(sql.ErrNoRows)

	_, err := svc.GetByOrder("O-404")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetByOrder_EmptyOrderID(t *testing.T) {
	svc, _ := newTaskService(t)

	_, err := svc.GetByOrder("")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestRetry_Success(t *testing.T) {
	svc, mock := newTaskService(t)

	mock.ExpectExec(`UPDATE tasks SET retry_count`).
		WithArgs("T-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	now := time.Now()
	rows := sqlmock.NewRows([]string{"task_id", "biz_no", "order_id", "type", "status", "retry_count", "next_retry_at", "created_at", "updated_at"}).
		AddRow("T-1", "BIZ-1", "O-1", taskTypeFulfill, statusPending, 1, now.Add(time.Minute), now, now)
	mock.ExpectQuery(`SELECT \* FROM tasks WHERE task_id`).
		WithArgs("T-1").
		WillReturnRows(rows)

	_, err := svc.Retry("T-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRetry_EmptyTaskID(t *testing.T) {
	svc, _ := newTaskService(t)

	_, err := svc.Retry("")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestMarkRunning_Success(t *testing.T) {
	svc, mock := newTaskService(t)

	mock.ExpectExec(`UPDATE tasks SET status=\?, updated_at=NOW\(\) WHERE task_id = \?`).
		WithArgs(statusRunning, "T-1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := svc.MarkRunning("T-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMarkRunningIfStatus_Success(t *testing.T) {
	svc, mock := newTaskService(t)

	mock.ExpectExec(`UPDATE tasks SET status=\?, updated_at=NOW\(\) WHERE task_id = \? AND status = \?`).
		WithArgs(statusRunning, "T-1", statusPending).
		WillReturnResult(sqlmock.NewResult(1, 1))

	ok, err := svc.MarkRunningIfStatus("T-1", statusPending)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMarkRunningIfStatus_AlreadyProcessed(t *testing.T) {
	svc, mock := newTaskService(t)

	mock.ExpectExec(`UPDATE tasks SET status=\?, updated_at=NOW\(\) WHERE task_id = \? AND status = \?`).
		WithArgs(statusRunning, "T-1", statusPending).
		WillReturnResult(sqlmock.NewResult(0, 0))

	ok, err := svc.MarkRunningIfStatus("T-1", statusPending)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatalf("expected ok=false")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMarkRunningIfFailed_Success(t *testing.T) {
	svc, mock := newTaskService(t)

	mock.ExpectExec(`UPDATE tasks SET status=\?, updated_at=NOW\(\) WHERE task_id = \? AND status = \?`).
		WithArgs(statusRunning, "T-1", statusFailed).
		WillReturnResult(sqlmock.NewResult(1, 1))

	ok, err := svc.MarkRunningIfFailed("T-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMarkSuccess_Success(t *testing.T) {
	svc, mock := newTaskService(t)

	mock.ExpectExec(`UPDATE tasks SET status=\?, updated_at=NOW\(\) WHERE task_id = \?`).
		WithArgs(statusSuccess, "T-1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := svc.MarkSuccess("T-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMarkFailed_Success(t *testing.T) {
	svc, mock := newTaskService(t)

	nextRetry := time.Now().Add(time.Minute)
	mock.ExpectExec(`UPDATE tasks SET status=\?, retry_count=\?, next_retry_at=\?, updated_at=NOW\(\) WHERE task_id = \?`).
		WithArgs(statusFailed, 2, sqlmock.AnyArg(), "T-1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := svc.MarkFailed("T-1", 2, nextRetry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMarkDead_Success(t *testing.T) {
	svc, mock := newTaskService(t)

	mock.ExpectExec(`UPDATE tasks SET status=\?, retry_count=\?, updated_at=NOW\(\) WHERE task_id = \?`).
		WithArgs(statusDead, 3, "T-1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := svc.MarkDead("T-1", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestListRetryTasks_Success(t *testing.T) {
	svc, mock := newTaskService(t)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"task_id", "biz_no", "order_id", "type", "status", "retry_count", "next_retry_at", "created_at", "updated_at"}).
		AddRow("T-1", "BIZ-1", "O-1", taskTypeFulfill, statusFailed, 1, now, now, now).
		AddRow("T-2", "BIZ-2", "O-2", taskTypeFulfill, statusFailed, 2, now, now, now)
	mock.ExpectQuery(`SELECT \* FROM tasks WHERE status = \? AND retry_count < \? AND next_retry_at <= NOW\(\) ORDER BY next_retry_at ASC LIMIT \?`).
		WithArgs(statusFailed, maxRetryCount, 20).
		WillReturnRows(rows)

	tasks, err := svc.ListRetryTasks(20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestListRetryTasks_DefaultLimit(t *testing.T) {
	svc, mock := newTaskService(t)

	mock.ExpectQuery(`SELECT \* FROM tasks WHERE status = \? AND retry_count < \? AND next_retry_at <= NOW\(\) ORDER BY next_retry_at ASC LIMIT \?`).
		WithArgs(statusFailed, maxRetryCount, 20).
		WillReturnRows(sqlmock.NewRows([]string{"task_id", "biz_no", "order_id", "type", "status", "retry_count", "next_retry_at", "created_at", "updated_at"}))

	_, err := svc.ListRetryTasks(0) // Should use default 20
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestListTimeoutTasks_Success(t *testing.T) {
	svc, mock := newTaskService(t)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"task_id", "biz_no", "order_id", "type", "status", "retry_count", "next_retry_at", "created_at", "updated_at"}).
		AddRow("T-1", "BIZ-1", "O-1", taskTypeTimeout, statusPending, 0, now.Add(-time.Minute), now, now)
	mock.ExpectQuery(`SELECT \* FROM tasks WHERE status = \? AND type = \? AND next_retry_at <= NOW\(\) ORDER BY next_retry_at ASC LIMIT \?`).
		WithArgs(statusPending, taskTypeTimeout, 20).
		WillReturnRows(rows)

	tasks, err := svc.ListTimeoutTasks(20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Type != taskTypeTimeout {
		t.Fatalf("expected type %s, got %s", taskTypeTimeout, tasks[0].Type)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCreateSaga_Success(t *testing.T) {
	svc, mock := newTaskService(t)

	mock.ExpectExec(`INSERT INTO sagas`).
		WithArgs("SAGA-1", "BIZ-1", sagaTypeOrder, sagaStatusStart, "").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := svc.CreateSaga("SAGA-1", "BIZ-1", sagaTypeOrder)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCreateSaga_Idempotent(t *testing.T) {
	svc, mock := newTaskService(t)

	mock.ExpectExec(`INSERT INTO sagas`).
		WithArgs("SAGA-1", "BIZ-1", sagaTypeOrder, sagaStatusStart, "").
		WillReturnError(&mysql.MySQLError{Number: 1062})

	err := svc.CreateSaga("SAGA-1", "BIZ-1", sagaTypeOrder)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCreateSaga_Invalid(t *testing.T) {
	svc, _ := newTaskService(t)

	cases := []struct {
		name   string
		sagaID string
		bizNo  string
	}{
		{"empty saga_id", "", "BIZ-1"},
		{"empty biz_no", "SAGA-1", ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := svc.CreateSaga(c.sagaID, c.bizNo, sagaTypeOrder)
			if err == nil {
				t.Fatalf("expected error for %s", c.name)
			}
		})
	}
}

func TestCreateSaga_DefaultType(t *testing.T) {
	svc, mock := newTaskService(t)

	mock.ExpectExec(`INSERT INTO sagas`).
		WithArgs("SAGA-1", "BIZ-1", sagaTypeOrder, sagaStatusStart, "").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := svc.CreateSaga("SAGA-1", "BIZ-1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMarkSagaCompleted_Success(t *testing.T) {
	svc, mock := newTaskService(t)

	mock.ExpectExec(`UPDATE sagas SET status=\?, updated_at=NOW\(\) WHERE saga_id = \? AND status = \?`).
		WithArgs(sagaStatusDone, "SAGA-1", sagaStatusStart).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := svc.MarkSagaCompleted("SAGA-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMarkSagaCompleted_EmptyID(t *testing.T) {
	svc, _ := newTaskService(t)

	err := svc.MarkSagaCompleted("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMarkSagaCompensated_Success(t *testing.T) {
	svc, mock := newTaskService(t)

	mock.ExpectExec(`UPDATE sagas SET status=\?, reason=\?, updated_at=NOW\(\) WHERE saga_id = \? AND status <> \?`).
		WithArgs(sagaStatusComp, "timeout", "SAGA-1", sagaStatusDone).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := svc.MarkSagaCompensated("SAGA-1", "timeout")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMarkSagaCompensated_EmptyID(t *testing.T) {
	svc, _ := newTaskService(t)

	err := svc.MarkSagaCompensated("", "reason")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateSagaStep_Success(t *testing.T) {
	svc, mock := newTaskService(t)

	mock.ExpectExec(`INSERT INTO saga_steps`).
		WithArgs("SAGA-1", stepOrderCancel, stepPending, "cancel reason", "{}", stepInvReleaseOrder).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := svc.CreateSagaStep("SAGA-1", stepOrderCancel, stepInvReleaseOrder, "cancel reason", "{}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCreateSagaStep_DefaultPayload(t *testing.T) {
	svc, mock := newTaskService(t)

	mock.ExpectExec(`INSERT INTO saga_steps`).
		WithArgs("SAGA-1", stepOrderCancel, stepPending, "", "{}", stepInvReleaseOrder).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := svc.CreateSagaStep("SAGA-1", stepOrderCancel, stepInvReleaseOrder, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCreateSagaStep_Invalid(t *testing.T) {
	svc, _ := newTaskService(t)

	cases := []struct {
		name string
		sid  string
		step string
	}{
		{"empty saga_id", "", "step1"},
		{"empty step", "SAGA-1", ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := svc.CreateSagaStep(c.sid, c.step, "next", "", "{}")
			if err == nil {
				t.Fatalf("expected error for %s", c.name)
			}
		})
	}
}

func TestListPendingSagaSteps_Success(t *testing.T) {
	svc, mock := newTaskService(t)

	rows := sqlmock.NewRows([]string{"saga_id", "step", "status", "reason", "payload", "next_step", "created_at", "updated_at"}).
		AddRow("SAGA-1", stepOrderCancel, stepPending, "", "{}", stepInvReleaseOrder, time.Now(), time.Now())
	mock.ExpectQuery(`SELECT \* FROM saga_steps WHERE status = \? ORDER BY created_at ASC LIMIT \?`).
		WithArgs(stepPending, 20).
		WillReturnRows(rows)

	steps, err := svc.ListPendingSagaSteps(20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(steps))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMarkSagaStepRunning_Success(t *testing.T) {
	svc, mock := newTaskService(t)

	mock.ExpectExec(`UPDATE saga_steps SET status=\?, updated_at=NOW\(\) WHERE saga_id = \? AND step = \? AND status = \?`).
		WithArgs(stepRunning, "SAGA-1", stepOrderCancel, stepPending).
		WillReturnResult(sqlmock.NewResult(1, 1))

	ok, err := svc.MarkSagaStepRunning("SAGA-1", stepOrderCancel)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMarkSagaStepCompleted_Success(t *testing.T) {
	svc, mock := newTaskService(t)

	mock.ExpectExec(`UPDATE saga_steps SET status=\?, updated_at=NOW\(\) WHERE saga_id = \? AND step = \?`).
		WithArgs(stepDone, "SAGA-1", stepOrderCancel).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := svc.MarkSagaStepCompleted("SAGA-1", stepOrderCancel)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMarkSagaStepFailed_Success(t *testing.T) {
	svc, mock := newTaskService(t)

	mock.ExpectExec(`UPDATE saga_steps SET status=\?, reason=\?, updated_at=NOW\(\) WHERE saga_id = \? AND step = \?`).
		WithArgs(stepFailed, "order not found", "SAGA-1", stepOrderCancel).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := svc.MarkSagaStepFailed("SAGA-1", stepOrderCancel, "order not found")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDelayTask_Success(t *testing.T) {
	svc, mock := newTaskService(t)

	nextRetry := time.Now().Add(2 * time.Minute)
	mock.ExpectExec(`UPDATE tasks SET retry_count = \?, next_retry_at = \?, updated_at = NOW\(\) WHERE task_id = \?`).
		WithArgs(2, sqlmock.AnyArg(), "T-1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := svc.DelayTask("T-1", 2, nextRetry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
