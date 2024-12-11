package sqs

import (
	"log"
	"sqs/src/sqs"

	"github.com/gin-gonic/gin"
)

func ProduceMessage(c *gin.Context) {
	var message struct {
		QueueName string `json:"queueName" binding:"required"`
		Body      string `json:"body" binding:"required"`
	}

	if err := c.ShouldBindJSON(&message); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	handler := sqs.GetHandler()
	if err := handler.PublishMessage(message.QueueName, message.Body, 0); err != nil {
		log.Printf("Failed to publish message: %v\n", err)
		c.JSON(500, gin.H{"error": "Failed to publish message"})
		return
	}

	c.JSON(200, gin.H{"message": "Message published successfully"})
}
