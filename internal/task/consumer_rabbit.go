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

type OrderCreatedEvent struct {
	OrderID    string `json:"order_id"`
	BizNo      string `json:"biz_no"`
	Status     string `json:"status"`
	UserID     string `json:"user_id"`
	ReservedID string `json:"reserved_id"`
}

func StartRabbitConsumer(r *mq.Rabbit, svc *Service, inv InventoryReleaser) {
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

		_, err := svc.Create(CreateTaskRequest{BizNo: evt.BizNo, OrderID: evt.OrderID, Type: "FULFILL"})
		if err == nil {
			_ = msg.Ack(false)
			continue
		}

		// Compensation: release reserved inventory on task creation failure
		if inv != nil && evt.ReservedID != "" {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			_ = inv.Release(ctx, evt.ReservedID)
			cancel()
			_ = msg.Ack(false)
			continue
		}

		_ = msg.Nack(false, true)
	}
}
