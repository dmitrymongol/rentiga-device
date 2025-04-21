// rabbitmq/rabbitmq.go
package rabbitmq

import (
	"fmt"

	"github.com/streadway/amqp"
)

type Client struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func New(uri string) (*Client, error) {
	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %v", err)
	}

	return &Client{conn, channel}, nil
}

func (c *Client) Publish(queue string, body []byte) error {
	_, err := c.channel.QueueDeclare(
		queue,  // name
		true,   // durable
		false,  // delete when unused
		false,  // exclusive
		false,  // no-wait
		nil,    // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %v", err)
	}

	return c.channel.Publish(
		"",     // exchange
		queue,  // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (c *Client) Consume(queue string) (<-chan amqp.Delivery, error) {
	return c.channel.Consume(
		queue,  // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
}

func (c *Client) Close() {
	c.channel.Close()
	c.conn.Close()
}