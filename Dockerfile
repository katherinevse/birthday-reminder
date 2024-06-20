FROM golang:1.18-alpine

WORKDIR /app

COPY app .

RUN go build -o server ./cmd/server

CMD ["./server"]
