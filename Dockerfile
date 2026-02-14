# Use the official Golang image as a builder
# Go 1.23 will auto-download Go 1.24 toolchain when required by go.mod
FROM golang:1.23 as builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files first
COPY go.mod go.sum ./

# Enable automatic Go toolchain download for Go 1.24 requirement
ENV GOTOOLCHAIN=auto
RUN go mod download

# Copy the entire application code to the container
COPY . .

# Build the Go application
# Adapted to build the entry point at ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o app ./cmd/api

# Use a smaller Debian Slim base image for the final container
FROM debian:bullseye-slim

# Install necessary libraries, CA certificates, and set up timezone
RUN apt-get update && apt-get install -y --no-install-recommends \
    libc6 \
    ca-certificates \
    tzdata \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* \
    && update-ca-certificates

# Set the working directory inside the final container
WORKDIR /app

# Copy the built Go application from the builder
COPY --from=builder /app/app .

# Expose the port if your application listens on a specific port
EXPOSE 8080

# Set environment variables for better TLS handling
ENV GODEBUG=x509ignoreCN=0
ENV SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt

# Run the Go application
CMD ["./app"]
