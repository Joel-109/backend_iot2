# Use the official Golang image as the base image for building
FROM golang:1.24.2 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project directory into the container
COPY . .

# Build the Go application
# Assuming the main.go is the entry point
RUN CGO_ENABLED=0 GOOS=linux go build -o myapp ./cmd

# Use a smaller base image for the final runtime
FROM alpine:latest

# Install any necessary runtime dependencies (e.g., for SQLite if used)
RUN apk add --no-cache libc6-compat

# Set the working directory
WORKDIR /root/

# Copy the compiled binary from the builder stage
COPY --from=builder /app/myapp .

# Copy any additional files needed at runtime (e.g., SQL files for schema)
COPY --from=builder /app/query.sql .
COPY --from=builder /app/schema.sql .
COPY --from=builder /app/Values.db .

# Expose the port your application will run on (default MQTT port is 1883)
EXPOSE 1883

# Command to run the application
CMD ["./myapp"]