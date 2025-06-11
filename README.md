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
docker build -t mcptls-server .
```

Run basic with basic configs

```bash
docker run --name mcptls-server \
  -p 9090:9090 \
  -d \
  mcptls-server
```

### API Endpoints

#### `POST /api/tools/validate`

Validates a single tool definition for schema and checksum integrity.

**Example**

```bash
curl -X POST https://localhost:9090/api/tools/validate \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer <JWT>" \
     -d @tool.json
```

#### Request Schema (`tool.json`)

```json
{
  "name": "example-tool",
  "description": "This tool performs a sample operation.",
  "arguments": {
    "inputA": "value1"
  },
  "parameters": {
    "param1": "value1",
    "param2": 42,
    "param3": true
  },
  "inputSchema": {
    "type": "object",
    "properties": {
      "inputA": {
        "type": "string"
      },
      "inputB": {
        "type": "number"
      }
    },
    "required": ["inputA"]
  },
  "outputSchema": {
    "type": "object",
    "properties": {
      "outputA": {
        "type": "boolean"
      }
    },
    "required": ["outputA"]
  },
  "annotations": {
    "title": "Sample Tool",
    "readOnlyHint": true,
    "destructiveHint": false,
    "idempotentHint": true,
    "openWorldHint": false
  },
  "secMetaData": {
    "source": "trusted-registry",
    "signature": "abc123signature",
    "public_key_id": "key-456",
    "version": "1.0.0",
    "checksum": "sha256:deadbeef"
  }
}
```

## 🧪 Testing

```bash
go test -v ./...
```
