# Build stage
FROM golang:alpine AS builder

LABEL stage=gobuilder

ARG TARGETARCH
ENV CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH}

RUN apk update --no-cache && apk add --no-cache tzdata ca-certificates

WORKDIR /build

# Copy go.mod and go.sum first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN go build -ldflags="-s -w" -o /app/telegram-bot ./cmd/main.go

# Final minimal image
FROM scratch

# Copy CA certificates and timezone data
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo/Asia/Tehran /usr/share/zoneinfo/Asia/Tehran

ENV TZ=Asia/Tehran

WORKDIR /app

# Copy binary
COPY --from=builder /app/telegram-bot /app/telegram-bot

# Copy locale files for i18n
COPY --from=builder /build/internal/i18n/locales /app/locales

ENTRYPOINT ["/app/telegram-bot"]
