# PiShop Operator Configuration

## Environment Variables

The PiShop operator can be configured using the following environment variables:

### Required Configuration

- `MONGO_URI`: MongoDB connection URI (required)
- `MONGO_USERNAME`: MongoDB admin username (default: "admin")
- `MONGO_PASSWORD`: MongoDB admin password (default: "password")

### Optional Configuration

- `BASE_DOMAIN`: Base domain for default PR domains (default: "shop.pilab.hu")
- `GITHUB_USERNAME`: GitHub username for container registry
- `GITHUB_TOKEN`: GitHub token for container registry
- `GITHUB_EMAIL`: GitHub email for container registry

## Command-Line Flags

The operator also supports command-line flags that override environment variables:

```bash
./pishop-operator \
  --mongo-uri="mongodb://localhost:27017" \
  --base-domain="example.com" \
  --github-username="myuser" \
  --github-token="ghp_xxx"
```

### Available Flags

- `--mongo-uri`: MongoDB connection URI (required)
- `--mongo-username`: MongoDB admin username
- `--mongo-password`: MongoDB admin password
- `--base-domain`: Base domain for default PR domains
- `--github-username`: GitHub username for container registry
- `--github-token`: GitHub token for container registry
- `--github-email`: GitHub email for container registry
- `--metrics-bind-address`: Metrics server bind address (default: ":8080")
- `--health-probe-bind-address`: Health probe bind address (default: ":8081")
- `--leader-elect`: Enable leader election (default: false)

## Base Domain Configuration

The `BASE_DOMAIN` environment variable (or `--base-domain` flag) controls the default domain pattern used for PR-based environments. When a PRStack doesn't specify a custom domain, the operator will use the pattern:

```
pr-{prNumber}.{BASE_DOMAIN}
```

### Examples

**Default configuration (BASE_DOMAIN="shop.pilab.hu"):**
- PR #123 → `pr-123.shop.pilab.hu`

**Custom base domain (BASE_DOMAIN="example.com"):**
- PR #123 → `pr-123.example.com`

**Custom domain override:**
```yaml
apiVersion: shop.pilab.hu/v1alpha1
kind: PRStack
metadata:
  name: production-tenant
spec:
  prNumber: "tenant1"
  customDomain: "magicshop.hu"  # Overrides default pattern
  # ... other fields
```

## Deployment Configuration

### Environment Variables in Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pishop-operator
spec:
  template:
    spec:
      containers:
      - name: manager
        env:
        - name: BASE_DOMAIN
          value: "your-domain.com"
        - name: MONGO_URI
          valueFrom:
            secretKeyRef:
              name: mongodb-credentials
              key: uri
        # ... other environment variables
```

### Using ConfigMaps for Configuration

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: pishop-operator-config
data:
  BASE_DOMAIN: "your-domain.com"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pishop-operator
spec:
  template:
    spec:
      containers:
      - name: manager
        envFrom:
        - configMapRef:
            name: pishop-operator-config
        # ... other configuration
```

## Multi-Tenant Deployment

For multi-tenant deployments across different domains, you can deploy multiple operator instances with different base domains:

```yaml
# Instance 1: shop.pilab.hu
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pishop-operator-shop
spec:
  template:
    spec:
      containers:
      - name: manager
        env:
        - name: BASE_DOMAIN
          value: "shop.pilab.hu"

---
# Instance 2: example.com
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pishop-operator-example
spec:
  template:
    spec:
      containers:
      - name: manager
        env:
        - name: BASE_DOMAIN
          value: "example.com"
```

## TLS Configuration

For production deployments with custom domains, configure TLS secrets:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: magicshop-tls
type: kubernetes.io/tls
data:
  tls.crt: # Base64 encoded certificate
  tls.key: # Base64 encoded private key
```

Then reference the secret in your PRStack:

```yaml
apiVersion: shop.pilab.hu/v1alpha1
kind: PRStack
metadata:
  name: production-tenant
spec:
  prNumber: "tenant1"
  customDomain: "magicshop.hu"
  ingressTlsSecretName: "magicshop-tls"
  # ... other fields
```

## Validation

The operator validates configuration at startup:

- `MONGO_URI` must be provided
- `BASE_DOMAIN` must be a valid domain name
- GitHub credentials are optional but recommended for image pulls

## Troubleshooting

### Common Issues

1. **Invalid base domain**: Ensure `BASE_DOMAIN` is a valid domain name without protocol or path
2. **MongoDB connection**: Verify `MONGO_URI` is accessible from the operator pod
3. **GitHub authentication**: Check that `GITHUB_TOKEN` has appropriate permissions

### Logs

Check operator logs for configuration validation:

```bash
kubectl logs -n pishop-operator-system deployment/pishop-operator
```

Look for startup messages indicating configuration values:

```
starting manager version=1.0.0 commit=abc123 buildDate=2024-01-01T00:00:00Z
```
