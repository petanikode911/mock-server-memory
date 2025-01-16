# Stage 1: Build the application
FROM golang:1.23-alpine AS builder

# Install git (needed for Go modules)
RUN apk add --no-cache git

# Create a non-root user to avoid permission issues
RUN addgroup -S go && adduser -S go -G go

# Set the working directory to /app and ensure it has correct ownership
WORKDIR /app

# Set permissions to the /app directory for the go user
RUN chown -R go:go /app

# Switch to the go user to avoid running as root
USER go

# Add the exception for the directory in Git config
RUN git config --global safe.directory /app

# Copy the source code into the container
COPY . .

# Build the Go application
RUN go build -v -o memory-stress .

# Stage 2: Create a minimal runtime image
FROM alpine:latest

# Set the working directory in the runtime image
WORKDIR /root/

# Copy the built binary from the builder stage
COPY --from=builder /app/memory-stress .

# Expose the port the application listens on
EXPOSE 8888

# Run the application
CMD ["./memory-stress"]
