# 🚀 Go-SMS-Gateway

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/go-1.18+-00ADD8.svg)
![SMPP](https://img.shields.io/badge/protocol-SMPP-orange.svg)

## 📱 Overview

A high-performance SMS gateway built in Go that bridges telecom networks with modern applications. This service provides robust SMPP protocol support, efficient message queuing, and reliable delivery tracking.

## ✨ Features

- 🔄 **Multi-Provider Support**: Connect to multiple SMPP providers simultaneously
- 📊 **Prometheus Metrics**: Real-time monitoring of message throughput and delivery rates
- 🚦 **Rate Limiting**: Configurable rate limiting per provider to prevent throttling
- 🔁 **Message Queuing**: Redis-backed message queue for reliable message handling
- 📋 **Delivery Reports**: Track message delivery status with detailed reporting
- 🔌 **Flexible Session Types**: Support for transceiver, receiver, and transmitter sessions
- 🛡️ **Graceful Shutdown**: Clean handling of connection termination and message processing

## 🛠️ Architecture

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

For more details on the system architecture, see our [Architecture Documentation](docs/architecture.md).

## 🚀 Getting Started

### Prerequisites

- Go 1.18+
- Redis server
- MySQL/MariaDB database

For detailed installation instructions, see our [Installation Guide](docs/installation.md).

### Configuration

Create a `config.yaml` file with your SMPP provider details:

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

For detailed configuration options, see our [Configuration Documentation](docs/configuration.md).

### Running the Service

```bash
# Start the service with default config
./run.sh

# Start with a specific config file
./run.sh custom-config.yaml
```

## 📱 API

The SMS Gateway provides a RESTful API for message submission and status tracking. For detailed API documentation, see our [API Documentation](docs/api.md).

## 📊 Monitoring

The service exposes Prometheus metrics for monitoring:

- Message throughput rates
- Delivery success/failure rates
- Queue depths
- Provider connection status

For detailed information about available metrics and monitoring setup, see our [Monitoring Documentation](docs/monitoring.md).

## 🔍 Future Enhancements

See our [Future Enhancements](docs/future-enhancements.md) documentation for detailed information about planned improvements.

### Testing Improvements
- 🧪 Comprehensive unit test coverage
- 🔄 Integration tests with mock SMPP servers
- 🧮 Load testing framework for performance benchmarking
- 📝 Test profiles for complex scenarios including throttling and packet loss

### Performance Optimizations
- ⚡ Connection pooling for higher throughput
- 🔄 Bulk database operations for status updates
- 🚀 Optimized message encoding/decoding
- 📊 Enhanced metrics collection and visualization

### Reliability Features
- 🔄 Automatic session reconnection with exponential backoff
- 💾 Message persistence across restarts
- 🛡️ Enhanced error handling and recovery mechanisms
- 🔄 Background synchronization for delivery reports

### Operational Improvements
- 🔧 Admin dashboard for monitoring and configuration
- 📱 REST API for message submission and status queries
- 🔐 Enhanced security features and authentication
- 📋 Comprehensive logging and audit trails

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
