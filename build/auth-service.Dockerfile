FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/auth-service ./cmd/auth-service

FROM alpine:3.19

RUN apk --no-cache add ca-certificates

RUN adduser -D -g '' appuser

WORKDIR /app

COPY --from=builder /app/auth-service .
COPY --from=builder /app/migrations ./migrations

RUN chown -R appuser:appuser /app

USER appuser

EXPOSE 50051

CMD ["./auth-service"]
