# Mortar Backend

A Kubernetes-based API server for managing ProxyRule custom resources in a Kubernetes cluster.

## Overview

Mortar Backend provides a REST API to perform CRUD operations on `Proxyrule` custom resources in the `proxy-rules` namespace. It uses the Kubernetes dynamic client to interact with your cluster.

## Prerequisites

- Go 1.25.1 or later
- Access to a Kubernetes cluster
- `~/.kube/config` file configured with cluster credentials
- `Proxyrule` CRD installed in your cluster

## Installation

```bash
# Clone the repository
git clone https://gitlab.bausteln.ch/net-core/reverse-proxy/mortar-backend.git
cd mortar-backend

# Install dependencies
go mod tidy

# Build the application
go build -o mortar-backend
```

## Running the Server

```bash
# Run directly with Go
go run main.go

# Or run the compiled binary
./mortar-backend
```

The server will start on port `8080` by default.

## Project Structure

```
mortar-backend/
├── main.go                           # Application entry point
├── internal/
│   ├── k8s/
│   │   └── client.go                # Kubernetes client initialization
│   ├── handlers/
│   │   └── proxyrules.go            # HTTP request handlers for CRUD operations
│   └── server/
│       └── server.go                # Server setup and routing
```

## ProxyRule Resource Schema

ProxyRule resources have the following structure:

```yaml
apiVersion: bausteln.io/v1
kind: Proxyrule
metadata:
  name: example-rule
  namespace: proxy-rules
spec:
  domain: example.com              # Required: The domain to proxy
  destination: backend-service     # Required: The destination to route traffic to
  port: 8080                       # Optional: The destination port
  tls: true                        # Optional: Enable TLS (default: true)
```

## API Endpoints

All endpoints are prefixed with `/api/proxyrules`.

### 1. List All ProxyRules

**Endpoint:** `GET /api/proxyrules`

**Description:** Retrieves all ProxyRule resources in the `proxy-rules` namespace.

**Example:**
```bash
curl http://localhost:8080/api/proxyrules
```

**Response:**
```json
{
  "apiVersion": "v1",
  "kind": "List",
  "items": [
    {
      "apiVersion": "bausteln.io/v1",
      "kind": "Proxyrule",
      "metadata": {
        "name": "example-rule",
        "namespace": "proxy-rules"
      },
      "spec": {
        "domain": "example.com",
        "destination": "backend-service",
        "port": 8080,
        "tls": true
      }
    }
  ]
}
```

---

### 2. Get a Specific ProxyRule

**Endpoint:** `GET /api/proxyrules/{name}`

**Description:** Retrieves a specific ProxyRule by name.

**Example:**
```bash
curl http://localhost:8080/api/proxyrules/example-rule
```

**Response:**
```json
{
  "apiVersion": "bausteln.io/v1",
  "kind": "Proxyrule",
  "metadata": {
    "name": "example-rule",
    "namespace": "proxy-rules",
    "uid": "abc-123",
    "resourceVersion": "12345"
  },
  "spec": {
    "domain": "example.com",
    "destination": "backend-service",
    "port": 8080,
    "tls": true
  }
}
```

---

### 3. Create a ProxyRule

**Endpoint:** `POST /api/proxyrules`

**Description:** Creates a new ProxyRule resource.

**Example:**
```bash
curl -X POST http://localhost:8080/api/proxyrules \
  -H "Content-Type: application/json" \
  -d '{
    "metadata": {
      "name": "my-app-rule"
    },
    "spec": {
      "domain": "myapp.example.com",
      "destination": "myapp-backend-service",
      "port": 8080,
      "tls": true
    }
  }'
```

**Minimal Example (without optional fields):**
```bash
curl -X POST http://localhost:8080/api/proxyrules \
  -H "Content-Type: application/json" \
  -d '{
    "metadata": {
      "name": "simple-rule"
    },
    "spec": {
      "domain": "simple.example.com",
      "destination": "simple-backend"
    }
  }'
```

**Response:** Returns the created resource with a `201 Created` status.

---

### 4. Update a ProxyRule

**Endpoint:** `PUT /api/proxyrules/{name}`

**Description:** Updates an existing ProxyRule resource.

**Example:**
```bash
curl -X PUT http://localhost:8080/api/proxyrules/my-app-rule \
  -H "Content-Type: application/json" \
  -d '{
    "metadata": {
      "name": "my-app-rule"
    },
    "spec": {
      "domain": "myapp.example.com",
      "destination": "myapp-backend-service-v2",
      "port": 9090,
      "tls": false
    }
  }'
```

**Response:** Returns the updated resource with a `200 OK` status.

**Notes:**
- The name in the URL must match the name in the request body
- Include all fields you want to keep (partial updates are not supported)

---

### 5. Delete a ProxyRule

**Endpoint:** `DELETE /api/proxyrules/{name}`

**Description:** Deletes a ProxyRule resource.

**Example:**
```bash
curl -X DELETE http://localhost:8080/api/proxyrules/my-app-rule
```

**Response:** Returns `204 No Content` on success.

---

## Complete Usage Example

Here's a complete workflow example:

```bash
# 1. Start the server
go run main.go

# 2. Create a new ProxyRule
curl -X POST http://localhost:8080/api/proxyrules \
  -H "Content-Type: application/json" \
  -d '{
    "metadata": {
      "name": "test-rule"
    },
    "spec": {
      "domain": "test.example.com",
      "destination": "test-backend",
      "port": 3000,
      "tls": true
    }
  }'

# 3. List all rules
curl http://localhost:8080/api/proxyrules

# 4. Get the specific rule
curl http://localhost:8080/api/proxyrules/test-rule

# 5. Update the rule
curl -X PUT http://localhost:8080/api/proxyrules/test-rule \
  -H "Content-Type: application/json" \
  -d '{
    "metadata": {
      "name": "test-rule"
    },
    "spec": {
      "domain": "test.example.com",
      "destination": "test-backend-v2",
      "port": 4000,
      "tls": true
    }
  }'

# 6. Delete the rule
curl -X DELETE http://localhost:8080/api/proxyrules/test-rule
```

## Error Handling

The API returns standard HTTP status codes:

- `200 OK` - Successful GET/PUT request
- `201 Created` - Successful POST request
- `204 No Content` - Successful DELETE request
- `400 Bad Request` - Invalid request format or missing required fields
- `404 Not Found` - Resource not found
- `405 Method Not Allowed` - HTTP method not supported for endpoint
- `500 Internal Server Error` - Server or Kubernetes API error

Error responses include a descriptive message in the response body.

## Configuration

The following configuration is currently hardcoded but can be customized in the code:

- **Namespace:** `proxy-rules` (defined in `internal/handlers/proxyrules.go`)
- **Server Port:** `8080` (defined in `main.go`)
- **API Group:** `bausteln.io` (defined in `internal/handlers/proxyrules.go`)
- **API Version:** `v1` (defined in `internal/handlers/proxyrules.go`)
- **Kubeconfig Path:** `~/.kube/config` (defined in `internal/k8s/client.go`)

## Development

### Building
```bash
go build -o mortar-backend
```

### Running Tests
```bash
go test ./...
```
