
FROM --platform=linux/amd64 golang:1.23 AS base-golang

RUN apt-get update

# Always set workdir into application root
WORKDIR /app

# Copy the source code into container for compiling
COPY . /app

# tidy
RUN go mod tidy

# Compile the binary
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -v -buildvcs=false -o /app/mock-server-memory

# Run the app inside distroless
FROM asia-southeast2-docker.pkg.dev/dogwood-wharf-316804/base-image/distroless-go

# Copy release binary that already compiled into distroless
COPY --from=base-golang /app/mock-server-memory mock-server-memory 

CMD ["./mock-server-memory"]

