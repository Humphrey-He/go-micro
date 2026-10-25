package task

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"

	"go-micro/proto/taskpb"
)

// ---- fakes ----

type fakeOrderUpdater struct {
	updateErr error
	calls     []string // 记录 "orderID:from->to" 调用序列
}

func (f *fakeOrderUpdater) UpdateStatus(_ context.Context, orderID, from, to string) error {
	f.calls = append(f.calls, orderID+":"+from+"->"+to)
	return f.updateErr
}

type fakeCancelUpdater struct {
	fakeOrderUpdater
	cancelErr error
	canceled  []string
}

func (f *fakeCancelUpdater) Cancel(_ context.Context, orderID string) error {
	f.canceled = append(f.canceled, orderID)
	return f.cancelErr
}

type fakeCompensationInventory struct {
	releaseErr        error
	releaseByOrderErr error
	released          []string
	releasedByOrder   []string
}

func (f *fakeCompensationInventory) Release(_ context.Context, reservedID string) error {
	f.released = append(f.released, reservedID)
	return f.releaseErr
}

func (f *fakeCompensationInventory) ReleaseByOrder(_ context.Context, orderID string) error {
	f.releasedByOrder = append(f.releasedByOrder, orderID)
	return f.releaseByOrderErr
}

// ---- executor ----

func TestProcessFulfillment_Success(t *testing.T) {
	svc, mock := newTaskService(t)
	ord := &fakeOrderUpdater{}

	mock.ExpectExec(`UPDATE tasks SET status=\?`).WithArgs(statusRunning, "T-1").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE tasks SET status=\?`).WithArgs(statusSuccess, "T-1").WillReturnResult(sqlmock.NewResult(0, 1))

	err := processFulfillment(context.Background(), svc, ord, &Task{TaskID: "T-1", OrderID: "O-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"O-1:RESERVED->PROCESSING", "O-1:PROCESSING->SUCCESS"}
	if len(ord.calls) != 2 || ord.calls[0] != want[0] || ord.calls[1] != want[1] {
		t.Fatalf("order status calls = %v, want %v", ord.calls, want)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestProcessFulfillment_NilOrderUpdater(t *testing.T) {
	svc, mock := newTaskService(t)

	mock.ExpectExec(`UPDATE tasks SET status=\?`).WithArgs(statusRunning, "T-1").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE tasks SET status=\?`).WithArgs(statusSuccess, "T-1").WillReturnResult(sqlmock.NewResult(0, 1))

	if err := processFulfillment(context.Background(), svc, nil, &Task{TaskID: "T-1", OrderID: "O-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProcessFulfillment_MarkRunningFails(t *testing.T) {
	svc, mock := newTaskService(t)
	ord := &fakeOrderUpdater{}

	mock.ExpectExec(`UPDATE tasks SET status=\?`).WithArgs(statusRunning, "T-1").
		WillReturnError(errors.New("db down"))

	if err := processFulfillment(context.Background(), svc, ord, &Task{TaskID: "T-1"}); err == nil {
		t.Fatal("expected error when MarkRunning fails")
	}
	if len(ord.calls) != 0 {
		t.Fatalf("order should not be touched, got calls %v", ord.calls)
	}
}

func TestHandleTaskFailure_ExhaustedRetries(t *testing.T) {
	svc, mock := newTaskService(t)
	ord := &fakeOrderUpdater{}

	// RetryCount=2 → retryCount+1=3 达到 maxRetryCount，任务应被判死
	mock.ExpectExec(`UPDATE tasks SET status=\?`).WithArgs(statusDead, 3, "T-1").WillReturnResult(sqlmock.NewResult(0, 1))

	handleTaskFailure(context.Background(), svc, ord, &Task{TaskID: "T-1", OrderID: "O-1", RetryCount: 2})

	if len(ord.calls) != 1 || ord.calls[0] != "O-1:PROCESSING->FAILED" {
		t.Fatalf("order calls = %v, want [O-1:PROCESSING->FAILED]", ord.calls)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestHandleTaskFailure_ScheduleRetry(t *testing.T) {
	svc, mock := newTaskService(t)
	ord := &fakeOrderUpdater{}

	// RetryCount=1 → 第 2 次重试，按退避表应安排在 ~5 分钟后
	mock.ExpectExec(`UPDATE tasks SET status=\?`).WithArgs(statusFailed, 2, sqlmock.AnyArg(), "T-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	before := time.Now()
	handleTaskFailure(context.Background(), svc, ord, &Task{TaskID: "T-1", RetryCount: 1})

	if len(ord.calls) != 0 {
		t.Fatalf("order should not be marked failed while retries remain, got %v", ord.calls)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
	_ = before
}

func TestNextRetryAt_BackoffTable(t *testing.T) {
	cases := []struct {
		retryCount int
		want       time.Duration
	}{
		{0, 1 * time.Minute},
		{1, 1 * time.Minute},  // backoff[0]
		{2, 5 * time.Minute},  // backoff[1]
		{3, 15 * time.Minute}, // backoff[2]
		{5, 15 * time.Minute}, // 越界回落到最后档
	}
	for _, c := range cases {
		before := time.Now()
		got := nextRetryAt(c.retryCount)
		elapsed := got.Sub(before)
		if elapsed < c.want-2*time.Second || elapsed > c.want+2*time.Second {
			t.Fatalf("nextRetryAt(%d) = %v after now, want ~%v", c.retryCount, elapsed, c.want)
		}
	}
}

// ---- compensation worker ----

func TestExecuteCompensation_Dispatch(t *testing.T) {
	ord := &fakeCancelUpdater{}
	inv := &fakeCompensationInventory{}
	ctx := context.Background()
	payload := StepPayload{OrderID: "O-1", ReservedID: "R-1", From: "RESERVED", To: "FAILED"}

	if err := executeCompensation(ctx, ord, inv, stepOrderFail, payload); err != nil {
		t.Fatalf("order_fail: %v", err)
	}
	if err := executeCompensation(ctx, ord, inv, stepOrderCancel, payload); err != nil {
		t.Fatalf("order_cancel: %v", err)
	}
	if err := executeCompensation(ctx, ord, inv, stepInvRelease, payload); err != nil {
		t.Fatalf("inv_release: %v", err)
	}
	if err := executeCompensation(ctx, ord, inv, stepInvReleaseOrder, payload); err != nil {
		t.Fatalf("inv_release_order: %v", err)
	}
	if err := executeCompensation(ctx, ord, inv, "UNKNOWN_STEP", payload); err != nil {
		t.Fatalf("unknown step should be no-op, got %v", err)
	}
	if err := executeCompensation(ctx, nil, nil, stepOrderCancel, payload); err != nil {
		t.Fatalf("nil deps should be no-op, got %v", err)
	}

	if len(ord.calls) != 1 || ord.calls[0] != "O-1:RESERVED->FAILED" {
		t.Fatalf("order calls = %v", ord.calls)
	}
	if len(ord.canceled) != 1 || ord.canceled[0] != "O-1" {
		t.Fatalf("canceled = %v", ord.canceled)
	}
	if len(inv.released) != 1 || inv.released[0] != "R-1" {
		t.Fatalf("released = %v", inv.released)
	}
	if len(inv.releasedByOrder) != 1 || inv.releasedByOrder[0] != "O-1" {
		t.Fatalf("releasedByOrder = %v", inv.releasedByOrder)
	}
}

func TestExecuteCompensation_Error(t *testing.T) {
	ord := &fakeCancelUpdater{cancelErr: errors.New("cancel failed")}
	inv := &fakeCompensationInventory{releaseErr: errors.New("release failed")}

	if err := executeCompensation(context.Background(), ord, inv, stepOrderCancel, StepPayload{OrderID: "O-1"}); err == nil {
		t.Fatal("expected cancel error")
	}
	if err := executeCompensation(context.Background(), ord, inv, stepInvRelease, StepPayload{ReservedID: "R-1"}); err == nil {
		t.Fatal("expected release error")
	}
}

func sagaStepRows(steps ...SagaStep) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{"id", "saga_id", "step", "status", "reason", "payload", "next_step", "created_at", "updated_at"})
	for _, s := range steps {
		rows.AddRow(s.ID, s.SagaID, s.Step, s.Status, s.Reason, s.Payload, s.NextStep, s.CreatedAt, s.UpdatedAt)
	}
	return rows
}

func TestProcessCompensations_CompletesSaga(t *testing.T) {
	svc, mock := newTaskService(t)
	ord := &fakeCancelUpdater{}
	inv := &fakeCompensationInventory{}

	payload, _ := json.Marshal(StepPayload{OrderID: "O-1", From: "RESERVED", To: "FAILED"})
	step := SagaStep{SagaID: "S-1", Step: stepOrderFail, Status: stepPending, Payload: string(payload)}

	mock.ExpectQuery(`SELECT \* FROM saga_steps`).WillReturnRows(sagaStepRows(step))
	mock.ExpectExec(`UPDATE saga_steps SET status=\?`).WithArgs(stepRunning, "S-1", stepOrderFail, stepPending).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE saga_steps SET status=\?`).WithArgs(stepDone, "S-1", stepOrderFail).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE sagas SET status=\?`).WithArgs(sagaStatusComp, driver.Value(""), "S-1", sagaStatusDone).
		WillReturnResult(sqlmock.NewResult(0, 1))

	processCompensations(svc, ord, inv)

	if len(ord.calls) != 1 || ord.calls[0] != "O-1:RESERVED->FAILED" {
		t.Fatalf("order calls = %v", ord.calls)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestProcessCompensations_ChainsNextStep(t *testing.T) {
	svc, mock := newTaskService(t)
	inv := &fakeCompensationInventory{}

	payload, _ := json.Marshal(StepPayload{OrderID: "O-1"})
	step := SagaStep{SagaID: "S-1", Step: stepOrderCancel, Status: stepPending,
		Payload: string(payload), NextStep: stepInvReleaseOrder, Reason: "timeout"}

	mock.ExpectQuery(`SELECT \* FROM saga_steps`).WillReturnRows(sagaStepRows(step))
	mock.ExpectExec(`UPDATE saga_steps SET status=\?`).WithArgs(stepRunning, "S-1", stepOrderCancel, stepPending).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE saga_steps SET status=\?`).WithArgs(stepDone, "S-1", stepOrderCancel).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// NextStep 非空 → 追加下一步而不是终结 saga
	mock.ExpectExec(`INSERT INTO saga_steps`).WithArgs("S-1", stepInvReleaseOrder, stepPending, "timeout", string(payload), "").
		WillReturnResult(sqlmock.NewResult(1, 1))

	processCompensations(svc, &fakeCancelUpdater{}, inv)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestProcessCompensations_StepExecutionFails(t *testing.T) {
	svc, mock := newTaskService(t)
	ord := &fakeCancelUpdater{cancelErr: errors.New("cancel failed")}

	step := SagaStep{SagaID: "S-1", Step: stepOrderCancel, Status: stepPending,
		Payload: `{"order_id":"O-1"}`}

	mock.ExpectQuery(`SELECT \* FROM saga_steps`).WillReturnRows(sagaStepRows(step))
	mock.ExpectExec(`UPDATE saga_steps SET status=\?`).WithArgs(stepRunning, "S-1", stepOrderCancel, stepPending).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE saga_steps SET status=\?`).WithArgs(stepFailed, sqlmock.AnyArg(), "S-1", stepOrderCancel).
		WillReturnResult(sqlmock.NewResult(0, 1))

	processCompensations(svc, ord, &fakeCompensationInventory{})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestProcessCompensations_LostRace(t *testing.T) {
	svc, mock := newTaskService(t)

	step := SagaStep{SagaID: "S-1", Step: stepInvRelease, Status: stepPending, Payload: "{}"}
	mock.ExpectQuery(`SELECT \* FROM saga_steps`).WillReturnRows(sagaStepRows(step))
	// 另一个实例先抢占（affected=0），本实例不应执行任何补偿
	mock.ExpectExec(`UPDATE saga_steps SET status=\?`).WithArgs(stepRunning, "S-1", stepInvRelease, stepPending).
		WillReturnResult(sqlmock.NewResult(0, 0))

	inv := &fakeCompensationInventory{}
	processCompensations(svc, &fakeCancelUpdater{}, inv)

	if len(inv.released) != 0 {
		t.Fatalf("no compensation should run after losing race, got %v", inv.released)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

// ---- HTTP handler ----

func newTaskRouter(svc *Service) *gin.Engine {
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

func taskRows(tasks ...Task) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{"id", "task_id", "biz_no", "order_id", "type", "status", "retry_count", "next_retry_at", "created_at", "updated_at"})
	for _, t := range tasks {
		rows.AddRow(t.ID, t.TaskID, t.BizNo, t.OrderID, t.Type, t.Status, t.RetryCount, t.NextRetryAt, t.CreatedAt, t.UpdatedAt)
	}
	return rows
}

func TestHandler_CreateTask(t *testing.T) {
	svc, mock := newTaskService(t)
	r := newTaskRouter(svc)

	mock.ExpectExec(`INSERT INTO tasks`).
		WithArgs(sqlmock.AnyArg(), "BIZ-1", "O-1", taskTypeFulfill, statusPending, 0, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	w := doJSON(r, http.MethodPost, "/tasks", `{"biz_no":"BIZ-1","order_id":"O-1","type":"FULFILL"}`)
	if !strings.Contains(w.Body.String(), `"code":0`) {
		t.Fatalf("body = %s", w.Body.String())
	}

	// 参数缺失
	w = doJSON(r, http.MethodPost, "/tasks", `{"biz_no":"BIZ-1"}`)
	if !strings.Contains(w.Body.String(), `"code":40001`) {
		t.Fatalf("missing fields should be 40001, body = %s", w.Body.String())
	}
}

func TestHandler_GetTask_NotFound(t *testing.T) {
	svc, mock := newTaskService(t)
	r := newTaskRouter(svc)

	mock.ExpectQuery(`SELECT \* FROM tasks WHERE task_id`).WithArgs("T-404").
		WillReturnRows(taskRows())

	w := doJSON(r, http.MethodGet, "/tasks/T-404", "")
	if !strings.Contains(w.Body.String(), `"code":40401`) {
		t.Fatalf("body = %s", w.Body.String())
	}
}

func TestHandler_RetryTask(t *testing.T) {
	svc, mock := newTaskService(t)
	r := newTaskRouter(svc)

	mock.ExpectExec(`UPDATE tasks SET retry_count`).WithArgs("T-1").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT \* FROM tasks WHERE task_id`).WithArgs("T-1").
		WillReturnRows(taskRows(Task{TaskID: "T-1", BizNo: "BIZ-1", OrderID: "O-1", Type: "FULFILL", Status: statusFailed, RetryCount: 2}))

	w := doJSON(r, http.MethodPost, "/tasks/T-1/retry", "")
	if !strings.Contains(w.Body.String(), `"retry_count":2`) {
		t.Fatalf("body = %s", w.Body.String())
	}
}

func TestHandler_GetTaskByOrder(t *testing.T) {
	svc, mock := newTaskService(t)
	r := newTaskRouter(svc)

	mock.ExpectQuery(`SELECT \* FROM tasks WHERE order_id`).WithArgs("O-1", taskTypeFulfill).
		WillReturnRows(taskRows(Task{TaskID: "T-1", BizNo: "BIZ-1", OrderID: "O-1", Type: taskTypeFulfill, Status: statusSuccess}))

	w := doJSON(r, http.MethodGet, "/tasks/by-order/O-1", "")
	if !strings.Contains(w.Body.String(), `"task_id":"T-1"`) {
		t.Fatalf("body = %s", w.Body.String())
	}
}

// ---- gRPC server ----

func TestGRPCServer_GetTaskByOrder(t *testing.T) {
	svc, mock := newTaskService(t)
	srv := NewGRPCServer(svc)

	mock.ExpectQuery(`SELECT \* FROM tasks WHERE order_id`).WithArgs("O-1", taskTypeFulfill).
		WillReturnRows(taskRows(Task{TaskID: "T-1", BizNo: "BIZ-1", OrderID: "O-1", Type: taskTypeFulfill,
			Status: statusRunning, RetryCount: 1, NextRetryAt: time.Unix(1700000000, 0)}))

	resp, err := srv.GetTaskByOrder(context.Background(), &taskpb.GetTaskByOrderRequest{OrderId: "O-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TaskId != "T-1" || resp.Status != statusRunning || resp.RetryCount != 1 || resp.NextRetryAt != 1700000000 {
		t.Fatalf("unexpected response: %+v", resp)
	}

	// 不存在的订单
	mock.ExpectQuery(`SELECT \* FROM tasks WHERE order_id`).WithArgs("O-404", taskTypeFulfill).
		WillReturnRows(taskRows())
	if _, err := srv.GetTaskByOrder(context.Background(), &taskpb.GetTaskByOrderRequest{OrderId: "O-404"}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}
