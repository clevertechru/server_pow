FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o server ./cmd/server

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/server .
COPY config/quotes.yml ./config/
COPY config/server.yml ./config/

EXPOSE 8080

CMD ["./server"] 