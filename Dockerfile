# Use latest Go version
FROM golang:1.22-alpine

# Set working directory
WORKDIR /app

# Install git for dependency fetching
RUN apk add --no-cache git

# Copy dependency files
COPY go.mod go.sum ./

# Fetch dependencies
RUN go mod download

# Copy all other project files
COPY . .

# Build your daemon binary
RUN go build -o /daemon ./cmd/daemon/main.go

# Run the daemon when the container starts
ENTRYPOINT ["/daemon"]
