# Creating PR Stacks

This guide shows you how to create and manage PRStack resources for different scenarios, from simple PR environments to production tenants.

## Basic PRStack Creation

### Simple PR Environment

Create a basic PR environment for pull request #123:

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

Apply the configuration:
```bash
kubectl apply -f pr-123.yaml
```

This creates:
- Namespace: `pr-123-shop-pilab-hu`
- Domain: `pr-123.shop.pilab.hu`
- MongoDB database and user for PR #123
- NATS subject isolation with prefix `pr-123`
- Redis keyspace isolation with prefix `pr-123:`
- Deployments for the specified services

### Check Status

```bash
# View the PRStack
kubectl get prstack pr-123

# Get detailed information
kubectl describe prstack pr-123

# Check all resources in the PR namespace
kubectl get all -n pr-123-shop-pilab-hu
```

## Advanced PRStack Examples

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

### Development Environment

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

## PRStack Management

### Activating/Deactivating Stacks

#### Deactivate a Stack (Scale to 0)
```bash
kubectl patch prstack pr-123 --type=merge -p '{"spec":{"active":false}}'
```

#### Activate a Stack (Scale to 1)
```bash
kubectl patch prstack pr-123 --type=merge -p '{"spec":{"active":true}}'
```

#### Reactivation Workaround
If a stack has expired, use this workaround to reactivate it:

```bash
# Set active=true AND update lastActiveAt in one operation
kubectl patch prstack pr-123 --type=merge -p "{
  \"spec\": {
    \"active\": true
  },
  \"status\": {
    \"lastActiveAt\": \"$(date -u +"%Y-%m-%dT%H:%M:%SZ")\"
  }
}"
```

### Triggering Deployment Rollouts

#### Rollout All Services
```bash
# Get current timestamp
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Trigger rollout
kubectl patch prstack pr-123 --type=merge -p "{\"spec\":{\"deployedAt\":\"$TIMESTAMP\"}}"
```

#### Update Image Tag and Rollout
```bash
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
kubectl patch prstack pr-123 --type=merge -p "{
  \"spec\": {
    \"imageTag\": \"pr-123-abc123\",
    \"deployedAt\": \"$TIMESTAMP\"
  }
}"
```

### Updating Service List

```bash
# Add services to existing stack
kubectl patch prstack pr-123 --type=merge -p '{
  "spec": {
    "services": ["product-service", "cart-service", "order-service", "payment-service", "customer-service"]
  }
}'
```

### Updating Resource Limits

```bash
# Update resource limits
kubectl patch prstack pr-123 --type=merge -p '{
  "spec": {
    "resourceLimits": {
      "cpuLimit": "1000m",
      "memoryLimit": "2Gi",
      "storageLimit": "20Gi"
    }
  }
}'
```

### Enabling/Disabling Backups

```bash
# Enable backups
kubectl patch prstack pr-123 --type=merge -p '{
  "spec": {
    "backupConfig": {
      "enabled": true,
      "schedule": "0 2 * * *",
      "retentionDays": 7,
      "storageSize": "5Gi"
    }
  }
}'

# Disable backups
kubectl patch prstack pr-123 --type=merge -p '{
  "spec": {
    "backupConfig": {
      "enabled": false
    }
  }
}'
```

## Monitoring PRStack Status

### Check PRStack Status

```bash
# List all PRStacks
kubectl get prstacks

# Get detailed status
kubectl get prstack pr-123 -o yaml

# Watch status changes
kubectl get prstacks -w
```

### View Status Information

The PRStack status includes comprehensive information:

```yaml
status:
  phase: "Running"
  message: "All services deployed successfully"
  createdAt: "2024-01-15T10:30:00Z"
  lastActiveAt: "2024-01-15T14:22:00Z"
  lastDeployedAt: "2024-01-15T10:35:00Z"
  mongodb:
    user: "pr-123-user"
    connectionString: "mongodb://pr-123-user:password@mongodb:27017/pr-123-db"
    databases: ["pr-123-product", "pr-123-cart", "pr-123-order"]
  nats:
    subjectPrefix: "pr-123"
    connectionString: "nats://nats:4222"
  redis:
    keyPrefix: "pr-123:"
    connectionString: "redis://redis:6379/0"
  services:
    - name: "product-service"
      status: "Running"
      url: "http://product-service.pr-123-shop-pilab-hu.svc.cluster.local:8080"
    - name: "cart-service"
      status: "Running"
      url: "http://cart-service.pr-123-shop-pilab-hu.svc.cluster.local:8080"
  backup:
    lastBackupTime: "2024-01-15T02:00:00Z"
    lastBackupName: "backup-pr-123-20240115"
    backupCount: 3
    lastBackupSize: "2.1Gi"
```

### Check Service Health

```bash
# Check all deployments in PR namespace
kubectl get deployments -n pr-123-shop-pilab-hu

# Check pod status
kubectl get pods -n pr-123-shop-pilab-hu

# Check service endpoints
kubectl get endpoints -n pr-123-shop-pilab-hu

# Check ingress
kubectl get ingress -n pr-123-shop-pilab-hu
```

### View Logs

```bash
# Check operator logs
kubectl logs -n pishop-operator-system deployment/pishop-operator

# Check service logs
kubectl logs -n pr-123-shop-pilab-hu deployment/product-service

# Follow logs in real-time
kubectl logs -f -n pr-123-shop-pilab-hu deployment/product-service
```

## Common Operations

### Copy PRStack Configuration

```bash
# Export PRStack to file
kubectl get prstack pr-123 -o yaml > pr-123-backup.yaml

# Remove status and metadata for reuse
kubectl get prstack pr-123 -o yaml | \
  grep -v "status:" -A 1000 | \
  grep -v "resourceVersion\|uid\|creationTimestamp" > pr-123-template.yaml
```

### Bulk Operations

#### Create Multiple PR Stacks

```bash
# Create PR stacks for multiple PRs
for pr in 123 124 125; do
  kubectl apply -f - <<EOF
apiVersion: shop.pilab.hu/v1alpha1
kind: PRStack
metadata:
  name: pr-$pr
spec:
  prNumber: "$pr"
  active: true
  services:
    - product-service
    - cart-service
    - order-service
    - payment-service
EOF
done
```

#### Activate Multiple Stacks

```bash
# Activate multiple stacks
for pr in 123 124 125; do
  kubectl patch prstack pr-$pr --type=merge -p '{"spec":{"active":true}}'
done
```

#### Delete Multiple Stacks

```bash
# Delete multiple stacks
for pr in 123 124 125; do
  kubectl delete prstack pr-$pr
done
```

### Cleanup Operations

#### Delete PRStack and Namespace

```bash
# Delete PRStack (this will also clean up the namespace)
kubectl delete prstack pr-123
```

#### Manual Cleanup (if needed)

```bash
# Delete namespace manually
kubectl delete namespace pr-123-shop-pilab-hu

# Clean up MongoDB databases
kubectl run mongodb-cleanup --image=mongo:latest --rm -it --restart=Never -- \
  mongo "mongodb://admin:password@mongodb:27017/admin" --eval "
    db.adminCommand({dropUser: 'pr-123-user'});
    db.adminCommand({dropDatabase: 'pr-123-product'});
    db.adminCommand({dropDatabase: 'pr-123-cart'});
    db.adminCommand({dropDatabase: 'pr-123-order'});
  "
```

## Troubleshooting PRStack Creation

### Common Issues

#### PRStack Stuck in Pending Phase

```bash
# Check operator logs
kubectl logs -n pishop-operator-system deployment/pishop-operator

# Check PRStack events
kubectl describe prstack pr-123 | grep -A 10 Events

# Check if namespace was created
kubectl get namespace pr-123-shop-pilab-hu
```

#### Services Not Starting

```bash
# Check deployment status
kubectl get deployments -n pr-123-shop-pilab-hu

# Check pod status and logs
kubectl get pods -n pr-123-shop-pilab-hu
kubectl logs -n pr-123-shop-pilab-hu deployment/product-service

# Check image pull issues
kubectl describe pod -n pr-123-shop-pilab-hu -l app=product-service
```

#### Database Connection Issues

```bash
# Check MongoDB credentials
kubectl get secret mongodb-credentials -n pishop-operator-system

# Test MongoDB connectivity
kubectl run mongodb-test --image=mongo:latest --rm -it --restart=Never -- \
  mongo "mongodb://admin:password@mongodb:27017/admin"

# Check MongoDB user creation
kubectl logs -n pishop-operator-system deployment/pishop-operator | grep -i mongo
```

#### Custom Domain Issues

```bash
# Check TLS secret
kubectl get secret magicshop-tls -n default

# Check ingress configuration
kubectl get ingress -n pr-123-shop-pilab-hu -o yaml

# Check DNS resolution
nslookup magicshop.hu
```

## Best Practices

### Naming Conventions

- Use descriptive names: `pr-123`, `production-tenant`, `staging-env`
- Include environment in name: `pr-123-dev`, `pr-123-staging`
- Use consistent naming across related resources

### Resource Management

- Set appropriate resource limits based on workload
- Monitor resource usage and adjust limits
- Use different limits for different environments

### Security

- Use TLS for production environments
- Rotate credentials regularly
- Follow principle of least privilege

### Monitoring

- Monitor PRStack status regularly
- Set up alerts for failed deployments
- Track resource usage and costs

## Next Steps

After creating PRStacks:

1. [Set up monitoring](Monitoring-Observability)
2. [Configure custom domains](Custom-Domains)
3. [Set up backup and restore](Backup-Restore)
4. [Learn about advanced features](Advanced-Features)
