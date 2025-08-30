# Use official golang image as builder
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build only the fabric binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o fabric ./cmd/fabric

# Use alpine for final image
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/fabric ./fabric

# Make binary executable
RUN chmod +x ./fabric

# Create config directory
RUN mkdir -p /root/.config/fabric

# Default port (Railway will override this)
ENV PORT=8080

# Create startup script that properly handles the PORT environment variable
RUN echo '#!/bin/sh' > start.sh && \
    echo 'PORT=${PORT:-8080}' >> start.sh && \
    echo 'echo "Starting Fabric API server on port $PORT"' >> start.sh && \
    echo 'exec ./fabric --serve --address "0.0.0.0:$PORT"' >> start.sh && \
    chmod +x start.sh

# Expose default port (Railway will map to the correct one)
EXPOSE 8080

# Run the startup script
CMD ["./start.sh"]