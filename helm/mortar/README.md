# Mortar Helm Chart

Helm chart for deploying Mortar - a Kubernetes Proxy Rules Management application with frontend and backend components.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+
- ProxyRule CRD installed in your cluster
- `proxy-rules` namespace (or configure a different namespace in values.yaml)

## Installation

### Basic Installation

Install the chart with default values:

```bash
helm install mortar ./helm/mortar
```

### Installation with Ingress

Install with ingress enabled:

```bash
helm install mortar ./helm/mortar \
  --set ingress.enabled=true \
  --set ingress.hosts[0].host=mortar.example.com \
  --set ingress.hosts[0].paths[0].path=/ \
  --set ingress.hosts[0].paths[0].pathType=Prefix
```

### Installation with Custom Values

Create a custom values file:

```yaml
# custom-values.yaml
backend:
  image:
    repository: your-registry/mortar-backend
    tag: "v1.0.0"

frontend:
  image:
    repository: your-registry/mortar-frontend
    tag: "v1.0.0"

ingress:
  enabled: true
  className: nginx
  hosts:
    - host: mortar.example.com
      paths:
        - path: /
          pathType: Prefix
```

Then install:

```bash
helm install mortar ./helm/mortar -f custom-values.yaml
```

## Configuration

The following table lists the configurable parameters and their default values:

### Global Settings

| Parameter | Description | Default |
|-----------|-------------|---------|
| `global.proxyRulesNamespace` | Namespace where ProxyRule CRDs are managed | `proxy-rules` |

### Backend Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `backend.enabled` | Enable backend deployment | `true` |
| `backend.replicaCount` | Number of backend replicas | `1` |
| `backend.image.repository` | Backend image repository | `registry.gitlab.com/bausteln/mortar-backend` |
| `backend.image.tag` | Backend image tag | `latest` |
| `backend.image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `backend.service.type` | Backend service type | `ClusterIP` |
| `backend.service.port` | Backend service port | `8080` |
| `backend.resources.limits.cpu` | CPU limit | `500m` |
| `backend.resources.limits.memory` | Memory limit | `512Mi` |
| `backend.resources.requests.cpu` | CPU request | `100m` |
| `backend.resources.requests.memory` | Memory request | `128Mi` |
| `backend.serviceAccount.create` | Create service account for backend | `true` |
| `backend.serviceAccount.name` | Service account name (auto-generated if empty) | `""` |

### Frontend Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `frontend.enabled` | Enable frontend deployment | `true` |
| `frontend.replicaCount` | Number of frontend replicas | `1` |
| `frontend.image.repository` | Frontend image repository | `registry.gitlab.com/bausteln/mortar-frontend` |
| `frontend.image.tag` | Frontend image tag | `latest` |
| `frontend.image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `frontend.service.type` | Frontend service type | `ClusterIP` |
| `frontend.service.port` | Frontend service port | `80` |
| `frontend.resources.limits.cpu` | CPU limit | `200m` |
| `frontend.resources.limits.memory` | Memory limit | `256Mi` |
| `frontend.resources.requests.cpu` | CPU request | `50m` |
| `frontend.resources.requests.memory` | Memory request | `64Mi` |

### Ingress Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `ingress.enabled` | Enable ingress | `false` |
| `ingress.className` | Ingress class name | `nginx` |
| `ingress.annotations` | Ingress annotations | `{}` |
| `ingress.hosts` | Ingress hosts configuration | See values.yaml |
| `ingress.tls` | Ingress TLS configuration | `[]` |

### RBAC Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `rbac.create` | Create RBAC resources | `true` |
| `rbac.rules` | RBAC rules for backend ServiceAccount | See values.yaml |

## RBAC Permissions

The backend pod runs with a ServiceAccount that has the following permissions:

- **API Group**: `bausteln.io`
- **Resources**: `proxyrules`
- **Verbs**: `get`, `list`, `watch`, `create`, `update`, `patch`, `delete`

These permissions are granted via a ClusterRole and ClusterRoleBinding.

## Accessing the Application

### With Ingress Disabled

Port-forward to access the frontend:

```bash
kubectl port-forward svc/mortar-frontend 8080:80
```

Visit http://localhost:8080

### With Ingress Enabled

Access via the configured hostname:

```bash
curl http://mortar.example.com
```

### Accessing the Backend API

Port-forward to the backend:

```bash
kubectl port-forward svc/mortar-backend 8080:8080
```

Test the API:

```bash
# List all proxyrules
curl http://localhost:8080/api/proxyrules

# Get a specific proxyrule
curl http://localhost:8080/api/proxyrules/my-rule

# Create a proxyrule
curl -X POST http://localhost:8080/api/proxyrules \
  -H "Content-Type: application/json" \
  -d '{"apiVersion":"bausteln.io/v1","kind":"Proxyrule","metadata":{"name":"test-rule"},"spec":{...}}'
```

## Upgrading

```bash
helm upgrade mortar ./helm/mortar -f custom-values.yaml
```

## Uninstalling

```bash
helm uninstall mortar
```

## Troubleshooting

### Backend Pod Cannot Access Kubernetes API

Ensure the `proxy-rules` namespace exists and the backend ServiceAccount has proper RBAC permissions:

```bash
kubectl create namespace proxy-rules
kubectl get clusterrole mortar-backend
kubectl get clusterrolebinding mortar-backend
```

### Frontend Cannot Connect to Backend

Check that both services are running:

```bash
kubectl get svc mortar-backend mortar-frontend
kubectl get pods
```

### Ingress Not Working

Verify ingress configuration:

```bash
kubectl get ingress
kubectl describe ingress mortar-frontend
```

Ensure your ingress controller is properly configured and the DNS is pointing to your cluster.

## License

Copyright Bausteln.ch
