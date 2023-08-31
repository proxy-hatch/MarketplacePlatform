# Use the standard Go image to create a build artifact
FROM golang:1.19 as builder

# Set the working directory
WORKDIR /app

# Copy the Go module files
COPY go.mod go.sum ./

# Download the Go modules
RUN go mod download

# Copy the source code
COPY cmd/ cmd/
COPY pkg/ pkg/

# Build the application
## dev env: arm64
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o go-cli-app cmd/main.go && chmod +x go-cli-app
#RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o go-cli-app cmd/main.go && chmod +x go-cli-app

# Use the alpine image for a minimal final image
FROM alpine:3.14

# Set the working directory
WORKDIR /app

# Copy the build artifact from the builder stage
COPY --from=builder /app/go-cli-app .

# Run the application
CMD ["./go-cli-app"]
