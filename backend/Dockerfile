# Build stage
FROM golang:1.24-bullseye AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY backend/ backend/
COPY models/ models/
RUN cd backend && CGO_ENABLED=0 GOOS=linux go build -o greed

# Runtime
FROM gcr.io/distroless/static
COPY --from=builder /app/backend/greed /greed
COPY --from=builder /app/backend/server/data/ /server/data/
ENTRYPOINT ["/greed"]