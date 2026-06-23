package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/Azure/go-amqp"
	rmq "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
	"github.com/uddinArsalan/devdeploy/internals/domain"
)

type RabbitMQAdapter struct {
	rmqClient *rmq.AmqpConnection
	Queue     string
}

func NewRabbitMQClient(ctx context.Context) (*RabbitMQAdapter, error) {
	brokerURI := os.Getenv("RABBITMQ_BROKER_URI")
	fmt.Printf("Broker URI %q", brokerURI)
	if brokerURI == "" {
		return nil, errors.New("RABBITMQ_BROKER_URI not set")
	}
	env := rmq.NewEnvironment(brokerURI, nil)
	conn, err := env.NewConnection(ctx)
	if err != nil {
		log.Printf("Failed to connect to RabbitMQ: %v", err)
		return nil, err
	}
	queue := os.Getenv("BUILD_QUEUE")
	_, err = conn.Management().DeclareQueue(ctx, &rmq.QuorumQueueSpecification{Name: queue})
	if err != nil {
		log.Printf("Failed to declare a queue: %v", err)
		return nil, err
	}

	return &RabbitMQAdapter{
		rmqClient: conn,
		Queue:     queue,
	}, nil
}

func (r *RabbitMQAdapter) Close(ctx context.Context) error {
	return r.rmqClient.Close(ctx)
}

func (rm *RabbitMQAdapter) PublishMessage(ctx context.Context, job domain.BuildJob) error {
	publisher, err := rm.rmqClient.NewPublisher(ctx, &rmq.QueueAddress{
		Queue: rm.Queue,
	}, nil)
	if err != nil {
		return err
	}
	defer func() { _ = publisher.Close(ctx) }()
	data, err := json.Marshal(job)
	if err != nil {
		log.Printf("Failed to publish a message: %v", err)
		return err
	}
	fmt.Printf("Publishing to queue %q\n", rm.Queue)
	res, err := publisher.Publish(ctx, &amqp.Message{
		Data: [][]byte{data},
	})
	if err != nil {
		log.Printf("Failed to publish a message: %v", err)
		return err
	}
	switch res.Outcome.(type) {
	case *rmq.StateAccepted:
	default:
		log.Printf("Unexpected publish outcome: %v", res.Outcome)
	}
	return nil
}

func (rm *RabbitMQAdapter) ConsumeMessage(ctx context.Context) (domain.BuildJob, error) {
	consumer, err := rm.rmqClient.NewConsumer(ctx, rm.Queue, &rmq.ConsumerOptions{InitialCredits: 1})
	if err != nil {
		log.Printf("Failed to create consumer: %v", err)
		return domain.BuildJob{}, err
	}
	defer consumer.Close(ctx)
	delivery, err := consumer.Receive(ctx)
	if err != nil {
		log.Printf("Failed to receive a message: %v", err)
		return domain.BuildJob{}, err
	}
	msg := delivery.Message()
	var job domain.BuildJob
	if err = json.Unmarshal(msg.GetData(), &job); err != nil {
		log.Printf("Failed to accept message: %v", err)
		return domain.BuildJob{}, err
	}
	err = delivery.Accept(ctx)
	if err != nil {
		log.Printf("Failed to accept message: %v", err)
		return domain.BuildJob{}, err
	}
	return job, nil
}
