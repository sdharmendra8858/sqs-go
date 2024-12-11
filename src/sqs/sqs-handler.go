package sqs

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type SQSHandler struct {
	client    *sqs.Client
	queues    map[string]string // Queue names mapped to their URLs
	ctx       context.Context
	cancelCtx context.CancelFunc
}

// NewSQSHandler initializes the SQS client for LocalStack.
func NewSQSHandler(queueURLs map[string]string) (*SQSHandler, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("ap-south-1"), // Default region for LocalStack
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")), // Dummy credentials
		config.WithEndpointResolver(aws.EndpointResolverFunc(
			func(service, region string) (aws.Endpoint, error) {
				return aws.Endpoint{URL: "http://localhost:4566"}, nil // LocalStack endpoint
			}),
		),
	)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	client := sqs.NewFromConfig(cfg)
	ctx, cancel := context.WithCancel(context.Background())

	return &SQSHandler{
		client:    client,
		queues:    queueURLs,
		ctx:       ctx,
		cancelCtx: cancel,
	}, nil
}

// PublishMessage sends a message to the specified SQS queue.
func (h *SQSHandler) PublishMessage(queueName, messageBody string, delaySeconds int32) error {
	queueURL, ok := h.queues[queueName]
	if !ok {
		return fmt.Errorf("queue %s not found", queueName)
	}

	_, err := h.client.SendMessage(h.ctx, &sqs.SendMessageInput{
		QueueUrl:     &queueURL,
		MessageBody:  &messageBody,
		DelaySeconds: delaySeconds,
	})
	if err != nil {
		return err
	}
	log.Printf("Message published to %s: %s\n", queueName, messageBody)
	return nil
}

// StartConsumer starts consuming messages from a specific SQS queue.
func (h *SQSHandler) StartConsumer(queueName string, workerFunc func(message string) error, visibilityTimeout, waitTimeSeconds int32) {
	queueURL, ok := h.queues[queueName]
	if !ok {
		log.Printf("Queue %s not found\n", queueName)
		return
	}

	go func() {
		for {
			select {
			case <-h.ctx.Done():
				log.Printf("Consumer for queue %s stopped.\n", queueName)
				return
			default:
				h.consumeMessages(queueName, queueURL, workerFunc, visibilityTimeout, waitTimeSeconds)
			}
		}
	}()
}

// consumeMessages fetches and processes messages from a specific queue.
func (h *SQSHandler) consumeMessages(queueName, queueURL string, workerFunc func(message string) error, visibilityTimeout, waitTimeSeconds int32) {
	output, err := h.client.ReceiveMessage(h.ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            &queueURL,
		MaxNumberOfMessages: 10,
		VisibilityTimeout:   visibilityTimeout,
		WaitTimeSeconds:     waitTimeSeconds,
	})
	if err != nil {
		log.Printf("Error receiving messages from queue %s: %v\n", queueName, err)
		return
	}

	for _, message := range output.Messages {
		log.Printf("Received message from %s: %s\n", queueName, *message.Body)
		if err := workerFunc(*message.Body); err != nil {
			log.Printf("Error processing message from %s: %v\n", queueName, err)
			continue
		}

		_, delErr := h.client.DeleteMessage(h.ctx, &sqs.DeleteMessageInput{
			QueueUrl:      &queueURL,
			ReceiptHandle: message.ReceiptHandle,
		})
		if delErr != nil {
			log.Printf("Error deleting message from %s: %v\n", queueName, delErr)
		}
	}
}

// GracefulShutdown ensures proper shutdown of the consumer.
func (h *SQSHandler) GracefulShutdown() {
	h.cancelCtx()
	log.Println("Shutting down...")
}
