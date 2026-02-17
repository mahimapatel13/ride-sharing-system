package queue

import (
	"github.com/mahimapatel13/ride-sharing-system/internal/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

type MQChannel struct {
	QueueName  string
	Channel *amqp.Channel
}

// QueueOptions Represents the configuration options to declare a queue.
type QueueOptions struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
}

// ConnectRabbitMQ connects to RabbitMQ using url stored in config object
func ConnectRabbitMQ(config config.RabbitMQConfig) (*amqp.Connection, error) {
	conn, err := amqp.Dial(config.URL)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// CreateChannel creates a channel to communicate with RabbitMQ
func CreateChannel(conn *amqp.Connection, queueName string) (*MQChannel, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	channel := MQChannel{
		Channel: ch,
		QueueName: queueName,
	}
	return &channel, nil
}

// QueueDeclare declares a queue to hold messages and deliver to consumers.
// Declaring creates a queue if it doesn't already exist, or ensures that an
// existing queue matches the same parameters.
func DeclareQueue(ch *amqp.Channel, opt QueueOptions) error {
	_, err := ch.QueueDeclare(
		opt.Name,
		opt.Durable,
		opt.AutoDelete,
		opt.Exclusive,
		opt.NoWait,
		nil,
	)

	if err != nil {
		return err
	}
	return nil
}
