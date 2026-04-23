# Multi-stage build for smaller image
FROM golang:1.26-alpine AS builder

# Install git and build dependencies
RUN apk add --no-cache git ca-certificates make

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/loviary ./cmd/api

# Final stage
FROM alpine:3.19

# Install ca certificates
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/loviary .

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

USER appuser

EXPOSE 8080

CMD ["./loviary"]
