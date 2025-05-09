FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o server ./cmd/server

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/server .
COPY config/quotes.yml ./config/quotes.yml
COPY config/server.yml ./config/server.yml

EXPOSE 8080

CMD ["./server"] 