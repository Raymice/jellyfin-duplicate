# Dockerfile for jellyfin-duplicate application
# Multi-stage build for optimized production image

# Build stage
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

# Set necessary environment variables using build-time platform variables
ARG TARGETOS
ARG TARGETARCH

ENV CGO_ENABLED=0 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH}

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Create and set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application for the target platform
RUN go build -o jellyfin-duplicate .

# Production stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create and set working directory
WORKDIR /app

# Copy built binary from builder stage
COPY --from=builder /app/jellyfin-duplicate .

# Copy configuration files
COPY configuration/files/ configuration/files/

# Copy HTML templates
COPY server/templates/ server/templates/

# Set environment variables (these can be overridden at runtime)
ENV ENVIRONMENT="production"

# Expose the port the app runs on
EXPOSE 8080

# Set the entry point
ENTRYPOINT ["./jellyfin-duplicate"]

# Default command (can be overridden)
CMD []