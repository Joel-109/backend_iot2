# Use the Golang Alpine image as the base image for building

FROM golang:alpine AS builder
 
# Install build dependencies for CGO and SQLite

RUN apk add --no-cache gcc musl-dev
 
# Set the working directory inside the container

WORKDIR /app
 
# Copy the go.mod and go.sum files to download dependencies

COPY go.mod go.sum ./

RUN go mod download
 
# Copy the entire project directory into the container

COPY . .
 
# Build the Go application with CGO enabled

RUN CGO_ENABLED=1 GOOS=linux go build -o myapp ./cmd
 
# Use a smaller base image for the final runtime

FROM alpine:latest
 
# Install any necessary runtime dependencies

# Since github.com/mattn/go-sqlite3 embeds SQLite, no additional SQLite libs are needed

RUN apk add --no-cache libc6-compat
 
# Set the working directory

WORKDIR /root/
 
# Copy the compiled binary from the builder stage

COPY --from=builder /app/myapp .
 
# Copy any additional files needed at runtime

COPY --from=builder /app/query.sql .

COPY --from=builder /app/schema.sql .

COPY --from=builder /app/Values.db .
 
# Expose the port your application will run on

EXPOSE 1883
 
# Command to run the application

CMD ["./myapp"]
 