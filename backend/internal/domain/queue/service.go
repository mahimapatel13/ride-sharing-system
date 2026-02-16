package queue

import amqp "github.com/rabbitmq/amqp091-go"

// QueueService interface representes the methods
// any queue service should implement
type QueueService interface {
	PublishMessage(queueName, message string) error
}

// service struct implements QueueService interface
type service struct{
    mqChannel *amqp.Channel
    queueName string
}

// NewQueueService function initialises a new queue service 
func NewQueueService(ch *amqp.Channel, queueName string) QueueService{
    return &service{
        mqChannel: ch,
        queueName: queueName,
    }
}


// PublishMessage publishes persistent message via queue channel.
func(s *service) PublishMessage(queueName, message string) error{

    err := s.mqChannel.Publish(
        "",
        s.queueName,
        true,
        false,
        amqp.Publishing{
            ContentType: "text/plain",
            Body: []byte(message),
        },
    )

    return err
}