FROM --platform=$BUILDPLATFORM golang:1.21-alpine AS builder

ARG BUILDPLATFORM
ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o slack-notifier

FROM alpine:3.19
WORKDIR /ko-app
COPY --from=builder /app/slack-notifier /ko-app/tekton-slack-notify

# Add CA certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create a directory for token files
RUN mkdir -p /ko-app/tokens

# Set default entrypoint with better error handling
ENTRYPOINT ["/ko-app/tekton-slack-notify"]
