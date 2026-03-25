package refund

import (
	"context"
	"encoding/json"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Message struct {
	RefundID string `json:"refund_id"`
}

type Consumer interface {
	Consume() (<-chan amqp.Delivery, error)
}

func StartConsumer(c Consumer, svc *Service, stop <-chan struct{}) {
	if c == nil || svc == nil {
		return
	}
	ch, err := c.Consume()
	if err != nil {
		return
	}
	for {
		select {
		case <-stop:
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			processMessage(svc, msg)
		}
	}
}

func processMessage(svc *Service, msg amqp.Delivery) {
	var payload Message
	_ = json.Unmarshal(msg.Body, &payload)
	if payload.RefundID == "" {
		_ = msg.Nack(false, false)
		return
	}
	if err := svc.Process(payload.RefundID); err != nil {
		_ = msg.Nack(false, false)
		return
	}
	_ = msg.Ack(false)
}

type RetryStore interface {
	ListRetryDue(limit int) ([]Refund, error)
}

type PublisherRetry interface {
	Publish(ctx context.Context, body []byte) error
}

func StartRetryWorker(svc RetryStore, pub PublisherRetry, stop <-chan struct{}) {
	if svc == nil || pub == nil {
		return
	}
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			retries, err := svc.ListRetryDue(20)
			if err != nil {
				continue
			}
			for i := range retries {
				ref := retries[i]
				body, _ := json.Marshal(map[string]string{"refund_id": ref.RefundID})
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				_ = pub.Publish(ctx, body)
				cancel()
			}
		}
	}
}
