FROM golang:1.21-alpine

WORKDIR /app

COPY server.go .
COPY go.mod .
COPY config ./config

RUN go build -o server .

EXPOSE 8080

CMD ["./server"] 