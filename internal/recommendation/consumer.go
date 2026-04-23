package recommendation

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"go-micro/pkg/mq"
)

type Consumer struct {
	svc    *Service
	rabbit *mq.Rabbit
}

func NewConsumer(svc *Service, rabbit *mq.Rabbit) *Consumer {
	return &Consumer{
		svc:    svc,
		rabbit: rabbit,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	msgs, err := c.rabbit.Consume()
	if err != nil {
		log.Printf("failed to start consumer: %v", err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("recommendation consumer stopped")
			return
		case msg, ok := <-msgs:
			if !ok {
				log.Println("consumer channel closed")
				return
			}

			if err := c.processMessage(ctx, msg.Body); err != nil {
				log.Printf("failed to process message: %v", err)
				msg.Nack(false, true) // Requeue
			} else {
				msg.Ack(false)
			}
		}
	}
}

func (c *Consumer) processMessage(ctx context.Context, body []byte) error {
	var req BehaviorReportRequest
	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("failed to unmarshal message: %v", err)
		return err
	}

	// Retry with backoff
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt*2) * time.Second)
		}

		err := c.svc.ReportBehavior(ctx, &req, 0)
		if err == nil {
			return nil
		}
		lastErr = err
		log.Printf("attempt %d failed: %v", attempt+1, err)
	}

	return lastErr
}