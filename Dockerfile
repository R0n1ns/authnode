# Use the official Golang image
FROM golang:1.22-alpine AS builder

# Set working directory
WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o authmicro ./cmd/auth

# Use a minimal alpine image for the final stage
FROM alpine:3.19

# Install CA certificates for HTTPS connections
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/authmicro /app/authmicro
COPY --from=builder /app/migrations /app/migrations

# Expose the ports
EXPOSE 8000 9000

# Set entry point
ENTRYPOINT ["/app/authmicro"]
