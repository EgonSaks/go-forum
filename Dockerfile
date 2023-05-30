# Start with a lightweight base image
FROM golang:latest

# Install SQLite3
RUN apt-get update && apt-get install -y sqlite3

# Set metadata for the image
LABEL version="1.0"

# Set the working directory
WORKDIR /app

# Copy the source code into the container
COPY . .

# Build the Go application
RUN go build ./cmd/web

# Expose port 8080 to the outside world
EXPOSE 8080

# Set the entrypoint command to start the server
ENTRYPOINT ["./web"]