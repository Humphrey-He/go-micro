package task

import (
	"context"
	"encoding/json"
	"time"

	"go-micro/pkg/mq"
)

type InventoryReleaser interface {
	Release(ctx context.Context, reservedID string) error
}

type OrderUpdater interface {
	UpdateStatus(ctx context.Context, orderID, from, to string) error
}

type OrderReader interface {
	GetStatus(ctx context.Context, orderID string) (string, error)
}

type OrderCanceler interface {
	Cancel(ctx context.Context, orderID string) error
}

type OrderCreatedEvent struct {
	OrderID    string `json:"order_id"`
	BizNo      string `json:"biz_no"`
	Status     string `json:"status"`
	UserID     string `json:"user_id"`
	ReservedID string `json:"reserved_id"`
}

func StartRabbitConsumer(r *mq.Rabbit, svc *Service, inv InventoryReleaser, ord OrderUpdater) {
	if r == nil {
		return
	}
	msgs, err := r.Consume()
	if err != nil {
		return
	}
	for msg := range msgs {
		var evt OrderCreatedEvent
		if err := json.Unmarshal(msg.Body, &evt); err != nil {
			_ = msg.Nack(false, false)
			continue
		}
		if evt.OrderID == "" || evt.BizNo == "" {
			_ = msg.Nack(false, false)
			continue
		}

		_ = svc.CreateSaga(evt.OrderID, evt.BizNo, sagaTypeOrder)

		t, err := svc.Create(CreateTaskRequest{BizNo: evt.BizNo, OrderID: evt.OrderID, Type: taskTypeFulfill})
		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			if err := processFulfillment(ctx, svc, ord, t); err != nil {
				handleTaskFailure(ctx, svc, ord, t)
			}
			cancel()
			_ = svc.MarkSagaCompleted(evt.OrderID)
			_, _ = svc.CreateTimeoutTask(evt.OrderID, evt.BizNo, 15*time.Minute)
			_ = msg.Ack(false)
			continue
		}

		// Compensation: release reserved inventory on task creation failure
		if inv != nil && evt.ReservedID != "" {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			_ = inv.Release(ctx, evt.ReservedID)
			cancel()
			if ord != nil {
				ctx2, cancel2 := context.WithTimeout(context.Background(), 3*time.Second)
				_ = ord.UpdateStatus(ctx2, evt.OrderID, orderStatusReserved, orderStatusFailed)
				cancel2()
			}
			_ = svc.MarkSagaCompensated(evt.OrderID, "task_create_failed")
			_ = msg.Ack(false)
			continue
		}

		_ = msg.Nack(false, true)
	}
}
