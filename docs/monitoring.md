# Monitoring and Metrics

This document describes the monitoring capabilities and metrics exposed by the Go-SMS-Gateway service.

## Prometheus Metrics

The SMS Gateway exposes Prometheus metrics for real-time monitoring of system performance and health. These metrics can be scraped by Prometheus and visualized using tools like Grafana.

### Metrics Endpoint

By default, metrics are exposed at:

```
http://localhost:9090/metrics
```

The port can be configured in the configuration file.

## Key Metrics

### Message Throughput

| Metric | Type | Description |
|--------|------|-------------|
| `sms_gateway_messages_submitted_total` | Counter | Total number of messages submitted to providers |
| `sms_gateway_messages_submitted_rate` | Gauge | Current rate of message submissions per second |
| `sms_gateway_messages_received_total` | Counter | Total number of messages received from providers |
| `sms_gateway_messages_received_rate` | Gauge | Current rate of message reception per second |

### Delivery Status

| Metric | Type | Description |
|--------|------|-------------|
| `sms_gateway_delivery_success_total` | Counter | Total number of successfully delivered messages |
| `sms_gateway_delivery_failure_total` | Counter | Total number of failed message deliveries |
| `sms_gateway_delivery_pending_total` | Gauge | Current number of messages pending delivery |
| `sms_gateway_delivery_success_rate` | Gauge | Success rate as a percentage of total deliveries |

### Provider Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `sms_gateway_provider_connection_status` | Gauge | Connection status per provider (1=connected, 0=disconnected) |
| `sms_gateway_provider_messages_submitted_total` | Counter | Messages submitted per provider |
| `sms_gateway_provider_delivery_success_rate` | Gauge | Success rate per provider |
| `sms_gateway_provider_latency_seconds` | Histogram | Message delivery latency per provider |

### Queue Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `sms_gateway_queue_depth` | Gauge | Current number of messages in each queue |
| `sms_gateway_queue_processing_time_seconds` | Histogram | Time taken to process messages from queue |
| `sms_gateway_queue_wait_time_seconds` | Histogram | Time messages spend waiting in queue |

### System Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `sms_gateway_uptime_seconds` | Counter | Time since service start |
| `sms_gateway_goroutines` | Gauge | Number of active goroutines |
| `sms_gateway_memory_usage_bytes` | Gauge | Memory usage of the service |

## Alerting

The following metrics are recommended for alerting:

- `sms_gateway_provider_connection_status` - Alert when providers disconnect
- `sms_gateway_delivery_success_rate` - Alert when success rate drops below threshold
- `sms_gateway_queue_depth` - Alert when queues grow beyond normal levels
- `sms_gateway_messages_submitted_rate` - Alert on unusual traffic patterns

## Grafana Dashboard

A sample Grafana dashboard is provided in the `monitoring` directory. To import it:

1. Access your Grafana instance
2. Navigate to Dashboards > Import
3. Upload the JSON file or paste its contents
4. Configure the Prometheus data source

## Logging

In addition to metrics, the SMS Gateway provides detailed logging:

- Logs are written to `logs/sms-gateway.log` by default
- Log level can be configured in the configuration file
- Structured JSON logging format for easy parsing
- Log rotation to prevent disk space issues

## Health Check Endpoint

A health check endpoint is available at:

```
http://localhost:8080/health
```

This endpoint returns:
- HTTP 200 OK when the service is healthy
- HTTP 503 Service Unavailable when there are issues

The health check verifies:
- Database connectivity
- Redis connectivity
- Provider connection status
- Message processing status

## Operational Monitoring

For operational monitoring, consider tracking:

1. **System Resources**:
   - CPU usage
   - Memory usage
   - Disk I/O
   - Network I/O

2. **External Dependencies**:
   - Redis latency and errors
   - Database query performance
   - SMPP provider response times

3. **Business Metrics**:
   - Messages per second
   - Revenue-generating traffic
   - Cost per message
   - Provider distribution
