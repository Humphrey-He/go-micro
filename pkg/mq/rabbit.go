package mq

import (
	"context"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Rabbit struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	Exchange string
	Queue    string
	RouteKey string
	DLX      string
	DLQ      string
}

func NewRabbit(url, exchange, queue, routeKey, dlx, dlq string) (*Rabbit, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	r := &Rabbit{conn: conn, channel: ch, Exchange: exchange, Queue: queue, RouteKey: routeKey, DLX: dlx, DLQ: dlq}
	if err := r.declare(); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}
	_ = ch.Qos(10, 0, false)
	return r, nil
}

func (r *Rabbit) declare() error {
	if err := r.channel.ExchangeDeclare(r.Exchange, "topic", true, false, false, false, nil); err != nil {
		return err
	}
	if err := r.channel.ExchangeDeclare(r.DLX, "topic", true, false, false, false, nil); err != nil {
		return err
	}
	args := amqp.Table{"x-dead-letter-exchange": r.DLX}
	if _, err := r.channel.QueueDeclare(r.Queue, true, false, false, false, args); err != nil {
		return err
	}
	if err := r.channel.QueueBind(r.Queue, r.RouteKey, r.Exchange, false, nil); err != nil {
		return err
	}
	if _, err := r.channel.QueueDeclare(r.DLQ, true, false, false, false, nil); err != nil {
		return err
	}
	if err := r.channel.QueueBind(r.DLQ, r.RouteKey, r.DLX, false, nil); err != nil {
		return err
	}
	return nil
}

func (r *Rabbit) Publish(ctx context.Context, body []byte) error {
	return r.channel.PublishWithContext(ctx, r.Exchange, r.RouteKey, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
		Timestamp:   time.Now(),
	})
}

func (r *Rabbit) Consume() (<-chan amqp.Delivery, error) {
	return r.channel.Consume(r.Queue, "", false, false, false, false, nil)
}

func (r *Rabbit) Close() {
	if r.channel != nil {
		_ = r.channel.Close()
	}
	if r.conn != nil {
		_ = r.conn.Close()
	}
}
