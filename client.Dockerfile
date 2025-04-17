FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY ./cmd/client ./cmd/client
COPY ./internal/client ./internal/client
COPY ./pkg ./pkg
COPY go.mod go.sum ./
RUN go mod download
RUN go build -o client ./cmd/client

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/client .
COPY config/client.yml ./config/

CMD ["./client"] 