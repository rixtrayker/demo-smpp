# Architecture Overview

This document provides a detailed overview of the Go-SMS-Gateway architecture and design principles.

## System Architecture

The SMS Gateway is designed as a high-performance, scalable service that bridges telecom networks with modern applications. The system follows a modular architecture with clear separation of concerns.

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Client    │     │  SMS Gateway │     │    SMSC     │
│  Application│────▶│    Service   │────▶│  Providers  │
└─────────────┘     └─────────────┘     └─────────────┘
                          │
                    ┌─────┴─────┐
                    │           │
               ┌────▼───┐  ┌────▼───┐
               │ Redis  │  │Database│
               │ Queue  │  │        │
               └────────┘  └────────┘
```

## Core Components

### 1. Provider Manager

The Provider Manager is responsible for:
- Establishing and maintaining SMPP connections to telecom providers
- Managing connection lifecycles and reconnection strategies
- Routing messages to appropriate providers based on rules and availability

### 2. Message Queue

The Redis-backed message queue system:
- Provides reliable message storage before delivery
- Enables asynchronous processing and decoupling
- Supports priority-based message handling
- Ensures message persistence across service restarts

### 3. Message Processor

The Message Processor:
- Pulls messages from Redis queues
- Applies rate limiting and throttling
- Handles message encoding and protocol conversion
- Submits messages to SMPP providers
- Processes delivery reports

### 4. Database Layer

The database layer:
- Stores message metadata and delivery status
- Tracks provider performance metrics
- Maintains message history for reporting
- Provides data for analytics and monitoring

### 5. Monitoring System

The monitoring system:
- Exposes Prometheus metrics for real-time monitoring
- Tracks message throughput and delivery rates
- Monitors provider connection status
- Provides alerts for system issues

## Message Flow

### Outbound Message Flow

1. Client applications submit messages to the gateway
2. Messages are validated and stored in Redis queues
3. Message Processor pulls messages from queues based on priority
4. Rate limiting is applied according to provider configuration
5. Messages are encoded and submitted to the appropriate SMPP provider
6. Delivery reports are processed and status is updated in the database

### Inbound Message Flow

1. SMPP providers deliver incoming messages to the gateway
2. Messages are decoded and validated
3. Messages are stored in the database
4. Notifications are sent to client applications (if configured)

## Connection Management

The SMS Gateway implements sophisticated connection management:

- **Connection Pooling**: Maintains multiple connections to each provider for higher throughput
- **Automatic Reconnection**: Handles connection failures with exponential backoff
- **Load Balancing**: Distributes messages across available connections
- **Session Monitoring**: Regularly checks connection health with enquire_link operations

## Rate Limiting

The system implements multi-level rate limiting:

- **Global Rate Limiting**: Caps the total message throughput across all providers
- **Provider Rate Limiting**: Enforces provider-specific rate limits
- **Burst Handling**: Allows temporary bursts while maintaining average limits
- **Adaptive Throttling**: Adjusts sending rates based on provider responses

## Error Handling

The SMS Gateway implements comprehensive error handling:

- **Retry Logic**: Automatically retries failed message submissions
- **Error Classification**: Categorizes errors as temporary or permanent
- **Fallback Routing**: Routes messages to alternative providers when primary providers fail
- **Dead Letter Queues**: Captures messages that cannot be delivered after retries

## Scalability

The architecture is designed for horizontal scalability:

- **Stateless Design**: Allows running multiple gateway instances
- **Shared Redis**: Enables coordination between multiple instances
- **Database Connection Pooling**: Efficiently manages database connections
- **Provider Connection Distribution**: Distributes provider connections across instances
