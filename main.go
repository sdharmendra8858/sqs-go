package main

import (
	"log"
	"os"
	"os/signal"
	HlthCntrl "sqs/src/module/health"
	SqsCntrl "sqs/src/module/sqs"
	"sqs/src/sqs"
	"syscall"

	"github.com/gin-gonic/gin"
)

func main() {

	queueURLs := map[string]string{
		"queue1": "http://localhost:4566/000000000000/queue1",
		"queue2": "http://localhost:4566/000000000000/queue2",
	}

	// Initialize the SQS handler singleton
	sqs.Initialize(queueURLs)

	// Get mode from environment variable
	mode := os.Getenv("MODE")
	if mode == "" {
		mode = "server" // Default mode
	}

	if mode == "server" {
		startServer()
	} else if mode == "consumer" {
		startConsumers()
	} else {
		log.Fatalf("Invalid MODE: %s. Use 'server' or 'consumer'.", mode)
	}
}

func startServer() {
	// Create a new Gin router
	router := gin.Default()

	// Define routes
	router.GET("/health", HlthCntrl.Health)
	router.POST("/sqs/produce", SqsCntrl.ProduceMessage)

	// Graceful shutdown in a separate goroutine
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		log.Println("Application shutting down.")
		os.Exit(0)
	}()

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}
	log.Printf("Starting server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %s", err)
	}
}

func startConsumers() {
	log.Println("Starting consumers...")

	handler := sqs.GetHandler()

	// Define worker functions for each queue
	workerQueue1 := func(message string) error {
		log.Printf("Processing message from queue1: %s\n", message)
		// Add your custom logic here
		return nil
	}

	workerQueue2 := func(message string) error {
		log.Printf("Processing message from queue2: %s\n", message)
		// Add your custom logic here
		return nil
	}

	// Start consumers for each queue
	go handler.StartConsumer("queue1", workerQueue1, 30, 10)
	go handler.StartConsumer("queue2", workerQueue2, 30, 10)

	// Wait for shutdown signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("Consumers shutting down.")
}
