# Step 1: Build the Go application
FROM golang:1.23 as builder

# Set the current working directory in the container
WORKDIR /app

# Copy go.mod and go.sum for dependency management
COPY go.mod go.sum ./

# Download dependencies
RUN go mod tidy

# Copy the rest of the application files into the container
COPY . .

# Build the Go app (change "main" to the name of your Go binary)
RUN go build -o company-service ./cmd

# Step 2: Create the production image
FROM debian:latest

# Install necessary libraries (e.g., for running MySQL client, SSL support)
RUN apt-get update && apt-get install -y \
    libssl-dev \
    libsasl2-dev \
    libz-dev \
    && rm -rf /var/lib/apt/lists/*

# Set the working directory
WORKDIR /app

# Copy the built Go binary from the builder stage
COPY --from=builder /app/company-service .

# Copy the .env file
COPY .env .


# Expose the application port (8080 for REST API)
EXPOSE 8080

# Command to run the app
CMD ["./company-service"]
