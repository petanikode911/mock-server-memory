# Step 1: Build the Go application
FROM golang:1.20-alpine AS build

# Set the working directory for the Go build
WORKDIR /app

# Copy the Go application code into the container
COPY . .

# Build the Go application
RUN go build -o memory-stress .

# Step 2: Create a minimal image to run the application
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /root/

# Copy the built Go binary from the build stage
COPY --from=build /app/memory-stress .

# Set the default command to run the Go application
CMD ["./memory-stress"]
