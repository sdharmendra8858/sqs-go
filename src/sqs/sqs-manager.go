package sqs

import (
	"log"
	"sync"
)

var (
	handlerInstance *SQSHandler
	once            sync.Once
)

// Initialize initializes the SQSHandler singleton.
func Initialize(queueURLs map[string]string) {
	once.Do(func() {
		var err error
		handlerInstance, err = NewSQSHandler(queueURLs)
		if err != nil {
			log.Fatalf("Failed to initialize SQSHandler: %v", err)
		}
		log.Println("SQSHandler initialized.")
	})
}

// GetHandler returns the singleton instance of SQSHandler.
func GetHandler() *SQSHandler {
	if handlerInstance == nil {
		log.Fatal("SQSHandler not initialized. Call sqsmanager.Initialize first.")
	}
	return handlerInstance
}
