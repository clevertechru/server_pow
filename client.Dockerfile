FROM golang:1.24-alpine

WORKDIR /app

COPY ./cmd/client ./cmd/client
COPY ./internal/client ./internal/client
COPY ./pkg ./pkg
COPY go.mod go.sum ./

RUN go build -o client ./cmd/client

CMD ["./client"] 