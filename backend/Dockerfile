# Build stage
FROM golang:1.23-bullseye AS builder
WORKDIR /app
COPY backend/. .
COPY models ./models
RUN CGO_ENABLED=0 GOOS=linux go build -o greed

# Runtime
FROM gcr.io/distroless/static
COPY --from=builder /app/greed /greed
ENTRYPOINT ["/greed"]