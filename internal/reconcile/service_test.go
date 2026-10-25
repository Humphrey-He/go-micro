package reconcile

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func newReconcile(t *testing.T) (*Service, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	sqlxDB := sqlx.NewDb(db, "mysql")
	t.Cleanup(func() { sqlxDB.Close() })
	return NewService(sqlxDB), mock
}

func TestRunOnce_BothDirections(t *testing.T) {
	svc, mock := newReconcile(t)
	ctx := context.Background()

	// 方向 A：订单 O-1 占用库存但无预占凭据
	mock.ExpectQuery(`FROM orders o`).
		WillReturnRows(sqlmock.NewRows([]string{"order_id"}).AddRow("O-1"))
	mock.ExpectExec(`INSERT IGNORE INTO reconcile_diff`).
		WithArgs("O-1", DiffOrderNoReservation, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	// 方向 B：O-2 的预占成为幽灵占用
	mock.ExpectQuery(`FROM inventory_reserved r`).
		WillReturnRows(sqlmock.NewRows([]string{"order_id"}).AddRow("O-2"))
	mock.ExpectExec(`INSERT IGNORE INTO reconcile_diff`).
		WithArgs("O-2", DiffReservationOrphan, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(2, 1))

	report, err := svc.RunOnce(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.OrdersChecked != 1 || report.OrphansFound != 1 || report.NewDiffs != 2 {
		t.Fatalf("report = %+v", report)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRunOnce_Clean(t *testing.T) {
	svc, mock := newReconcile(t)

	mock.ExpectQuery(`FROM orders o`).WillReturnRows(sqlmock.NewRows([]string{"order_id"}))
	mock.ExpectQuery(`FROM inventory_reserved r`).WillReturnRows(sqlmock.NewRows([]string{"order_id"}))

	report, err := svc.RunOnce(context.Background())
	if err != nil || report.NewDiffs != 0 {
		t.Fatalf("report = %+v err = %v", report, err)
	}
}

func TestTransition_ResolveAndInvalid(t *testing.T) {
	svc, mock := newReconcile(t)
	ctx := context.Background()

	// PENDING → RESOLVED 成功
	mock.ExpectExec(`SET status = \?`).
		WithArgs(DiffResolved, int64(7), DiffPending).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := svc.Resolve(ctx, 7); err != nil {
		t.Fatalf("resolve: %v", err)
	}

	// 已是终态再流转：affected=0 → ErrInvalidTransition
	mock.ExpectExec(`SET status = \?`).
		WithArgs(DiffIgnored, int64(7), DiffPending).
		WillReturnResult(sqlmock.NewResult(0, 0))
	if err := svc.Ignore(ctx, 7); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("err = %v, want ErrInvalidTransition", err)
	}
}

func TestListDiffs_FilterByStatus(t *testing.T) {
	svc, mock := newReconcile(t)

	mock.ExpectQuery(`AND status = \?`).
		WithArgs(DiffPending, 50).
		WillReturnRows(sqlmock.NewRows([]string{"id", "biz_type", "sku_id", "request_id", "diff_type", "detail", "status", "created_at", "resolved_at"}).
			AddRow(3, "ORDER_INV", "", "O-9", DiffOrderNoReservation, "d", DiffPending, time.Now(), nil))

	diffs, err := svc.ListDiffs(context.Background(), DiffPending, 0)
	if err != nil || len(diffs) != 1 || diffs[0].RequestID != "O-9" {
		t.Fatalf("diffs=%v err=%v", diffs, err)
	}
}

// ---- handler ----

func TestHandler_RunAndResolve(t *testing.T) {
	svc, mock := newReconcile(t)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	NewHandler(svc).Register(r)

	// 触发对账
	mock.ExpectQuery(`FROM orders o`).WillReturnRows(sqlmock.NewRows([]string{"order_id"}))
	mock.ExpectQuery(`FROM inventory_reserved r`).WillReturnRows(sqlmock.NewRows([]string{"order_id"}))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/reconcile/run", nil))
	if !strings.Contains(w.Body.String(), `"code":0`) {
		t.Fatalf("run body = %s", w.Body.String())
	}

	// 关闭台账
	mock.ExpectExec(`SET status = \?`).WithArgs(DiffResolved, int64(7), DiffPending).
		WillReturnResult(sqlmock.NewResult(0, 1))
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/reconcile/diffs/7/resolve", nil))
	if !strings.Contains(w.Body.String(), `"status":"RESOLVED"`) {
		t.Fatalf("resolve body = %s", w.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
