FROM golang:1.21-alpine

WORKDIR /app

COPY client.go .
COPY go.mod .

RUN go build -o client .

CMD ["./client"] 