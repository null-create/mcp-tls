# MCP-TLS Tool Validation Server

### ⚠️ This project is in early stage development ⚠️

A lightweight utility server that validates tool definitions for integrity and schema correctness. This server is intended to be used as part of a broader MCP-compatible toolchain but can run independently for testing or CI verification of tool definitions.

## 🔧 Features

- 📦 JSON-RPC 2.0-compliant request validation
- 🔐 TLS transport support (with optional mTLS enforcement)
- 🔍 Tool schema fingerprinting and checksum validation
- ⚡ Fast HTTP API built with [Chi](https://github.com/go-chi/chi)
- 🧪 Unit tested components with Go test support

## 📁 Project Structure

```
.
├── .github
│   └── workflows/        # CI and release GitHub Actions
├── .gitignore
├── Dockerfile
├── README.md
├── VERSION
├── certs                 # Optional certs directory
├── cmd                   # Application entry points
│   └── server/
├── go.mod
├── go.sum
└── pkg
    ├── config/           # Project configurations
    ├── logs/             # Log output directory
    ├── mcp/              # Core MCP-TLS data structures
    ├── server/           # HTTP server, routes, and handlers
    ├── tls/              # TLS transport encryption support
    ├── util/             # JSON helpers
    └── validate/         # Tool validation logic
```

## 🚀 Getting Started

### Prerequisites

- Go 1.21+
- TLS certificate (self-signed or CA-issued)

### Configuration

Optional environment variables

| Environment Variable | Description                                   | Required | Default s        |
| -------------------- | --------------------------------------------- | -------- | ---------------- |
| `MCPTLS_SERVER_PORT` | Port the server listens on                    | No       | `9090`           |
| `MCPTLS_SERVER_ADDR` | Server address                                | No       | `localhost:9090` |
| `MCPTLS_LOG_LEVEL`   | Log verbosity level (`debug`, `info`, `warn`) | No       | `info`           |

### Build and Run a binary

```bash
go build -o bin/server ./cmd/server
chmod +x ./bin/server
./bin/server
```

### Build and run with Docker

```bash
docker build -t mcp-tls-server .
```

Run basic with basic configs

```bash
docker run --name mcp-tls-server \
  -p 9090:9090 \
  -e MCPTLS_SERVER_ADDR="0.0.0.0:9090" \
  -d \
  mcp-tls-server
```

Run using `docker compose`

```bash
docker compose up -d
```

```bash
docker compose down
```

### API Endpoints

#### `POST /api/tools/validate`

Validates a single tool definition for schema and checksum integrity.

**Example**

```bash
curl -X POST https://localhost:8443/api/tools/validate \
     -H "Content-Type: application/json" \
     -d @tool.json
```

**Example with TLS enabled:**

```bash
curl -X POST https://localhost:8443/api/tools/validate \
     -H "Content-Type: application/json" \
     --cacert certs/ca.crt \
     --cert certs/client.crt \
     --key certs/client.key \
     -d @tool.json
```

#### Request Schema (`tool.json`)

```json
{
  "name": "echo",
  "description": "Simple echo tool",
  "parameters": {
    "message": {
      "type": "string"
    }
  },
  "inputSchema": {
    "type": "object",
    "properties": {
      "message": { "type": "string" }
    },
    "required": ["message"]
  }
}
```

## 🧪 Testing

```bash
go test -v ./...
```

## 🔐 TLS Configuration

TLS is mandatory by default.

### Supported Flags:

| Flag             | Description                               |
| ---------------- | ----------------------------------------- |
| `--cert`         | Path to TLS certificate file (PEM format) |
| `--key`          | Path to TLS private key (PEM format)      |
| `--ca`           | Path to CA cert for verifying clients     |
| `--require-mtls` | Require client certificate verification   |
| `--addr`         | Listen address (default: `:8443`)         |
