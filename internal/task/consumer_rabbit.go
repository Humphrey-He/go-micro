package task

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go-micro/pkg/idemp"
	"go-micro/pkg/mq"
)

// consumerGroupTaskOrderCreated task 服务消费 order.created 的消费组名
const consumerGroupTaskOrderCreated = "task-order-created"

// messageIDOf MQ 消息的业务标识：优先 MessageId，缺失时退化为 body 摘要。
func messageIDOf(msg amqp.Delivery) string {
	if msg.MessageId != "" {
		return msg.MessageId
	}
	sum := sha256.Sum256(msg.Body)
	return "body-" + hex.EncodeToString(sum[:])[:24]
}

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

func StartRabbitConsumer(r *mq.Rabbit, svc *Service, inv InventoryReleaser, ord OrderUpdater, recorder *idemp.ConsumeRecorder) {
	if r == nil {
		return
	}
	msgs, err := r.Consume()
	if err != nil {
		return
	}
	for msg := range msgs {
		// 消费端幂等：同一条消息重复投递（at-least-once）直接 ACK 吸收
		if recorder != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			first, err := recorder.TryRecord(ctx, consumerGroupTaskOrderCreated, messageIDOf(msg))
			cancel()
			if err == nil && !first {
				_ = msg.Ack(false)
				continue
			}
		}
		// Consume order_reserved event to create fulfillment task.
		var evt OrderCreatedEvent
		if err := json.Unmarshal(msg.Body, &evt); err != nil {
			_ = msg.Nack(false, false)
			continue
		}
		if evt.OrderID == "" || evt.BizNo == "" {
			_ = msg.Nack(false, false)
			continue
		}

		// Record saga for tracing and compensation.
		_ = svc.CreateSaga(evt.OrderID, evt.BizNo, sagaTypeOrder)

		t, err := svc.Create(CreateTaskRequest{BizNo: evt.BizNo, OrderID: evt.OrderID, Type: taskTypeFulfill})
		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			// Execute fulfillment and update order status.
			if err := processFulfillment(ctx, svc, ord, t); err != nil {
				handleTaskFailure(ctx, svc, ord, t)
			}
			cancel()
			_ = svc.MarkSagaCompleted(evt.OrderID)
			// Schedule timeout cancel task for compensation if not completed.
			_, _ = svc.CreateTimeoutTask(evt.OrderID, evt.BizNo, 15*time.Minute)
			_ = msg.Ack(false)
			continue
		}

		// Compensation: enqueue saga steps on task creation failure
		if evt.ReservedID != "" {
			payload := `{"order_id":"` + evt.OrderID + `","reserved_id":"` + evt.ReservedID + `","from":"` + orderStatusReserved + `","to":"` + orderStatusFailed + `"}`
			_ = svc.CreateSagaStep(evt.OrderID, stepOrderFail, stepInvRelease, "task_create_failed", payload)
			_ = msg.Ack(false)
			continue
		}

		_ = msg.Nack(false, true)
	}
}
