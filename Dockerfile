# Use the official Golang image to create a binary
FROM golang:latest AS build

# Set the current working directory inside the container
WORKDIR /app

# Copy the Go modules manifests
COPY go.mod ./
COPY go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -o app .

# Use a minimal base image to reduce size
FROM alpine:latest

# Set the current working directory inside the container
WORKDIR /root/

# Copy the binary from the build stage to the final stage
COPY --from=build /app/app .

# Expose the port on which the application will run
EXPOSE 8080

# Command to run the executable
CMD ["./app"]
