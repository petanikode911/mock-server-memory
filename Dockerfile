# Stage 1: Build the application
FROM golang:1.23-alpine AS builder

# Install git (needed for Go modules)
RUN apk add --no-cache git

# Set the working directory in the container
WORKDIR /app

# Copy go.mod and go.sum to the container to leverage Docker cache
COPY go.mod ./

# Run go mod tidy to ensure dependencies are updated
RUN go mod tidy

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
