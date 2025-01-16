# Stage 1: Build the application
FROM golang:1.23-alpine AS builder

# Install git (needed for Go modules)
RUN apk add --no-cache git

# Create a non-root user to avoid permission issues
RUN addgroup -S go && adduser -S go -G go

# Set the working directory to /app
WORKDIR /app

# Add the exception for the directory in Git config
RUN git config --global --add safe.directory /app

# Copy go.mod and go.sum to the container, and set proper permissions
COPY go.mod ./
RUN chown -R go:go /app

# Switch to the go user to avoid running as root
USER go

# Initialize the Go module if it's not already present
RUN [ ! -f go.mod ] && go mod init || echo "go.mod already initialized"

# Run go mod tidy to ensure dependencies are updated
RUN go mod tidy

# Copy the source code into the container
COPY . .

# Build the Go application with VCS stamping disabled
RUN go build -v -buildvcs=false -o /app/memory-stress .

# Stage 2: Create a minimal runtime image
FROM alpine:latest

# Set the working directory in the runtime image
WORKDIR /root/

# Copy the built binary from the builder stage and ensure permissions are set correctly
COPY --from=builder /app/memory-stress .

# Ensure the copied binary has the correct permissions
RUN chmod +x /root/memory-stress

# Expose the port the application listens on
EXPOSE 8888

