FROM golang:1.16-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o slack-notifier

FROM alpine:3.14
WORKDIR /app
COPY --from=builder /app/slack-notifier /app/
ENTRYPOINT ["/app/slack-notifier"]
