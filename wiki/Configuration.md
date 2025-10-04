# Configuration Guide

This guide covers all configuration options for the PiShop Operator, including environment variables, PRStack specifications, and deployment settings.

## Operator Configuration

### Environment Variables

The operator can be configured using the following environment variables:

#### Required Configuration

| Variable | Description | Example |
|----------|-------------|---------|
| `MONGO_URI` | MongoDB connection URI | `mongodb://admin:password@mongodb:27017` |
| `MONGO_USERNAME` | MongoDB admin username | `admin` |
| `MONGO_PASSWORD` | MongoDB admin password | `password` |

#### Optional Configuration

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `BASE_DOMAIN` | Base domain for default PR domains | `shop.pilab.hu` | `example.com` |
| `GITHUB_USERNAME` | GitHub username for container registry | - | `myuser` |
| `GITHUB_TOKEN` | GitHub token for container registry | - | `ghp_xxx` |
| `GITHUB_EMAIL` | GitHub email for container registry | - | `user@example.com` |
| `NATS_URL` | NATS server URL | `nats://nats:4222` | `nats://nats.example.com:4222` |
| `REDIS_URL` | Redis server URL | `redis://redis:6379` | `redis://redis.example.com:6379` |

### Command-Line Flags

The operator supports command-line flags that override environment variables:

```bash
./pishop-operator \
  --mongo-uri="mongodb://localhost:27017" \
  --base-domain="example.com" \
  --github-username="myuser" \
  --github-token="ghp_xxx"
```

#### Available Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--mongo-uri` | MongoDB connection URI | Required |
| `--mongo-username` | MongoDB admin username | `admin` |
| `--mongo-password` | MongoDB admin password | `password` |
| `--base-domain` | Base domain for default PR domains | `shop.pilab.hu` |
| `--github-username` | GitHub username for container registry | - |
| `--github-token` | GitHub token for container registry | - |
| `--github-email` | GitHub email for container registry | - |
| `--metrics-bind-address` | Metrics server bind address | `:8080` |
| `--health-probe-bind-address` | Health probe bind address | `:8081` |
| `--leader-elect` | Enable leader election | `false` |

## PRStack Configuration

### Required Fields

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `prNumber` | string | Pull request number or tenant identifier | `"123"` |

### Optional Fields

#### Basic Configuration

| Field | Type | Description | Default | Example |
|-------|------|-------------|---------|---------|
| `imageTag` | string | Docker image tag for services | `pr-{prNumber}` | `"v1.2.3"` |
| `customDomain` | string | Custom domain for ingress | `pr-{prNumber}.{BASE_DOMAIN}` | `"magicshop.hu"` |
| `ingressTlsSecretName` | string | TLS secret name for HTTPS | - | `"magicshop-tls"` |
| `active` | boolean | Controls whether stack is active | `true` | `true` |
| `environment` | string | Environment configuration | - | `"production"` |

#### Service Configuration

| Field | Type | Description | Default | Example |
|-------|------|-------------|---------|---------|
| `services` | []string | List of services to deploy | All services | `["product-service", "cart-service"]` |

Available services:
- `product-service`
- `cart-service`
- `order-service`
- `payment-service`
- `customer-service`
- `inventory-service`
- `notification-service`
- `discount-service`
- `checkout-service`
- `analytics-service`
- `auth-service`
- `graphql-service`

#### Connection Configuration

| Field | Type | Description | Default | Example |
|-------|------|-------------|---------|---------|
| `mongoURI` | string | MongoDB connection URI | Operator default | `"mongodb://custom-mongo:27017"` |
| `mongoUsername` | string | MongoDB username | Operator default | `"custom-user"` |
| `mongoPassword` | string | MongoDB password | Operator default | `"custom-pass"` |
| `natsURL` | string | NATS server URL | Operator default | `"nats://custom-nats:4222"` |
| `redisURL` | string | Redis server URL | Operator default | `"redis://custom-redis:6379"` |

#### Resource Configuration

| Field | Type | Description | Default | Example |
|-------|------|-------------|---------|---------|
| `resourceLimits.cpuLimit` | string | CPU limit per service | `"500m"` | `"1000m"` |
| `resourceLimits.memoryLimit` | string | Memory limit per service | `"1Gi"` | `"2Gi"` |
| `resourceLimits.storageLimit` | string | Storage limit for databases | `"10Gi"` | `"50Gi"` |

#### Backup Configuration

| Field | Type | Description | Default | Example |
|-------|------|-------------|---------|---------|
| `backupConfig.enabled` | boolean | Enable automatic backups | `false` | `true` |
| `backupConfig.schedule` | string | Backup schedule (cron format) | - | `"0 2 * * *"` |
| `backupConfig.retentionDays` | integer | Days to keep backups | `7` | `30` |
| `backupConfig.storageClass` | string | Storage class for backup PVCs | `"standard"` | `"fast-ssd"` |
| `backupConfig.storageSize` | string | Size of backup storage | `"5Gi"` | `"100Gi"` |

## Configuration Examples

### Basic PRStack

```yaml
apiVersion: shop.pilab.hu/v1alpha1
kind: PRStack
metadata:
  name: pr-123
spec:
  prNumber: "123"
  active: true
  services:
    - product-service
    - cart-service
    - order-service
    - payment-service
```

### Production Tenant with Custom Domain

```yaml
apiVersion: shop.pilab.hu/v1alpha1
kind: PRStack
metadata:
  name: production-tenant
spec:
  prNumber: "tenant1"
  imageTag: "v1.5.2"
  customDomain: "magicshop.hu"
  ingressTlsSecretName: "magicshop-tls"
  active: true
  environment: "production"
  services:
    - product-service
    - cart-service
    - order-service
    - payment-service
    - customer-service
    - inventory-service
    - notification-service
    - discount-service
    - checkout-service
    - analytics-service
    - auth-service
    - graphql-service
  resourceLimits:
    cpuLimit: "1000m"
    memoryLimit: "2Gi"
    storageLimit: "50Gi"
  backupConfig:
    enabled: true
    schedule: "0 2 * * *"  # Daily at 2 AM
    retentionDays: 30
    storageClass: "fast-ssd"
    storageSize: "100Gi"
```

### Development Environment with Custom Services

```yaml
apiVersion: shop.pilab.hu/v1alpha1
kind: PRStack
metadata:
  name: dev-environment
spec:
  prNumber: "dev"
  imageTag: "latest"
  customDomain: "dev.shop.pilab.hu"
  active: true
  environment: "development"
  services:
    - product-service
    - cart-service
    - order-service
  resourceLimits:
    cpuLimit: "250m"
    memoryLimit: "512Mi"
    storageLimit: "5Gi"
```

### Staging Environment with Custom Connections

```yaml
apiVersion: shop.pilab.hu/v1alpha1
kind: PRStack
metadata:
  name: staging-env
spec:
  prNumber: "staging"
  imageTag: "staging-v1.3.0"
  active: true
  environment: "staging"
  # Custom connection details
  mongoURI: "mongodb://staging-mongo:27017"
  mongoUsername: "staging-admin"
  mongoPassword: "staging-password"
  natsURL: "nats://staging-nats:4222"
  redisURL: "redis://staging-redis:6379"
  services:
    - product-service
    - cart-service
    - order-service
    - payment-service
  resourceLimits:
    cpuLimit: "750m"
    memoryLimit: "1.5Gi"
    storageLimit: "20Gi"
  backupConfig:
    enabled: true
    schedule: "0 */6 * * *"  # Every 6 hours
    retentionDays: 14
    storageSize: "10Gi"
```

## Deployment Configuration

### Using ConfigMaps

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: pishop-operator-config
  namespace: pishop-operator-system
data:
  BASE_DOMAIN: "your-domain.com"
  NATS_URL: "nats://nats.your-domain.com:4222"
  REDIS_URL: "redis://redis.your-domain.com:6379"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pishop-operator
  namespace: pishop-operator-system
spec:
  template:
    spec:
      containers:
      - name: manager
        envFrom:
        - configMapRef:
            name: pishop-operator-config
        env:
        - name: MONGO_URI
          valueFrom:
            secretKeyRef:
              name: mongodb-credentials
              key: uri
        - name: GITHUB_TOKEN
          valueFrom:
            secretKeyRef:
              name: github-registry-credentials
              key: token
```

### Multi-Tenant Configuration

For multi-tenant deployments across different domains:

```yaml
# Instance 1: shop.pilab.hu
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pishop-operator-shop
  namespace: pishop-operator-system
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
  namespace: pishop-operator-system
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

### Manual TLS Secret Creation

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: magicshop-tls
  namespace: default
type: kubernetes.io/tls
data:
  tls.crt: # Base64 encoded certificate
  tls.key: # Base64 encoded private key
```

Command-line creation:
```bash
kubectl create secret tls magicshop-tls \
  --cert=path/to/cert.crt \
  --key=path/to/private.key \
  --namespace=default
```

### Cert-Manager Integration

For automatic certificate management:

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: magicshop-tls
  namespace: default
spec:
  secretName: magicshop-tls
  issuerRef:
    name: letsencrypt-prod
    kind: ClusterIssuer
  dnsNames:
  - magicshop.hu
```

## Validation Rules

### PRStack Validation

The operator validates PRStack configurations:

- `prNumber` must be provided and non-empty
- `customDomain` must be a valid domain name if specified
- `ingressTlsSecretName` must reference an existing secret if specified
- `resourceLimits` values must be valid Kubernetes resource specifications
- `backupConfig.schedule` must be a valid cron expression if specified

### Operator Validation

The operator validates its own configuration at startup:

- `MONGO_URI` must be provided and accessible
- `BASE_DOMAIN` must be a valid domain name
- GitHub credentials are optional but recommended for image pulls
- External service URLs must be accessible if custom values are provided

## Configuration Best Practices

### Security
- Use Kubernetes secrets for sensitive data (passwords, tokens)
- Enable TLS for production deployments
- Use RBAC to restrict operator permissions
- Regularly rotate credentials

### Performance
- Set appropriate resource limits based on workload
- Use fast storage classes for databases
- Configure backup schedules during low-traffic periods
- Monitor resource usage and adjust limits accordingly

### Reliability
- Enable automatic backups for production environments
- Use multiple replicas for high availability
- Configure health checks and monitoring
- Test disaster recovery procedures

### Maintenance
- Use semantic versioning for image tags
- Document custom configurations
- Regularly update operator and dependencies
- Monitor logs for configuration issues

## Troubleshooting Configuration

### Common Issues

1. **Invalid base domain**: Ensure `BASE_DOMAIN` is a valid domain name without protocol or path
2. **MongoDB connection**: Verify `MONGO_URI` is accessible from the operator pod
3. **GitHub authentication**: Check that `GITHUB_TOKEN` has appropriate permissions
4. **TLS secret not found**: Ensure TLS secret exists in the same namespace as PRStack
5. **Invalid TLS secret**: Verify secret contains `tls.crt` and `tls.key` keys
6. **Custom domain DNS**: Ensure custom domain DNS points to ingress controller

### Debugging Configuration

Check operator logs for configuration validation:
```bash
kubectl logs -n pishop-operator-system deployment/pishop-operator
```

Look for startup messages indicating configuration values:
```
starting manager version=1.0.0 commit=abc123 buildDate=2024-01-01T00:00:00Z
```

## Next Steps

After configuring the operator:

1. [Create your first PRStack](Creating-PR-Stacks)
2. [Set up custom domains](Custom-Domains)
3. [Configure monitoring](Monitoring-Observability)
4. [Test backup and restore](Backup-Restore)
