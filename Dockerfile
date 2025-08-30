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

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o fabric ./cmd/fabric

# Use alpine for final image for smaller size
FROM alpine:latest

# Install ca-certificates for HTTPS requests and bash for startup script
RUN apk --no-cache add ca-certificates tzdata bash

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/fabric .

# Create startup script to handle PORT environment variable
RUN echo '#!/bin/bash' > start.sh && \
    echo 'PORT=${PORT:-8080}' >> start.sh && \
    echo './fabric --serve --address ":$PORT"' >> start.sh && \
    chmod +x start.sh

# Create necessary directories
RUN mkdir -p /root/.config/fabric

# Set the binary as executable
RUN chmod +x ./fabric

# Run the startup script
CMD ["./start.sh"]