package mysqlit

import (
	"context"
	"testing"

	"go-micro/internal/order"
)

type fakeInventory struct {
	reservedID string
	err        error
}

func (f *fakeInventory) Reserve(ctx context.Context, orderID string, items []order.Item) (string, error) {
	return f.reservedID, f.err
}

type fakePublisher struct {
	last []byte
	err  error
}

func (f *fakePublisher) Publish(ctx context.Context, body []byte) error {
	f.last = body
	return f.err
}

func TestOrderService_CreateAndGet(t *testing.T) {
	db, teardown := NewDB(t)
	defer teardown()

	inv := &fakeInventory{reservedID: "RESV-123", err: nil}
	pub := &fakePublisher{}
	svc := order.NewService(db.DB, db.RDB, inv, pub)

	t.Run("CreateOrder_Success", func(t *testing.T) {
		req := order.CreateOrderRequest{
			RequestID: "req-" + randomID(),
			UserID:    "user-123",
			Items: []order.Item{
				{SkuID: "SKU-001", Quantity: 2, Price: 1000},
				{SkuID: "SKU-002", Quantity: 1, Price: 2000},
			},
			Remark: "test order",
		}
		resp, err := svc.Create(req)
		if err != nil {
			t.Fatalf("Failed to create order: %v", err)
		}
		if resp.OrderID == "" {
			t.Error("Expected non-empty orderID")
		}
		if resp.BizNo == "" {
			t.Error("Expected non-empty bizNo")
		}
		if resp.Status != "RESERVED" {
			t.Errorf("Expected status RESERVED, got %s", resp.Status)
		}
	})

	t.Run("CreateOrder_InvalidRequest_EmptyRequestID", func(t *testing.T) {
		req := order.CreateOrderRequest{
			RequestID: "",
			UserID:    "user-123",
			Items: []order.Item{
				{SkuID: "SKU-001", Quantity: 2, Price: 1000},
			},
		}
		_, err := svc.Create(req)
		if err == nil {
			t.Error("Expected error for empty requestID")
		}
	})

	t.Run("CreateOrder_InvalidRequest_NoItems", func(t *testing.T) {
		req := order.CreateOrderRequest{
			RequestID: "req-" + randomID(),
			UserID:    "user-123",
			Items:     []order.Item{},
		}
		_, err := svc.Create(req)
		if err == nil {
			t.Error("Expected error for empty items")
		}
	})

	t.Run("CreateOrder_IdempotentHit", func(t *testing.T) {
		reqID := "req-idempotent-" + randomID()
		req := order.CreateOrderRequest{
			RequestID: reqID,
			UserID:    "user-123",
			Items: []order.Item{
				{SkuID: "SKU-001", Quantity: 1, Price: 1000},
			},
		}
		resp1, err := svc.Create(req)
		if err != nil {
			t.Fatalf("Failed to create first order: %v", err)
		}

		resp2, err := svc.Create(req)
		if err != order.ErrIdempotentHit {
			t.Errorf("Expected ErrIdempotentHit, got %v", err)
		}
		if resp1.OrderID != resp2.OrderID {
			t.Error("Expected same orderID on idempotent hit")
		}
	})

	t.Run("GetOrder_Success", func(t *testing.T) {
		req := order.CreateOrderRequest{
			RequestID: "req-get-" + randomID(),
			UserID:    "user-123",
			Items: []order.Item{
				{SkuID: "SKU-001", Quantity: 1, Price: 1000},
			},
		}
		created, err := svc.Create(req)
		if err != nil {
			t.Fatalf("Failed to create order: %v", err)
		}

		ord, err := svc.Get(created.OrderID)
		if err != nil {
			t.Fatalf("Failed to get order: %v", err)
		}
		if ord.OrderID != created.OrderID {
			t.Errorf("Expected orderID %s, got %s", created.OrderID, ord.OrderID)
		}
		if len(ord.Items) != 1 {
			t.Errorf("Expected 1 item, got %d", len(ord.Items))
		}
	})

	t.Run("GetOrder_NotFound", func(t *testing.T) {
		_, err := svc.Get("nonexistent-order-id")
		if err != order.ErrNotFound {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})

	t.Run("GetByBizNo_Success", func(t *testing.T) {
		req := order.CreateOrderRequest{
			RequestID: "req-bizno-" + randomID(),
			UserID:    "user-123",
			Items: []order.Item{
				{SkuID: "SKU-001", Quantity: 1, Price: 1000},
			},
		}
		created, err := svc.Create(req)
		if err != nil {
			t.Fatalf("Failed to create order: %v", err)
		}

		ord, err := svc.GetByBizNo(created.BizNo)
		if err != nil {
			t.Fatalf("Failed to get order by bizNo: %v", err)
		}
		if ord.OrderID != created.OrderID {
			t.Errorf("Expected orderID %s, got %s", created.OrderID, ord.OrderID)
		}
	})

	t.Run("Cancel_Success", func(t *testing.T) {
		req := order.CreateOrderRequest{
			RequestID: "req-cancel-" + randomID(),
			UserID:    "user-123",
			Items: []order.Item{
				{SkuID: "SKU-001", Quantity: 1, Price: 1000},
			},
		}
		created, err := svc.Create(req)
		if err != nil {
			t.Fatalf("Failed to create order: %v", err)
		}

		err = svc.Cancel(created.OrderID)
		if err != nil {
			t.Fatalf("Failed to cancel order: %v", err)
		}

		ord, err := svc.Get(created.OrderID)
		if err != nil {
			t.Fatalf("Failed to get order after cancel: %v", err)
		}
		if ord.Status != "CANCELED" {
			t.Errorf("Expected status CANCELED, got %s", ord.Status)
		}
	})
}

func TestOrderService_InventoryFailure(t *testing.T) {
	db, teardown := NewDB(t)
	defer teardown()

	inv := &fakeInventory{reservedID: "", err: order.ErrInventoryFail}
	pub := &fakePublisher{}
	svc := order.NewService(db.DB, db.RDB, inv, pub)

	t.Run("CreateOrder_InventoryFailure", func(t *testing.T) {
		req := order.CreateOrderRequest{
			RequestID: "req-invfail-" + randomID(),
			UserID:    "user-123",
			Items: []order.Item{
				{SkuID: "SKU-001", Quantity: 1, Price: 1000},
			},
		}
		_, err := svc.Create(req)
		if err != order.ErrInventoryFail {
			t.Errorf("Expected ErrInventoryFail, got %v", err)
		}
	})
}

func TestOrderService_CanTransition(t *testing.T) {
	db, teardown := NewDB(t)
	defer teardown()

	inv := &fakeInventory{reservedID: "RESV-123", err: nil}
	pub := &fakePublisher{}
	svc := order.NewService(db.DB, db.RDB, inv, pub)

	// After Create, order is in RESERVED status (not CREATED)
	// Test transitions from RESERVED state
	reservedTransitions := []struct {
		from   string
		to     string
		expect bool
	}{
		{"RESERVED", "PROCESSING", true},
		{"RESERVED", "FAILED", true},
		{"RESERVED", "CANCELED", true},
	}

	for _, vt := range reservedTransitions {
		t.Run(vt.from+"_to_"+vt.to, func(t *testing.T) {
			req := order.CreateOrderRequest{
				RequestID: "req-" + vt.from + "-" + vt.to + "-" + randomID(),
				UserID:    "user-123",
				Items: []order.Item{
					{SkuID: "SKU-001", Quantity: 1, Price: 1000},
				},
			}
			created, err := svc.Create(req)
			if err != nil {
				t.Fatalf("Failed to create order: %v", err)
			}

			err = svc.UpdateStatus(created.OrderID, vt.from, vt.to)
			if vt.expect && err != nil {
				t.Errorf("Expected transition %s->%s to succeed, got %v", vt.from, vt.to, err)
			}
		})
	}

	// Test invalid transition from RESERVED
	t.Run("RESERVED_to_SUCCESS_Invalid", func(t *testing.T) {
		req := order.CreateOrderRequest{
			RequestID: "req-reserved-success-" + randomID(),
			UserID:    "user-123",
			Items: []order.Item{
				{SkuID: "SKU-001", Quantity: 1, Price: 1000},
			},
		}
		created, err := svc.Create(req)
		if err != nil {
			t.Fatalf("Failed to create order: %v", err)
		}

		// RESERVED -> SUCCESS is not allowed
		err = svc.UpdateStatus(created.OrderID, "RESERVED", "SUCCESS")
		if err != order.ErrInvalidState {
			t.Errorf("Expected ErrInvalidState, got %v", err)
		}
	})
}
