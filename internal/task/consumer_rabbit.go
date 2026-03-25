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

		_, err := svc.Create(CreateTaskRequest{BizNo: evt.BizNo, OrderID: evt.OrderID, Type: taskTypeFulfill})
		if err == nil {
			if ord != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				_ = ord.UpdateStatus(ctx, evt.OrderID, "RESERVED", "PROCESSING")
				_ = ord.UpdateStatus(ctx, evt.OrderID, "PROCESSING", "SUCCESS")
				cancel()
			}
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
				_ = ord.UpdateStatus(ctx2, evt.OrderID, "RESERVED", "FAILED")
				cancel2()
			}
			_ = msg.Ack(false)
			continue
		}

		_ = msg.Nack(false, true)
	}
}
