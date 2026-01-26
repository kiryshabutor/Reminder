FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/notification-service ./cmd/notification-service

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/notification-service .

CMD ["./notification-service"]
