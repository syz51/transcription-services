package model

import (
	"encoding/json"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

// ProcessedEvent represents a processed SQS message containing an S3 event
type ProcessedEvent struct {
	MessageID     string          `json:"messageId"`
	ReceiptHandle string          `json:"receiptHandle"`
	S3Event       *events.S3Event `json:"s3Event,omitempty"`
	S3EventError  string          `json:"s3EventError,omitempty"`
	ProcessedAt   time.Time       `json:"processedAt"`
	SourceQueue   string          `json:"sourceQueue"`
	Region        string          `json:"region"`
}

// ParseS3EventFromSQSMessage attempts to parse an S3 event from an SQS message body
func ParseS3EventFromSQSMessage(message events.SQSMessage) (*events.S3Event, error) {
	var s3Event events.S3Event
	err := json.Unmarshal([]byte(message.Body), &s3Event)
	if err != nil {
		return nil, err
	}
	return &s3Event, nil
}
