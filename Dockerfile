FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .

RUN go mod download
RUN go build -o server ./server
RUN go build -o client ./client

EXPOSE 3000

FROM alpine:latest AS server
WORKDIR /app
COPY --from=builder /app/server .
CMD ["./server"]

FROM alpine:latest AS client
WORKDIR /app
COPY --from=builder /app/client .
CMD ["./client"]
