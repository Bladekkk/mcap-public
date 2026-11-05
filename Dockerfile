# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy source
COPY backend/ ./backend/

# Build
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o mcap ./backend/cmd/main.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates bash

WORKDIR /root/

COPY --from=builder /app/mcap .
COPY --from=builder /app/backend/migrations ./migrations/

EXPOSE 8080

CMD ["./mcap"]
