[![Go](https://github.com/clevertechru/server_pow/actions/workflows/go.yml/badge.svg)](https://github.com/clevertechru/server_pow/actions/workflows/go.yml)
[![Build Status](https://app.travis-ci.com/clevertechru/server_pow.svg?token=bbzT95wZRUs7cXAPJccG&branch=main)](https://app.travis-ci.com/clevertechru/server_pow)
# TCP server protected from DDOS attacks with the Proof of Work
## General
* Proper connection handling
* Timeouts to prevent hanging connections
* Worker pool for controlled concurrency
* Rate limiting
* Connection limiting
* Nonce tracking with window expiration
* Graceful shutdown
## PoW Algorithm Choice:
* Using SHA256 with a nonce-based challenge
* The server generates a random 32-byte challenge and a target (currently set to "0000")
* The client must find a nonce that, when combined with the challenge and timestamp, produces a hash starting with the target
* This is computationally intensive but verifiable quickly
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
CHALLENGE_DIFFICULTY=0000

# default server host and port
HOST=0.0.0.0
PORT=8080
```
# Client environment settings
```
#default pow server host and port
SERVER_HOST=server
SERVER_PORT=8080

# default delay of requests
REQUESTS_DELAY_MS=100
```
