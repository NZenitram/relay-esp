# Use the official Golang image as the base image
FROM golang:1.20-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies and verify
RUN go mod download
# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Attempt to build with verbose output
RUN go build -o relay-esp .

# Start a new stage from scratch
FROM alpine:latest  

WORKDIR /root/

COPY --from=builder /app/relay-esp .

CMD ["./relay-esp"]
