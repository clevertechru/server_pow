[![Go](https://github.com/clevertechru/server_pow/actions/workflows/go.yml/badge.svg)](https://github.com/clevertechru/server_pow/actions/workflows/go.yml)
[![Build Status](https://app.travis-ci.com/clevertechru/server_pow.svg?token=bbzT95wZRUs7cXAPJccG&branch=main)](https://app.travis-ci.com/clevertechru/server_pow)
# TCP server protected from DDOS attacks with the Proof of Work
## Features
* Proof of Work challenge using random quotes
* Configurable difficulty levels
* Rate limiting and connection limiting
* Worker pool for controlled concurrency
* Proper connection handling
* Timeouts to prevent hanging connections
* Nonce tracking with window expiration
* Graceful shutdown
* Comprehensive test coverage
* Docker support
* Prometheus metrics for monitoring
## PoW Algorithm Choice:
* Using SHA256 with a nonce-based challenge
* The server generates a random quote and a target (currently set to 2)
* The client must find a nonce that, when combined with the quote and timestamp, produces a hash starting with the target
* This is computationally intensive but verifiable quickly
* The difficulty can be adjusted by changing the number of zeros in the target

# Project Structure
```
.
├── cmd/                               # Main application entry points
│   ├── client/                        # Client application
│   └── server/                        # Server application
├── internal/                          # Private application code
│   ├── client/                        # Client implementation
│   └── server/                        # Server implementation
│       ├── service/                   # Server services
│       │   ├── connection_manager.go  # Connection handling and timeouts
│       │   ├── pow_service.go         # Proof of Work service
│       │   └── quote_service.go       # Quote service
│       └── handler.go                 # Main server handler
├── pkg/                               # Public libraries
│   ├── backoff/                       # Backoff queue implementation
│   ├── config/                        # Configuration handling
│   ├── connlimit/                     # Connection limiting
│   ├── metrics/                       # Prometheus metrics
│   ├── nonce/                         # Nonce tracking
│   ├── pow/                           # Proof of Work implementation
│   ├── quotes/                        # Quotes storage and retrieval
│   ├── ratelimit/                     # Rate limiting
│   └── workerpool/                    # Worker pool implementation
├── config/                            # Configuration files
├── vendor/                            # Dependencies
├── docker-compose.yml                 # Docker Compose configuration
├── client.Dockerfile                  # Client Dockerfile
└── server.Dockerfile                  # Server Dockerfile
```

# Metrics
The server exposes the following Prometheus metrics:

## Connection Metrics
* `server_active_connections` - Number of active connections
* `server_total_connections` - Total number of connections

## PoW Metrics
* `server_pow_challenges_generated` - Number of PoW challenges generated
* `server_pow_challenges_verified` - Number of PoW challenges verified
* `server_pow_verification_failures` - Number of PoW verification failures

## Rate Limiting Metrics
* `server_rate_limit_hits` - Number of rate limit hits

## Worker Pool Metrics
* `server_worker_pool_size` - Current size of the worker pool
* `server_worker_pool_queue_size` - Current size of the worker pool queue

## Response Time Metrics
* `server_response_time_seconds` - Response time in seconds (histogram)

# How to RUN
```
git clone https://github.com/clevertechru/server_pow.git
cd server_pow
docker compose up --build
```

# How to Run tests and linter
```
go test ./...
golangci-lint run ./...
```

# Server environment settings
```
# Adjust difficulty by changing number of zeros
CHALLENGE_DIFFICULTY=2

# default server host and port
HOST=0.0.0.0
PORT=8080

# quotes configuration file path
QUOTES_CONFIG_PATH=/path/to/quotes.yml
```

# Client environment settings
```
#default pow server host and port
SERVER_HOST=server
SERVER_PORT=8080

# default delay of requests
REQUESTS_DELAY_MS=100
```

# Quotes Configuration
The server uses a YAML file to store quotes. Example format:
```yaml
quotes:
  - "Quote 1"
  - "Quote 2"
  - "Quote 3"
```

# Development
* Go 1.24 or later required
* Docker and Docker Compose for containerized deployment
* golangci-lint for code quality checks
