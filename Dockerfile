# Stage 1: Build the Go binary
FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o mcp-tls-server .

# Stage 2: Run the binary in a minimal image
FROM alpine:latest

RUN apk --no-cache update && \
    apk --no-cache upgrade && \
    apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/mcp-tls-server .

EXPOSE 8080

ENTRYPOINT ["./mcp-tls-server"]
