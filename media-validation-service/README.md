# Media Validation Service

This service processes SQS events containing S3 events for media validation workflows.

## Architecture

The service receives SQS events that contain S3 events in their message body. When an S3 object is created, it triggers an S3 event notification that gets sent to an SQS queue, which then triggers this Lambda function via AWS Lambda Web Adapter.

## Event Flow

1. S3 object uploaded/created → S3 Event Notification
2. S3 Event sent to SQS queue 
3. SQS event (containing S3 event in body) sent to Lambda function
4. Lambda function processes the event via this service

## Event Processing

The service parses incoming SQS events and extracts S3 events from the message body. It logs detailed information about:

- SQS message details (message ID, receipt handle, source queue, etc.)
- S3 event details (bucket name, object key, size, event type, etc.)
- Processing results and any errors

## API Endpoints

### POST /events

Receives and processes SQS events containing S3 events.

**Request Body**: SQS Event JSON (as shown in `sample_sqs_s3_event.json`)

**Response**:
```json
{
  "status": "success",
  "message": "SQS events processed successfully", 
  "processedEvents": 1,
  "successfulEvents": 1,
  "failedEvents": 0
}
```

## Testing

You can test the service locally by sending a POST request to `/events` with the sample event:

```bash
curl -X POST http://localhost:8080/events \
  -H "Content-Type: application/json" \
  -d @sample_sqs_s3_event.json
```

## Implementation Details

### Key Components

- **EventProcessor**: Parses SQS events and extracts S3 events from message bodies
- **ProcessedEvent**: Model representing a processed SQS message with parsed S3 event
- **Handler**: HTTP handler that receives events and orchestrates processing

### Event Types Supported

- Standard SQS queues
- S3 Object Created events (ObjectCreated:Put, ObjectCreated:Post, etc.)
- S3 Object Deleted events (ObjectRemoved:Delete, ObjectRemoved:DeleteMarkerCreated, etc.)

### Logging

The service provides comprehensive logging including:
- Raw SQS message details
- Parsed S3 event information
- Processing statistics
- Complete JSON dumps of events for debugging

## Configuration

The service uses the existing configuration system with server settings in `configs/config.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 8080
```

## Deployment

The service is deployed as a Lambda function using AWS Lambda Web Adapter. See `template.yml` for CloudFormation configuration. 