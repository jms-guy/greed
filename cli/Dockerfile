# Build 
FROM golang:1.24-bullseye AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY cli/ cli/
COPY models/ models/
RUN cd cli && CGO_ENABLED=0 GOOS=linux go build -o greed-cli

# Runtime 
FROM debian:stable-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /app/cli/greed-cli /app/greed-cli
ENTRYPOINT ["/app/greed-cli"]