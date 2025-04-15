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
## PoW Algorithm Choice:
* Using SHA256 with a nonce-based challenge
* The server generates a random quote and a target (currently set to 2)
* The client must find a nonce that, when combined with the quote and timestamp, produces a hash starting with the target
* This is computationally intensive but verifiable quickly
* The difficulty can be adjusted by changing the number of zeros in the target
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
* Go 1.21 or later required
* Docker and Docker Compose for containerized deployment
* golangci-lint for code quality checks
