# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the verificator binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o verificator ./cmd/verificator

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/verificator .

# Copy email templates
COPY --from=builder /app/web/templates/email ./web/templates/email

CMD ["./verificator"]
