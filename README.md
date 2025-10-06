# 🧱 Mortar

![Go](https://img.shields.io/badge/Go-00ADD8?style=flat&logo=go&logoColor=white)
![Kubernetes](https://img.shields.io/badge/Kubernetes-326CE5?style=flat&logo=kubernetes&logoColor=white)
![React](https://img.shields.io/badge/React-61DAFB?style=flat&logo=react&logoColor=black)
![Helm](https://img.shields.io/badge/Helm-0F1689?style=flat&logo=helm&logoColor=white)
![License](https://img.shields.io/badge/license-BSD--3--Clause-blue.svg)

> A complete solution for managing Kubernetes proxy rules with a REST API and web portal

## 🎯 Overview

Mortar provides a user-friendly way to manage reverse proxy rules in Kubernetes through custom resources (CRDs). It consists of a Go backend API, a React frontend portal, and Crossplane integration for GitOps workflows.

## 🏗️ Architecture

```
┌─────────────────┐
│  Mortar Portal  │  ← React UI
│   (Frontend)    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Mortar Backend  │  ← Go REST API
│      (API)      │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   Kubernetes    │
│  ProxyRule CRD  │  ← Custom Resources
└─────────────────┘
```

## ⚡ Quick Start

### Deploy with Helm

```bash
# Add Helm repo
helm repo add mortar https://gitlab.bausteln.ch/api/v4/projects/16/packages/helm/stable

# Install with default values
helm install mortar mortar/mortar

# Or install from source
helm install mortar ./helm/mortar
```

### Custom Installation

```bash
# With custom values
helm install mortar ./helm/mortar \
  --set frontend.ingress.enabled=true \
  --set frontend.ingress.hosts[0].host=mortar.example.com

# With external values file
helm install mortar ./helm/mortar -f custom-values.yaml
```

## 📦 Components

### 🔧 Backend (Go API)
REST API for CRUD operations on ProxyRule resources
- **Port:** 8080
- **Namespace:** `proxy-rules`
- **Image:** `reg.bausteln.ch/foss/reverse-proxy/mortar-backend`

### 🎨 Frontend (React Portal)
Web UI for managing proxy rules
- **Port:** 80
- **Features:** Create, edit, delete, and list proxy rules
- **Image:** `reg.bausteln.ch/foss/reverse-proxy/mortar-portal`

### ☸️ Crossplane
CRD definitions and compositions for GitOps workflows

## 🔌 API Endpoints

Base path: `/api/proxyrules`

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/` | List all rules |
| `GET` | `/{name}` | Get specific rule |
| `POST` | `/` | Create rule |
| `PUT` | `/{name}` | Update rule |
| `DELETE` | `/{name}` | Delete rule |

### ProxyRule Schema

```yaml
apiVersion: bausteln.io/v1
kind: Proxyrule
metadata:
  name: my-app
  namespace: proxy-rules
spec:
  domain: app.example.com    # Required
  destination: backend-svc    # Required
  port: 8080                  # Optional
  tls: true                   # Optional (default: true)
```

### Example API Call

```bash
curl -X POST http://localhost:8080/api/proxyrules \
  -H "Content-Type: application/json" \
  -d '{
    "metadata": {"name": "my-rule"},
    "spec": {
      "domain": "app.example.com",
      "destination": "backend-svc",
      "port": 8080
    }
  }'
```

## 🛠️ Development

### Prerequisites

- Go 1.25.1+
- Node.js 18+ (for portal)
- Kubernetes cluster
- `kubectl` configured
- Docker (optional)

### Backend Development

```bash
# Install dependencies
go mod tidy

# Run locally
go run main.go

# Build
go build -o mortar-backend

# Build Docker image
docker build -t mortar-backend .

# Test
go test ./...
```

### Frontend Development

```bash
cd portal

# Install dependencies
npm install

# Run dev server
npm run dev

# Build for production
npm run build

# Build Docker image
docker build -t mortar-portal .
```

### Install Crossplane CRDs

```bash
kubectl apply -f crossplane/rp/xrd-proxy.yaml
kubectl apply -f crossplane/rp/composition-proxy.yaml
kubectl apply -f crossplane/functions/
```

## 📂 Project Structure

```
mortar-backend/
├── internal/
│   ├── k8s/client.go          # Kubernetes client
│   ├── handlers/proxyrules.go # API handlers
│   └── server/server.go        # HTTP server
├── portal/                     # React frontend
│   └── src/
│       ├── components/         # React components
│       └── App.jsx            # Main app
├── helm/mortar/               # Helm chart
│   ├── Chart.yaml
│   ├── values.yaml
│   └── templates/
├── crossplane/                 # Crossplane resources
│   ├── rp/                    # ProxyRule CRD & composition
│   └── functions/             # Composition functions
├── Dockerfile                  # Backend container
└── main.go                     # Entry point
```

## ⚙️ Configuration

### Helm Values

Key configuration options in `helm/mortar/values.yaml`:

```yaml
global:
  proxyRulesNamespace: proxy-rules

backend:
  enabled: true
  replicaCount: 1
  image:
    repository: reg.bausteln.ch/foss/reverse-proxy/mortar-backend
    tag: latest

frontend:
  enabled: true
  replicaCount: 1

ingress:
  enabled: false
  className: nginx
  hosts:
    - host: mortar.example.com
```

### Environment Variables

Backend supports in-cluster and local kubeconfig:
- **In-cluster:** Automatically uses ServiceAccount
- **Local:** Uses `~/.kube/config`

## 🔐 RBAC

The Helm chart creates necessary RBAC resources:
- ServiceAccount for backend
- ClusterRole with ProxyRule permissions
- ClusterRoleBinding

## 🚀 CI/CD

Automated pipeline with GitLab CI:
1. **Test** - Go fmt/vet
2. **Build** - Docker images (backend + portal)
3. **Package** - Helm chart publishing
4. **Deploy** - Automatic deployment (optional)

## 📊 Status Codes

| Code | Description |
|------|-------------|
| 200 | Success (GET/PUT) |
| 201 | Created (POST) |
| 204 | Deleted (DELETE) |
| 400 | Bad Request |
| 404 | Not Found |
| 500 | Server Error |

## 📄 License

BSD 3-Clause License - see [LICENSE](LICENSE) file for details

## 🤝 Contributing

Contributions welcome! Please ensure:
- Code follows `go fmt` standards
- Tests pass with `go test ./...`
- Frontend builds without errors

## 🔗 Links

- **Backend Registry:** `reg.bausteln.ch/foss/reverse-proxy/mortar-backend`
- **Portal Registry:** `reg.bausteln.ch/foss/reverse-proxy/mortar-portal`
- **Helm Repository:** `https://gitlab.bausteln.ch/api/v4/projects/16/packages/helm/stable`
