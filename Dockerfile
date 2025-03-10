FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o slack-notifier

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/slack-notifier /app/

# Add CA certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

ENTRYPOINT ["/app/slack-notifier"]
