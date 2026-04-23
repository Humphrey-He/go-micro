package recommendation

import (
	"context"
	"log"
)

type Consumer struct {
	svc    *Service
	rabbit interface{}
}

func NewConsumer(svc *Service, rabbit interface{}) *Consumer {
	return &Consumer{
		svc:    svc,
		rabbit: rabbit,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	log.Println("recommendation consumer started")
}