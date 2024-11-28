# Stage 1: Build the Go application
FROM golang:1.21 AS builder
# Metadata
LABEL maintainer="your_email@example.com"
LABEL description="Dockerfile for Forum app"
LABEL version="1.0"
# Set environment variables for Go
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64
# Set the working directory
WORKDIR /app
# Copy the Go modules manifest and download dependencies
COPY go.mod go.sum ./
RUN go mod download
# Copy the rest of the source code
COPY . .
# Build the Go application
RUN go build -o forum main.go
# Stage 2: Run the application
FROM alpine:latest
# Metadata
LABEL maintainer="your_email@example.com"
LABEL description="Runtime container for Forum app"
LABEL version="1.0"
# Install necessary certificates
RUN apk --no-cache add ca-certificates
# Set the working directory
WORKDIR /app
# Copy the built binary and required resources from the builder stage
COPY --from=builder /app/forum ./
COPY --from=builder /app/static ./static
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/forum.db ./forum.db
# Expose the application port
EXPOSE 8080
# Run the application
CMD ["./forum"]