FROM golang:1.21-alpine

WORKDIR /app

COPY ./cmd/server ./cmd/server
COPY ./internal/server ./internal/server
COPY ./pkg ./pkg
COPY go.mod .

RUN go build -o server ./cmd/server

EXPOSE 8080

CMD ["./server"] 