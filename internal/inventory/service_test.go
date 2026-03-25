package inventory

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

func newInventoryService(t *testing.T) (*Service, sqlmock.Sqlmock, *miniredis.Miniredis) {
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

func TestReserveInventory_Success(t *testing.T) {
	svc, mock, mr := newInventoryService(t)
	defer mr.Close()

	req := ReserveRequest{OrderID: "O-1", Items: []Item{{SkuID: "SKU-1001", Quantity: 2}}}

	mock.ExpectBegin()
	rows := sqlmock.NewRows([]string{"id", "sku_id", "available", "reserved", "updated_at"}).
		AddRow(1, "SKU-1001", 10, 0, time.Now())
	mock.ExpectQuery(`SELECT \* FROM inventory WHERE sku_id = \? FOR UPDATE`).
		WithArgs("SKU-1001").
		WillReturnRows(rows)
	mock.ExpectExec("UPDATE inventory SET available = available -").
		WithArgs(2, 2, "SKU-1001").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO inventory_reserved").
		WithArgs(sqlmock.AnyArg(), "O-1", resvReserved).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO inventory_reserved_item").
		WithArgs(sqlmock.AnyArg(), "SKU-1001", 2).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	_, err := svc.Reserve(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetInventory_NotFound(t *testing.T) {
	svc, mock, mr := newInventoryService(t)
	defer mr.Close()

	mock.ExpectQuery(`SELECT \* FROM inventory WHERE sku_id = \?`).
		WithArgs("SKU-404").
		WillReturnError(sql.ErrNoRows)

	_, err := svc.GetInventory("SKU-404")
	if err != ErrSkuNotFound {
		t.Fatalf("expected sku not found, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestReleaseByOrder_Idempotent(t *testing.T) {
	svc, mock, mr := newInventoryService(t)
	defer mr.Close()

	// First release: RESERVED -> RELEASED
	mock.ExpectQuery(`SELECT \* FROM inventory_reserved WHERE order_id = \? ORDER BY created_at DESC LIMIT 1`).
		WithArgs("O-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "reserved_id", "order_id", "status", "created_at", "updated_at"}).
			AddRow(1, "R-1", "O-1", resvReserved, time.Now(), time.Now()))

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM inventory_reserved WHERE reserved_id = \? FOR UPDATE`).
		WithArgs("R-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "reserved_id", "order_id", "status", "created_at", "updated_at"}).
			AddRow(1, "R-1", "O-1", resvReserved, time.Now(), time.Now()))
	mock.ExpectQuery(`SELECT sku_id,quantity FROM inventory_reserved_item WHERE reserved_id = \?`).
		WithArgs("R-1").
		WillReturnRows(sqlmock.NewRows([]string{"sku_id", "quantity"}).
			AddRow("SKU-1", 2))
	mock.ExpectExec(`UPDATE inventory SET available = available \+ \?, reserved = reserved - \?, updated_at = NOW\(\) WHERE sku_id = \?`).
		WithArgs(2, 2, "SKU-1").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`UPDATE inventory_reserved SET status=\?, updated_at=NOW\(\) WHERE reserved_id = \?`).
		WithArgs(resvReleased, "R-1").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := svc.ReleaseByOrder("O-1"); err != nil {
		t.Fatalf("first release error: %v", err)
	}

	// Second release: already RELEASED -> idempotent success
	mock.ExpectQuery(`SELECT \* FROM inventory_reserved WHERE order_id = \? ORDER BY created_at DESC LIMIT 1`).
		WithArgs("O-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "reserved_id", "order_id", "status", "created_at", "updated_at"}).
			AddRow(1, "R-1", "O-1", resvReleased, time.Now(), time.Now()))

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM inventory_reserved WHERE reserved_id = \? FOR UPDATE`).
		WithArgs("R-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "reserved_id", "order_id", "status", "created_at", "updated_at"}).
			AddRow(1, "R-1", "O-1", resvReleased, time.Now(), time.Now()))
	mock.ExpectCommit()

	if err := svc.ReleaseByOrder("O-1"); err != nil {
		t.Fatalf("second release error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestReleaseByOrder_NotFound(t *testing.T) {
	svc, mock, mr := newInventoryService(t)
	defer mr.Close()

	mock.ExpectQuery(`SELECT \* FROM inventory_reserved WHERE order_id = \? ORDER BY created_at DESC LIMIT 1`).
		WithArgs("O-404").
		WillReturnError(sql.ErrNoRows)

	err := svc.ReleaseByOrder("O-404")
	if err != ErrResvNotFound {
		t.Fatalf("expected %v, got %v", ErrResvNotFound, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
