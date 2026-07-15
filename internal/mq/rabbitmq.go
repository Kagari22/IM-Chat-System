package mq

import (
	"context"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitPublisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewRabbitPublisher(url string) (*RabbitPublisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	if err := declareTopology(ch); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}

	return &RabbitPublisher{conn: conn, channel: ch}, nil
}

func (p *RabbitPublisher) PublishMessageCreated(ctx context.Context, event MessageCreatedEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.channel.PublishWithContext(
		ctx,
		ExchangeChatEvents,
		RoutingKeyMessageCreated,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (p *RabbitPublisher) Close() error {
	if p.channel != nil {
		_ = p.channel.Close()
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

type MessageCreatedConsumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewMessageCreatedConsumer(url string) (*MessageCreatedConsumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	if err := declareTopology(ch); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}

	return &MessageCreatedConsumer{conn: conn, channel: ch}, nil
}

func (c *MessageCreatedConsumer) Start(ctx context.Context, handler MessageCreatedHandler) error {
	deliveries, err := c.channel.Consume(
		QueueMessageCreated,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case delivery, ok := <-deliveries:
				if !ok {
					return
				}

				var event MessageCreatedEvent
				if err := json.Unmarshal(delivery.Body, &event); err != nil {
					log.Println("rabbitmq unmarshal message created:", err)
					_ = delivery.Nack(false, false)
					continue
				}

				if handler != nil {
					if err := handler(ctx, event); err != nil {
						log.Println("rabbitmq handle message created:", err)
						_ = delivery.Nack(false, true)
						continue
					}
				}
				log.Printf("rabbitmq message.created consumed: message=%d from=%d to=%d\n", event.MessageID, event.FromUserID, event.ToUserID)
				_ = delivery.Ack(false)
			}
		}
	}()

	return nil
}

func (c *MessageCreatedConsumer) Close() error {
	if c.channel != nil {
		_ = c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func declareTopology(ch *amqp.Channel) error {
	if err := ch.ExchangeDeclare(
		ExchangeChatEvents,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}

	if _, err := ch.QueueDeclare(
		QueueMessageCreated,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}

	return ch.QueueBind(
		QueueMessageCreated,
		RoutingKeyMessageCreated,
		ExchangeChatEvents,
		false,
		nil,
	)
}
