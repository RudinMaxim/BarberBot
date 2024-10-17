# Start from the official golang image
FROM golang:1.23-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the application
RUN go build -o main ./cmd/main.go

# Expose the port the app runs on
EXPOSE 8080

# Command to run the executable
CMD ["./main"]