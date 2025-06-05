# API Documentation

This document describes the API endpoints provided by the Go-SMS-Gateway service for message submission and status tracking.

## API Overview

The SMS Gateway provides a RESTful API for:
- Submitting SMS messages
- Checking message delivery status
- Managing provider connections
- Retrieving system statistics

## Base URL

All API endpoints are relative to the base URL:

```
http://localhost:8080/api/v1
```

The port and base path can be configured in the configuration file.

## Authentication

API requests require authentication using an API key. The key should be included in the `X-API-Key` header:

```
X-API-Key: your-api-key-here
```

API keys can be managed through the administration interface or configuration file.

## Message Submission

### Submit a Single Message

**Endpoint:** `POST /messages`

**Request Body:**

```json
{
  "destination": "1234567890",
  "source": "SENDER",
  "message": "Hello, this is a test message",
  "provider": "provider1",
  "priority": 1,
  "validity_period": 86400,
  "callback_url": "https://your-callback-url.com/delivery-status"
}
```

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `destination` | string | Yes | Recipient phone number |
| `source` | string | Yes | Sender ID or phone number |
| `message` | string | Yes | Message content |
| `provider` | string | No | Specific provider to use (optional) |
| `priority` | integer | No | Message priority (1-5, default: 3) |
| `validity_period` | integer | No | Validity period in seconds (default: 86400) |
| `callback_url` | string | No | URL for delivery status callbacks |

**Response:**

```json
{
  "message_id": "msg-123456789",
  "status": "queued",
  "provider": "provider1",
  "submitted_at": "2023-06-05T12:34:56Z"
}
```

### Submit Multiple Messages

**Endpoint:** `POST /messages/batch`

**Request Body:**

```json
{
  "messages": [
    {
      "destination": "1234567890",
      "source": "SENDER",
      "message": "Hello, this is message 1"
    },
    {
      "destination": "0987654321",
      "source": "SENDER",
      "message": "Hello, this is message 2"
    }
  ],
  "provider": "provider1",
  "priority": 2,
  "callback_url": "https://your-callback-url.com/delivery-status"
}
```

**Response:**

```json
{
  "batch_id": "batch-123456789",
  "message_count": 2,
  "successful": 2,
  "failed": 0,
  "messages": [
    {
      "message_id": "msg-123456789",
      "status": "queued"
    },
    {
      "message_id": "msg-987654321",
      "status": "queued"
    }
  ]
}
```

## Message Status

### Check Message Status

**Endpoint:** `GET /messages/{message_id}`

**Response:**

```json
{
  "message_id": "msg-123456789",
  "destination": "1234567890",
  "source": "SENDER",
  "status": "delivered",
  "provider": "provider1",
  "submitted_at": "2023-06-05T12:34:56Z",
  "delivered_at": "2023-06-05T12:35:10Z",
  "delivery_code": "0",
  "delivery_description": "Delivered to handset"
}
```

### Check Batch Status

**Endpoint:** `GET /messages/batch/{batch_id}`

**Response:**

```json
{
  "batch_id": "batch-123456789",
  "message_count": 2,
  "status_summary": {
    "delivered": 1,
    "failed": 0,
    "pending": 1
  },
  "messages": [
    {
      "message_id": "msg-123456789",
      "status": "delivered"
    },
    {
      "message_id": "msg-987654321",
      "status": "pending"
    }
  ]
}
```

## Provider Management

### List Providers

**Endpoint:** `GET /providers`

**Response:**

```json
{
  "providers": [
    {
      "name": "provider1",
      "status": "connected",
      "connected_since": "2023-06-05T10:00:00Z",
      "message_count": 1250,
      "success_rate": 98.5
    },
    {
      "name": "provider2",
      "status": "disconnected",
      "last_connected": "2023-06-05T09:45:00Z",
      "message_count": 750,
      "success_rate": 97.2
    }
  ]
}
```

### Provider Status

**Endpoint:** `GET /providers/{provider_name}`

**Response:**

```json
{
  "name": "provider1",
  "status": "connected",
  "connected_since": "2023-06-05T10:00:00Z",
  "message_count": 1250,
  "success_rate": 98.5,
  "current_tps": 42.5,
  "queue_depth": 15
}
```

## System Status

### Get System Status

**Endpoint:** `GET /status`

**Response:**

```json
{
  "status": "healthy",
  "uptime": 86400,
  "version": "1.2.3",
  "message_stats": {
    "total_submitted": 10000,
    "total_delivered": 9850,
    "total_failed": 150,
    "current_tps": 45.2
  },
  "providers": {
    "connected": 3,
    "disconnected": 1
  },
  "queue_depth": 25
}
```

## Webhook Callbacks

When a `callback_url` is provided during message submission, the SMS Gateway will send delivery status updates to that URL.

**Callback Payload:**

```json
{
  "message_id": "msg-123456789",
  "status": "delivered",
  "destination": "1234567890",
  "delivered_at": "2023-06-05T12:35:10Z",
  "delivery_code": "0",
  "delivery_description": "Delivered to handset"
}
```

## Error Responses

All API errors follow a standard format:

```json
{
  "error": {
    "code": "invalid_request",
    "message": "Invalid destination number format",
    "details": {
      "field": "destination",
      "reason": "Must be a valid E.164 format phone number"
    }
  }
}
```

## Rate Limiting

API endpoints are rate-limited. The current limits are included in response headers:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1623456789
```

When rate limits are exceeded, the API returns HTTP 429 Too Many Requests.
