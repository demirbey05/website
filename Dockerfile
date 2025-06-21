# Stage 1: Build the Go application
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum and download dependencies first to leverage Docker layer caching.
# If you don't have go.mod and go.sum files yet, you can generate them by running:
# go mod init your-module-name (e.g., go mod init academic-blog)
# go mod tidy
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application, creating a static binary.
# CGO_ENABLED=0 is important for creating a static binary that can run in a minimal image.
# -ldflags "-w -s" strips debug information, reducing the binary size.
RUN CGO_ENABLED=0 GOOS=linux go build -v -o /app/main -ldflags "-w -s" .

# Stage 2: Create the final, minimal image
FROM alpine:latest

# It's a good practice to run as a non-root user.
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

WORKDIR /home/appuser

# Copy the compiled application binary from the builder stage
COPY --from=builder /app/main .

# Copy the posts directory
# Make sure you have a 'posts' directory with your Markdown files in the same directory as your Dockerfile.
COPY --chown=appuser:appgroup posts ./posts

# Expose the port the server runs on
EXPOSE 8080

# The command to run the application
ENTRYPOINT ["./main"]