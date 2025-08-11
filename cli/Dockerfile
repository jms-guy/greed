# Build stage
FROM golang:1.23-bullseye AS builder
WORKDIR /app
COPY cli/. .
COPY models ./models
RUN CGO_ENABLED=0 GOOS=linux go build -o greed-cli

# Runtime stage (optional, for smaller images)
FROM debian:stable-slim
WORKDIR /app
COPY --from=builder /app/greed-cli /app/greed-cli
ENTRYPOINT ["/app/greed-cli"]