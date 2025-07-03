package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/syz51/media-validation-service/internal/model"
)

// EventProcessor handles processing of SQS events containing S3 events
type EventProcessor struct {
	logger *log.Logger
}

// NewEventProcessor creates a new EventProcessor instance
func NewEventProcessor() *EventProcessor {
	return &EventProcessor{
		logger: log.New(log.Writer(), "[EventProcessor] ", log.LstdFlags|log.Lshortfile),
	}
}

// ProcessSQSEvent processes an SQS event containing S3 events
func (ep *EventProcessor) ProcessSQSEvent(ctx context.Context, sqsEvent events.SQSEvent) ([]model.ProcessedEvent, error) {
	ep.logger.Printf("Processing SQS event with %d records", len(sqsEvent.Records))

	var processedEvents []model.ProcessedEvent

	for i, record := range sqsEvent.Records {
		ep.logger.Printf("Processing SQS record %d/%d - MessageID: %s", i+1, len(sqsEvent.Records), record.MessageId)

		processedEvent := model.ProcessedEvent{
			MessageID:     record.MessageId,
			ReceiptHandle: record.ReceiptHandle,
			ProcessedAt:   time.Now(),
			SourceQueue:   record.EventSourceARN,
			Region:        record.AWSRegion,
		}

		// Log the raw SQS message details
		ep.logSQSRecord(record)

		// Attempt to parse S3 event from SQS message body
		s3Event, err := model.ParseS3EventFromSQSMessage(record)
		if err != nil {
			ep.logger.Printf("Failed to parse S3 event from SQS message %s: %v", record.MessageId, err)
			ep.logger.Printf("Raw message body: %s", record.Body)
			processedEvent.S3EventError = fmt.Sprintf("Failed to parse S3 event: %v", err)
		} else {
			ep.logger.Printf("Successfully parsed S3 event from SQS message %s", record.MessageId)
			processedEvent.S3Event = s3Event
			ep.logS3Event(*s3Event)
		}

		processedEvents = append(processedEvents, processedEvent)
	}

	ep.logger.Printf("Completed processing %d SQS records", len(processedEvents))
	return processedEvents, nil
}

// logSQSRecord logs details about an SQS record
func (ep *EventProcessor) logSQSRecord(record events.SQSMessage) {
	ep.logger.Printf("=== SQS Record Details ===")
	ep.logger.Printf("Message ID: %s", record.MessageId)
	ep.logger.Printf("Receipt Handle: %s", record.ReceiptHandle)
	ep.logger.Printf("Event Source: %s", record.EventSource)
	ep.logger.Printf("Event Source ARN: %s", record.EventSourceARN)
	ep.logger.Printf("AWS Region: %s", record.AWSRegion)
	ep.logger.Printf("Message Body Length: %d", len(record.Body))

	// Log attributes if present
	if len(record.Attributes) > 0 {
		ep.logger.Printf("Attributes:")
		for key, value := range record.Attributes {
			ep.logger.Printf("  %s: %s", key, value)
		}
	}

	// Log message attributes if present
	if len(record.MessageAttributes) > 0 {
		ep.logger.Printf("Message Attributes:")
		for key, attr := range record.MessageAttributes {
			stringValue := "nil"
			if attr.StringValue != nil {
				stringValue = *attr.StringValue
			}
			ep.logger.Printf("  %s: %s (type: %s)", key, stringValue, attr.DataType)
		}
	}
}

// logS3Event logs details about an S3 event
func (ep *EventProcessor) logS3Event(s3Event events.S3Event) {
	ep.logger.Printf("=== S3 Event Details ===")
	ep.logger.Printf("Number of S3 records: %d", len(s3Event.Records))

	for i, record := range s3Event.Records {
		ep.logger.Printf("--- S3 Record %d/%d ---", i+1, len(s3Event.Records))
		ep.logger.Printf("Event Version: %s", record.EventVersion)
		ep.logger.Printf("Event Source: %s", record.EventSource)
		ep.logger.Printf("AWS Region: %s", record.AWSRegion)
		ep.logger.Printf("Event Time: %s", record.EventTime)
		ep.logger.Printf("Event Name: %s", record.EventName)

		// Log user identity if available
		// Note: UserIdentity might not be available in all S3 event structures

		// Log request parameters
		if record.RequestParameters.SourceIPAddress != "" {
			ep.logger.Printf("Source IP: %s", record.RequestParameters.SourceIPAddress)
		}

		// Log S3 bucket and object details
		ep.logger.Printf("S3 Bucket Name: %s", record.S3.Bucket.Name)
		ep.logger.Printf("S3 Bucket ARN: %s", record.S3.Bucket.Arn)
		ep.logger.Printf("S3 Object Key: %s", record.S3.Object.Key)
		ep.logger.Printf("S3 Object Size: %d bytes", record.S3.Object.Size)
		ep.logger.Printf("S3 Object ETag: %s", record.S3.Object.ETag)
		ep.logger.Printf("S3 Object Sequencer: %s", record.S3.Object.Sequencer)

		// Log as JSON for complete details
		recordJSON, err := json.MarshalIndent(record, "", "  ")
		if err == nil {
			ep.logger.Printf("Complete S3 Record JSON:\n%s", string(recordJSON))
		}
	}
}

// LogProcessedEvents logs a summary of all processed events
func (ep *EventProcessor) LogProcessedEvents(processedEvents []model.ProcessedEvent) {
	ep.logger.Printf("=== Processing Summary ===")
	ep.logger.Printf("Total events processed: %d", len(processedEvents))

	successCount := 0
	errorCount := 0

	for _, event := range processedEvents {
		if event.S3EventError == "" {
			successCount++
		} else {
			errorCount++
		}
	}

	ep.logger.Printf("Successfully parsed S3 events: %d", successCount)
	ep.logger.Printf("Failed to parse S3 events: %d", errorCount)

	// Log complete processed events as JSON
	if processedJSON, err := json.MarshalIndent(processedEvents, "", "  "); err == nil {
		ep.logger.Printf("Complete processed events JSON:\n%s", string(processedJSON))
	}
}
