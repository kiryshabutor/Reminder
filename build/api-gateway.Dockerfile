FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/api-gateway ./cmd/api-gateway

FROM alpine:3.19

RUN apk --no-cache add ca-certificates

RUN adduser -D -g '' appuser

WORKDIR /app

COPY --from=builder /app/api-gateway .

RUN chown -R appuser:appuser /app

USER appuser

EXPOSE 8080

CMD ["./api-gateway"]
