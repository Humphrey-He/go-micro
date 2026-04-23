package recommendation

import (
	"context"
	"log"

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
	log.Println("recommendation consumer started")
}