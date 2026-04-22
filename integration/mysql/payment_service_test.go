package mysqlit

import (
	"context"
	"testing"
	"time"

	"go-micro/internal/payment"
)

type fakeOrderCanceler struct {
	err error
}

func (f *fakeOrderCanceler) Cancel(ctx context.Context, orderID string) error {
	return f.err
}

type fakeInventoryReleaser struct {
	err error
}

func (f *fakeInventoryReleaser) ReleaseByOrder(ctx context.Context, orderID string) error {
	return f.err
}

func TestPaymentService_CreateAndGet(t *testing.T) {
	db, teardown := NewDB(t)
	defer teardown()

	orderCanceler := &fakeOrderCanceler{}
	invReleaser := &fakeInventoryReleaser{}
	svc := payment.NewService(db.DB, orderCanceler, invReleaser)

	t.Run("CreatePayment_Success", func(t *testing.T) {
		req := payment.CreatePaymentRequest{
			OrderID:   "order-001",
			Amount:    10000,
			RequestID: "req-" + randomID(),
		}
		p, err := svc.Create(req)
		if err != nil {
			t.Fatalf("Failed to create payment: %v", err)
		}
		if p.PaymentID == "" {
			t.Error("Expected non-empty paymentID")
		}
		if p.OrderID != req.OrderID {
			t.Errorf("Expected orderID %s, got %s", req.OrderID, p.OrderID)
		}
		if p.Amount != req.Amount {
			t.Errorf("Expected amount %d, got %d", req.Amount, p.Amount)
		}
		if p.Status != "CREATED" {
			t.Errorf("Expected status CREATED, got %s", p.Status)
		}
	})

	t.Run("CreatePayment_InvalidRequest_EmptyOrderID", func(t *testing.T) {
		req := payment.CreatePaymentRequest{
			OrderID:   "",
			Amount:    10000,
			RequestID: "req-" + randomID(),
		}
		_, err := svc.Create(req)
		if err == nil {
			t.Error("Expected error for empty orderID")
		}
	})

	t.Run("CreatePayment_InvalidRequest_ZeroAmount", func(t *testing.T) {
		req := payment.CreatePaymentRequest{
			OrderID:   "order-001",
			Amount:    0,
			RequestID: "req-" + randomID(),
		}
		_, err := svc.Create(req)
		if err == nil {
			t.Error("Expected error for zero amount")
		}
	})

	t.Run("CreatePayment_IdempotentHit", func(t *testing.T) {
		reqID := "req-idempotent-" + randomID()
		req := payment.CreatePaymentRequest{
			OrderID:   "order-001",
			Amount:    10000,
			RequestID: reqID,
		}
		p1, err := svc.Create(req)
		if err != nil {
			t.Fatalf("Failed to create first payment: %v", err)
		}

		p2, err := svc.Create(req)
		if err != payment.ErrIdempotentHit {
			t.Errorf("Expected ErrIdempotentHit, got %v", err)
		}
		if p1.PaymentID != p2.PaymentID {
			t.Error("Expected same paymentID on idempotent hit")
		}
	})

	t.Run("GetPayment_Success", func(t *testing.T) {
		req := payment.CreatePaymentRequest{
			OrderID:   "order-002",
			Amount:    20000,
			RequestID: "req-get-" + randomID(),
		}
		created, err := svc.Create(req)
		if err != nil {
			t.Fatalf("Failed to create payment: %v", err)
		}

		p, err := svc.Get(created.PaymentID)
		if err != nil {
			t.Fatalf("Failed to get payment: %v", err)
		}
		if p.PaymentID != created.PaymentID {
			t.Errorf("Expected paymentID %s, got %s", created.PaymentID, p.PaymentID)
		}
	})

	t.Run("GetPayment_NotFound", func(t *testing.T) {
		_, err := svc.Get("nonexistent-payment-id")
		if err != payment.ErrNotFound {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})
}

func TestPaymentService_StatusTransitions(t *testing.T) {
	db, teardown := NewDB(t)
	defer teardown()

	orderCanceler := &fakeOrderCanceler{}
	invReleaser := &fakeInventoryReleaser{}
	svc := payment.NewService(db.DB, orderCanceler, invReleaser)

	t.Run("MarkSuccess_FromCreated", func(t *testing.T) {
		req := payment.CreatePaymentRequest{
			OrderID:   "order-success-" + randomID(),
			Amount:    10000,
			RequestID: "req-success-" + randomID(),
		}
		p, err := svc.Create(req)
		if err != nil {
			t.Fatalf("Failed to create payment: %v", err)
		}

		err = svc.MarkSuccess(p.PaymentID)
		if err != nil {
			t.Fatalf("Failed to mark success: %v", err)
		}

		p2, err := svc.Get(p.PaymentID)
		if err != nil {
			t.Fatalf("Failed to get payment: %v", err)
		}
		if p2.Status != "SUCCESS" {
			t.Errorf("Expected status SUCCESS, got %s", p2.Status)
		}
	})

	t.Run("MarkFailed_FromCreated", func(t *testing.T) {
		req := payment.CreatePaymentRequest{
			OrderID:   "order-failed-" + randomID(),
			Amount:    10000,
			RequestID: "req-failed-" + randomID(),
		}
		p, err := svc.Create(req)
		if err != nil {
			t.Fatalf("Failed to create payment: %v", err)
		}

		err = svc.MarkFailed(p.PaymentID)
		if err != nil {
			t.Fatalf("Failed to mark failed: %v", err)
		}

		p2, err := svc.Get(p.PaymentID)
		if err != nil {
			t.Fatalf("Failed to get payment: %v", err)
		}
		if p2.Status != "FAILED" {
			t.Errorf("Expected status FAILED, got %s", p2.Status)
		}
	})

	t.Run("MarkTimeout_FromCreated", func(t *testing.T) {
		req := payment.CreatePaymentRequest{
			OrderID:   "order-timeout-" + randomID(),
			Amount:    10000,
			RequestID: "req-timeout-" + randomID(),
		}
		p, err := svc.Create(req)
		if err != nil {
			t.Fatalf("Failed to create payment: %v", err)
		}

		err = svc.MarkTimeout(p.PaymentID)
		if err != nil {
			t.Fatalf("Failed to mark timeout: %v", err)
		}

		p2, err := svc.Get(p.PaymentID)
		if err != nil {
			t.Fatalf("Failed to get payment: %v", err)
		}
		if p2.Status != "TIMEOUT" {
			t.Errorf("Expected status TIMEOUT, got %s", p2.Status)
		}
	})

	t.Run("Refund_FromSuccess", func(t *testing.T) {
		req := payment.CreatePaymentRequest{
			OrderID:   "order-refund-" + randomID(),
			Amount:    10000,
			RequestID: "req-refund-" + randomID(),
		}
		p, err := svc.Create(req)
		if err != nil {
			t.Fatalf("Failed to create payment: %v", err)
		}

		err = svc.MarkSuccess(p.PaymentID)
		if err != nil {
			t.Fatalf("Failed to mark success: %v", err)
		}

		err = svc.Refund(p.PaymentID)
		if err != nil {
			t.Fatalf("Failed to refund: %v", err)
		}

		p2, err := svc.Get(p.PaymentID)
		if err != nil {
			t.Fatalf("Failed to get payment: %v", err)
		}
		if p2.Status != "REFUNDED" {
			t.Errorf("Expected status REFUNDED, got %s", p2.Status)
		}
	})

	t.Run("InvalidTransition_CreatedToRefunded", func(t *testing.T) {
		req := payment.CreatePaymentRequest{
			OrderID:   "order-invalid-" + randomID(),
			Amount:    10000,
			RequestID: "req-invalid-" + randomID(),
		}
		p, err := svc.Create(req)
		if err != nil {
			t.Fatalf("Failed to create payment: %v", err)
		}

		// Cannot refund directly from CREATED
		err = svc.Refund(p.PaymentID)
		if err != payment.ErrInvalidState {
			t.Errorf("Expected ErrInvalidState, got %v", err)
		}
	})

	t.Run("InvalidTransition_SuccessToFailed", func(t *testing.T) {
		req := payment.CreatePaymentRequest{
			OrderID:   "order-invalid2-" + randomID(),
			Amount:    10000,
			RequestID: "req-invalid2-" + randomID(),
		}
		p, err := svc.Create(req)
		if err != nil {
			t.Fatalf("Failed to create payment: %v", err)
		}

		err = svc.MarkSuccess(p.PaymentID)
		if err != nil {
			t.Fatalf("Failed to mark success: %v", err)
		}

		// Cannot mark failed after success
		err = svc.MarkFailed(p.PaymentID)
		if err != payment.ErrInvalidState {
			t.Errorf("Expected ErrInvalidState, got %v", err)
		}
	})
}

func TestPaymentService_ListTimeoutPayments(t *testing.T) {
	db, teardown := NewDB(t)
	defer teardown()

	orderCanceler := &fakeOrderCanceler{}
	invReleaser := &fakeInventoryReleaser{}
	svc := payment.NewService(db.DB, orderCanceler, invReleaser)

	// Create some payments
	for i := 0; i < 5; i++ {
		req := payment.CreatePaymentRequest{
			OrderID:   "order-timeout-list-" + randomID(),
			Amount:    10000,
			RequestID: "req-list-" + randomID(),
		}
		_, err := svc.Create(req)
		if err != nil {
			t.Fatalf("Failed to create payment: %v", err)
		}
	}

	payments, err := svc.ListTimeoutPayments(time.Now().Add(time.Hour), 10)
	if err != nil {
		t.Fatalf("Failed to list timeout payments: %v", err)
	}
	if len(payments) != 5 {
		t.Errorf("Expected 5 payments, got %d", len(payments))
	}
}
