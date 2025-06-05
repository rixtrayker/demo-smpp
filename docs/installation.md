# Installation Guide

This guide will help you set up and run the Go-SMS-Gateway service.

## Prerequisites

Before installing the SMS Gateway, ensure you have the following prerequisites:

- Go 1.18 or higher
- Redis server (for message queuing)
- MySQL/MariaDB database (for message storage and tracking)
- Git (for cloning the repository)

## Step-by-Step Installation

### 1. Clone the Repository

```bash
git clone https://github.com/yourusername/go-sms-gateway.git
cd go-sms-gateway
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Database Setup

Create a new database for the SMS Gateway:

```sql
CREATE DATABASE sms_gateway;
```

The application will automatically create the necessary tables on first run.

### 4. Configure Redis

Ensure Redis is running and accessible. The default configuration expects Redis to be available at `127.0.0.1:6379`.

### 5. Configuration

Copy the example configuration file and modify it according to your needs:

```bash
cp example-config.yaml config.yaml
```

Edit `config.yaml` with your preferred text editor to configure your SMPP providers and other settings.

### 6. Running the Service

Use the provided script to run the service:

```bash
# Start with default config.yaml
./run.sh

# Start with a specific config file
./run.sh custom-config.yaml
```

### 7. Verify Installation

Check the logs to ensure the service has started correctly:

```bash
tail -f logs/sms-gateway.log
```

You should see messages indicating successful connections to your configured SMPP providers.

## Docker Installation

You can also run the SMS Gateway using Docker:

```bash
# Build the Docker image
docker build -t go-sms-gateway .

# Run the container
docker run -p 8080:8080 -v $(pwd)/config.yaml:/app/config.yaml go-sms-gateway
```

## Troubleshooting

If you encounter issues during installation:

1. Check that all prerequisites are correctly installed
2. Verify your database connection settings
3. Ensure Redis is running and accessible
4. Check the logs for specific error messages
5. Verify your SMPP provider credentials and connection details

For more detailed information, refer to the [Configuration Guide](configuration.md).
