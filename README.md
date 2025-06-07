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

| Environment Variable | Description                                   | Required | Default |
| -------------------- | --------------------------------------------- | -------- | ------- |
| `MCPTLS_CERT_FILE`   | Path to the TLS certificate file              | No       | _unset_ |
| `MCPTLS_KEY_FILE`    | Path to the TLS private key file              | No       | _unset_ |
| `MCPTLS_ENABLED`     | Whether TLS will be enforced                  | No       | _unset_ |
| `MCPTLS_SERVER_PORT` | Port the server listens on                    | No       | `8080`  |
| `MCPTLS_LOG_LEVEL`   | Log verbosity level (`debug`, `info`, `warn`) | No       | `info`  |

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

```bash
docker run --name mcp-tls-server \
  -p 8080:8080 \
  -v "$(pwd)/certs:/app/certs:ro" \
  -d \
  mcp-tls-server
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
curl --cacert certs/ca.crt \
     --cert certs/client.crt \
     --key certs/client.key \
     -X POST https://localhost:8443/api/tools/validate \
     -H "Content-Type: application/json" \
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
