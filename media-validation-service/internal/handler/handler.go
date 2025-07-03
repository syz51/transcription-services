package handler

import (
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/labstack/echo/v4"

	"github.com/syz51/media-validation-service/internal/config"
	"github.com/syz51/media-validation-service/internal/service"
)

// Handler contains all the handlers
type Handler struct {
	config         *config.Config
	eventProcessor *service.EventProcessor
	logger         *log.Logger
}

// New creates a new handler instance
func New(cfg *config.Config) *Handler {
	return &Handler{
		config:         cfg,
		eventProcessor: service.NewEventProcessor(),
		logger:         log.New(log.Writer(), "[Handler] ", log.LstdFlags|log.Lshortfile),
	}
}

// EventsResponse represents the response structure for the events endpoint
type EventsResponse struct {
	Status           string `json:"status"`
	Message          string `json:"message"`
	ProcessedEvents  int    `json:"processedEvents"`
	SuccessfulEvents int    `json:"successfulEvents"`
	FailedEvents     int    `json:"failedEvents"`
}

// Events handles incoming SQS events containing S3 events
func (h *Handler) Events(c echo.Context) error {
	h.logger.Printf("Received request to /events endpoint")

	// Parse the request body as an SQS event
	var sqsEvent events.SQSEvent
	if err := c.Bind(&sqsEvent); err != nil {
		h.logger.Printf("Failed to parse SQS event from request body: %v", err)

		return c.JSON(http.StatusBadRequest, EventsResponse{
			Status:  "error",
			Message: "Failed to parse SQS event from request body",
		})
	}

	h.logger.Printf("Successfully parsed SQS event with %d records", len(sqsEvent.Records))

	// Process the SQS event
	processedEvents, err := h.eventProcessor.ProcessSQSEvent(c.Request().Context(), sqsEvent)
	if err != nil {
		h.logger.Printf("Failed to process SQS event: %v", err)
		return c.JSON(http.StatusInternalServerError, EventsResponse{
			Status:  "error",
			Message: "Failed to process SQS event",
		})
	}

	// Log the processed events summary
	h.eventProcessor.LogProcessedEvents(processedEvents)

	// Count successful and failed events
	successfulEvents := 0
	failedEvents := 0
	for _, event := range processedEvents {
		if event.S3EventError == "" {
			successfulEvents++
		} else {
			failedEvents++
		}
	}

	response := EventsResponse{
		Status:           "success",
		Message:          "SQS events processed successfully",
		ProcessedEvents:  len(processedEvents),
		SuccessfulEvents: successfulEvents,
		FailedEvents:     failedEvents,
	}

	h.logger.Printf("Processing complete: %d total, %d successful, %d failed",
		len(processedEvents), successfulEvents, failedEvents)

	return c.JSON(http.StatusOK, response)
}
