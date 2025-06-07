# Stage 1: Build the Go binary
FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build -o mcp-tls-server ./cmd/server

# Stage 2: Run the binary in a minimal image
FROM alpine:latest

ENV TLS_ENABLED=""
ENV SERVER_PORT=8080
ENV LOG_LEVEL=info

# Update image and add certificates support
RUN apk --no-cache update && \
    apk --no-cache upgrade && \
    apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/mcp-tls-server .
COPY certs/ /root/certs/

EXPOSE 8080

RUN chmod +x ./mcp-tls-server

ENTRYPOINT ["./mcp-tls-server"]
