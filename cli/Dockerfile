# Build 
FROM golang:1.23-bullseye AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY cli/ cli/
COPY models/ models/
RUN cd cli && CGO_ENABLED=0 GOOS=linux go build -o greed-cli

# Runtime 
FROM debian:stable-slim
WORKDIR /app
COPY --from=builder /app/cli/greed-cli /app/greed-cli
ENTRYPOINT ["/app/greed-cli"]