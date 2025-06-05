# Configuration Guide

This document provides detailed information on configuring the Go-SMS-Gateway service.

## Configuration File Format

The SMS Gateway uses a YAML configuration file. By default, it looks for `config.yaml` in the root directory, but you can specify a different file when starting the service.

## Basic Configuration

Here's a basic configuration example:

```yaml
rate_limit: 100
redis_url: "127.0.0.1:6379"

database_config:
  host: "localhost"
  port: 3306
  user: "root"
  password: "your_password"
  dbname: "sms_gateway"
  max_conn: 50
  max_idle: 25

providers:
  - provider1:
    name: "provider1"
    session_type: "transceiver"
    address: "smpp.provider1.com"
    port: 2775
    system_id: "your_id"
    password: "your_password"
    system_type: ""
    rate_limit: 500
    burst_limit: 500
    max_outstanding: 500
    has_outstanding: true
    max_retries: 3
    queues:
      - "go-provider1-ported"
      - "go-provider1"
```

## Configuration Sections

### Global Settings

| Parameter | Description | Default |
|-----------|-------------|---------|
| `rate_limit` | Global rate limit for all providers combined (messages per second) | 100 |
| `redis_url` | Connection URL for Redis server | "127.0.0.1:6379" |

### Database Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `host` | Database server hostname | "localhost" |
| `port` | Database server port | 3306 |
| `user` | Database username | "root" |
| `password` | Database password | - |
| `dbname` | Database name | "sms_gateway" |
| `max_conn` | Maximum number of database connections | 50 |
| `max_idle` | Maximum number of idle connections | 25 |

### Provider Configuration

Each provider requires the following configuration:

| Parameter | Description | Required |
|-----------|-------------|----------|
| `name` | Unique provider identifier | Yes |
| `session_type` | SMPP session type: "transceiver", "transmitter", or "receiver" | Yes |
| `address` | SMPP server hostname or IP address | Yes |
| `port` | SMPP server port | Yes |
| `system_id` | SMPP account username | Yes |
| `password` | SMPP account password | Yes |
| `system_type` | SMPP system type (provider-specific) | No |
| `rate_limit` | Provider-specific rate limit (messages per second) | Yes |
| `burst_limit` | Maximum burst size for rate limiting | Yes |
| `max_outstanding` | Maximum number of messages awaiting delivery reports | Yes |
| `has_outstanding` | Whether to track outstanding messages | Yes |
| `max_retries` | Maximum number of retry attempts for failed messages | Yes |
| `queues` | List of Redis queue names to pull messages from | Yes |

## Session Types

The SMS Gateway supports three SMPP session types:

1. **Transceiver**: Both sends and receives messages over a single connection
2. **Transmitter**: Only sends messages
3. **Receiver**: Only receives messages (typically delivery reports)

## Queue Configuration

Each provider can be configured to pull messages from multiple Redis queues. Queues are processed in the order they are listed, allowing for priority-based message handling.

## Advanced Configuration

### TLS Configuration

To enable secure SMPP connections:

```yaml
providers:
  - provider1:
    # ... other settings ...
    use_tls: true
    tls_verify: true
    tls_client_cert: "/path/to/client.crt"
    tls_client_key: "/path/to/client.key"
    tls_ca_cert: "/path/to/ca.crt"
```

### Connection Settings

Fine-tune connection parameters:

```yaml
providers:
  - provider1:
    # ... other settings ...
    conn_timeout: 10s
    read_timeout: 30s
    write_timeout: 30s
    enquire_link_interval: 15s
```

## Environment Variable Substitution

The configuration file supports environment variable substitution using the `${ENV_VAR}` syntax:

```yaml
database_config:
  password: "${DB_PASSWORD}"
```

## Multiple Provider Example

```yaml
providers:
  - provider1:
    name: "provider1"
    # ... provider1 settings ...
    
  - provider2:
    name: "provider2"
    # ... provider2 settings ...
```

For more examples, refer to the `example-config.yaml` file in the repository.
