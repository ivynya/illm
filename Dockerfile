FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .

RUN go get -d -v
RUN go build -o server ./server

EXPOSE 3000

FROM alpine:latest AS runner
WORKDIR /app
COPY --from=builder /app/server .

# Run the app when the container launches
CMD ["server"]
