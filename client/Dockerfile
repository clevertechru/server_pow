FROM golang:1.21-alpine

WORKDIR /app

COPY client.go .
COPY go.mod .
COPY config ./config

RUN go build -o client .

CMD ["./client"] 