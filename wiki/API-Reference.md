# API Reference

Complete reference documentation for the PRStack Custom Resource Definition (CRD) and related APIs.

## PRStack Resource

The PRStack is the primary custom resource managed by the PiShop Operator. It defines a complete isolated environment for a pull request or tenant.

### Resource Definition

```yaml
apiVersion: shop.pilab.hu/v1alpha1
kind: PRStack
metadata:
  name: <stack-name>
  namespace: <namespace>
spec:
  # PRStack specification
status:
  # PRStack status (read-only)
```

### Spec Fields

#### Required Fields

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `prNumber` | string | Pull request number or tenant identifier | `"123"` |

#### Optional Fields

##### Basic Configuration

| Field | Type | Description | Default | Example |
|-------|------|-------------|---------|---------|
| `imageTag` | string | Docker image tag for services | `pr-{prNumber}` | `"v1.2.3"` |
| `customDomain` | string | Custom domain for ingress | `pr-{prNumber}.{BASE_DOMAIN}` | `"magicshop.hu"` |
| `ingressTlsSecretName` | string | TLS secret name for HTTPS | - | `"magicshop-tls"` |
| `active` | boolean | Controls whether stack is active | `true` | `true` |
| `environment` | string | Environment configuration | - | `"production"` |

##### Service Configuration

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

##### Connection Configuration

| Field | Type | Description | Default | Example |
|-------|------|-------------|---------|---------|
| `mongoURI` | string | MongoDB connection URI | Operator default | `"mongodb://custom-mongo:27017"` |
| `mongoUsername` | string | MongoDB username | Operator default | `"custom-user"` |
| `mongoPassword` | string | MongoDB password | Operator default | `"custom-pass"` |
| `natsURL` | string | NATS server URL | Operator default | `"nats://custom-nats:4222"` |
| `redisURL` | string | Redis server URL | Operator default | `"redis://custom-redis:6379"` |

##### Resource Configuration

| Field | Type | Description | Default | Example |
|-------|------|-------------|---------|---------|
| `resourceLimits.cpuLimit` | string | CPU limit per service | `"500m"` | `"1000m"` |
| `resourceLimits.memoryLimit` | string | Memory limit per service | `"1Gi"` | `"2Gi"` |
| `resourceLimits.storageLimit` | string | Storage limit for databases | `"10Gi"` | `"50Gi"` |

##### Backup Configuration

| Field | Type | Description | Default | Example |
|-------|------|-------------|---------|---------|
| `backupConfig.enabled` | boolean | Enable automatic backups | `false` | `true` |
| `backupConfig.schedule` | string | Backup schedule (cron format) | - | `"0 2 * * *"` |
| `backupConfig.retentionDays` | integer | Days to keep backups | `7` | `30` |
| `backupConfig.storageClass` | string | Storage class for backup PVCs | `"standard"` | `"fast-ssd"` |
| `backupConfig.storageSize` | string | Size of backup storage | `"5Gi"` | `"100Gi"` |

##### Rollout Configuration

| Field | Type | Description | Default | Example |
|-------|------|-------------|---------|---------|
| `deployedAt` | *metav1.Time | Timestamp that triggers rollout when changed | - | `"2025-01-15T14:30:00Z"` |

### Status Fields

The status section is read-only and managed by the operator.

#### Basic Status

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `phase` | string | Current phase of the PR stack | `"Running"` |
| `message` | string | Additional status information | `"All services deployed successfully"` |
| `createdAt` | *metav1.Time | Timestamp when stack was created | `"2024-01-15T10:30:00Z"` |
| `lastActiveAt` | *metav1.Time | Timestamp of last activity | `"2024-01-15T14:22:00Z"` |
| `lastDeployedAt` | *metav1.Time | Timestamp of last rollout | `"2024-01-15T10:35:00Z"` |

#### Service Status

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `services[].name` | string | Service name | `"product-service"` |
| `services[].status` | string | Service status | `"Running"` |
| `services[].url` | string | Service URL | `"http://product-service.pr-123-shop-pilab-hu.svc.cluster.local:8080"` |
| `services[].message` | string | Service status message | `"Deployment successful"` |

#### Database Status

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `mongodb.user` | string | Created MongoDB user | `"pr-123-user"` |
| `mongodb.connectionString` | string | MongoDB connection string | `"mongodb://pr-123-user:password@mongodb:27017/pr-123-db"` |
| `mongodb.databases` | []string | Created databases | `["pr-123-product", "pr-123-cart"]` |
| `nats.subjectPrefix` | string | NATS subject prefix | `"pr-123"` |
| `nats.connectionString` | string | NATS connection string | `"nats://nats:4222"` |
| `redis.keyPrefix` | string | Redis key prefix | `"pr-123:"` |
| `redis.connectionString` | string | Redis connection string | `"redis://redis:6379/0"` |

#### Backup Status

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `backup.lastBackupTime` | *metav1.Time | Last successful backup time | `"2024-01-15T02:00:00Z"` |
| `backup.lastBackupName` | string | Last backup name | `"backup-pr-123-20240115"` |
| `backup.backupCount` | integer | Total number of backups | `3` |
| `backup.lastBackupSize` | string | Size of last backup | `"2.1Gi"` |
| `backup.backupJobs[]` | []BackupJobStatus | Running backup/restore jobs | See below |

#### Conditions

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `conditions[].type` | string | Condition type | `"Ready"` |
| `conditions[].status` | string | Condition status | `"True"` |
| `conditions[].reason` | string | Condition reason | `"AllServicesRunning"` |
| `conditions[].message` | string | Condition message | `"All services are running successfully"` |

### BackupJobStatus

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `name` | string | Job name | `"backup-pr-123-20240115"` |
| `type` | string | Job type | `"backup"` |
| `status` | string | Job status | `"Completed"` |
| `startTime` | *metav1.Time | Job start time | `"2024-01-15T02:00:00Z"` |
| `completionTime` | *metav1.Time | Job completion time | `"2024-01-15T02:15:00Z"` |
| `message` | string | Job status message | `"Backup completed successfully"` |

## CRD Definition

### Full CRD YAML

```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: prstacks.shop.pilab.hu
spec:
  group: shop.pilab.hu
  versions:
  - name: v1alpha1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              prNumber:
                type: string
                description: Pull request number or tenant identifier
              imageTag:
                type: string
                description: Docker image tag for services
              customDomain:
                type: string
                description: Custom domain for ingress
              ingressTlsSecretName:
                type: string
                description: TLS secret name for HTTPS
              active:
                type: boolean
                default: true
                description: Controls whether stack is active
              deployedAt:
                type: string
                format: date-time
                description: Timestamp that triggers rollout when changed
              mongoURI:
                type: string
                description: MongoDB connection URI
              mongoUsername:
                type: string
                description: MongoDB username
              mongoPassword:
                type: string
                description: MongoDB password
              natsURL:
                type: string
                description: NATS server URL
              redisURL:
                type: string
                description: Redis server URL
              services:
                type: array
                items:
                  type: string
                description: List of services to deploy
              environment:
                type: string
                description: Environment configuration
              resourceLimits:
                type: object
                properties:
                  cpuLimit:
                    type: string
                    description: CPU limit per service
                  memoryLimit:
                    type: string
                    description: Memory limit per service
                  storageLimit:
                    type: string
                    description: Storage limit for databases
              backupConfig:
                type: object
                properties:
                  enabled:
                    type: boolean
                    description: Enable automatic backups
                  schedule:
                    type: string
                    description: Backup schedule in cron format
                  retentionDays:
                    type: integer
                    description: Days to keep backups
                  storageClass:
                    type: string
                    description: Storage class for backup PVCs
                  storageSize:
                    type: string
                    description: Size of backup storage
            required:
            - prNumber
          status:
            type: object
            properties:
              phase:
                type: string
                description: Current phase of the PR stack
              message:
                type: string
                description: Additional status information
              createdAt:
                type: string
                format: date-time
                description: Timestamp when stack was created
              lastActiveAt:
                type: string
                format: date-time
                description: Timestamp of last activity
              lastDeployedAt:
                type: string
                format: date-time
                description: Timestamp of last rollout
              mongodb:
                type: object
                properties:
                  user:
                    type: string
                  connectionString:
                    type: string
                  databases:
                    type: array
                    items:
                      type: string
              nats:
                type: object
                properties:
                  subjectPrefix:
                    type: string
                  connectionString:
                    type: string
              redis:
                type: object
                properties:
                  keyPrefix:
                    type: string
                  connectionString:
                    type: string
              services:
                type: array
                items:
                  type: object
                  properties:
                    name:
                      type: string
                    status:
                      type: string
                    url:
                      type: string
                    message:
                      type: string
              conditions:
                type: array
                items:
                  type: object
                  properties:
                    type:
                      type: string
                    status:
                      type: string
                    reason:
                      type: string
                    message:
                      type: string
                    lastTransitionTime:
                      type: string
                      format: date-time
              backup:
                type: object
                properties:
                  lastBackupTime:
                    type: string
                    format: date-time
                  lastBackupName:
                    type: string
                  backupCount:
                    type: integer
                  lastBackupSize:
                    type: string
                  backupJobs:
                    type: array
                    items:
                      type: object
                      properties:
                        name:
                          type: string
                        type:
                          type: string
                        status:
                          type: string
                        startTime:
                          type: string
                          format: date-time
                        completionTime:
                          type: string
                          format: date-time
                        message:
                          type: string
    subresources:
      status: {}
  scope: Cluster
  names:
    plural: prstacks
    singular: prstack
    kind: PRStack
    shortNames:
    - prs
    categories:
    - all
```

## API Operations

### Create PRStack

```bash
kubectl apply -f - <<EOF
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
EOF
```

### Get PRStack

```bash
# Get PRStack
kubectl get prstack pr-123

# Get PRStack with details
kubectl describe prstack pr-123

# Get PRStack YAML
kubectl get prstack pr-123 -o yaml
```

### Update PRStack

```bash
# Update using patch
kubectl patch prstack pr-123 --type=merge -p '{"spec":{"active":false}}'

# Update using edit
kubectl edit prstack pr-123

# Update using apply
kubectl apply -f updated-prstack.yaml
```

### Delete PRStack

```bash
# Delete PRStack
kubectl delete prstack pr-123

# Delete with force
kubectl delete prstack pr-123 --force --grace-period=0
```

## Validation Rules

### Required Fields

- `prNumber` must be provided and non-empty
- `prNumber` must be a valid string identifier

### Optional Field Validation

- `customDomain` must be a valid domain name if specified
- `ingressTlsSecretName` must reference an existing secret if specified
- `resourceLimits` values must be valid Kubernetes resource specifications
- `backupConfig.schedule` must be a valid cron expression if specified
- `deployedAt` must be a valid RFC3339 timestamp if specified

### Resource Limits Validation

- CPU limits must be valid Kubernetes CPU units (e.g., "100m", "1", "2")
- Memory limits must be valid Kubernetes memory units (e.g., "128Mi", "1Gi", "2Gi")
- Storage limits must be valid Kubernetes storage units (e.g., "1Gi", "10Gi", "100Gi")

## Status Phases

The PRStack goes through several phases during its lifecycle:

### Phase Values

| Phase | Description | Next Phases |
|-------|-------------|-------------|
| `Pending` | PRStack created, waiting for processing | `Provisioning`, `Failed` |
| `Provisioning` | Creating resources (databases, services) | `Running`, `Failed` |
| `Running` | All services deployed and running | `Stopping`, `Failed` |
| `Stopping` | Shutting down services | `Stopped`, `Failed` |
| `Stopped` | All services stopped | `Running`, `Deleting` |
| `Failed` | Error occurred, requires intervention | `Pending`, `Deleting` |
| `Deleting` | Resources being cleaned up | - |

### Phase Transitions

```mermaid
graph TD
    A[Pending] --> B[Provisioning]
    A --> C[Failed]
    B --> D[Running]
    B --> C
    D --> E[Stopping]
    D --> C
    E --> F[Stopped]
    E --> C
    F --> D
    F --> G[Deleting]
    C --> A
    C --> G
    G --> H[Deleted]
```

## Conditions

The PRStack status includes conditions that provide detailed information about the current state:

### Condition Types

| Type | Description | Possible Values |
|------|-------------|-----------------|
| `Ready` | Overall readiness of the PRStack | `True`, `False`, `Unknown` |
| `ServicesReady` | All services are running | `True`, `False`, `Unknown` |
| `DatabaseReady` | Database connections are ready | `True`, `False`, `Unknown` |
| `IngressReady` | Ingress is configured and ready | `True`, `False`, `Unknown` |
| `BackupReady` | Backup configuration is ready | `True`, `False`, `Unknown` |

### Condition Status Values

- `True`: Condition is satisfied
- `False`: Condition is not satisfied
- `Unknown`: Condition status is unknown

## Events

The operator generates events for important operations:

### Event Types

| Type | Reason | Description |
|------|--------|-------------|
| `Normal` | `Created` | PRStack created successfully |
| `Normal` | `ProvisioningStarted` | Started provisioning resources |
| `Normal` | `ProvisioningCompleted` | Resource provisioning completed |
| `Normal` | `ServicesDeployed` | Services deployed successfully |
| `Normal` | `RolloutTriggered` | Deployment rollout triggered |
| `Normal` | `BackupCreated` | Backup created successfully |
| `Normal` | `RestoreCompleted` | Restore operation completed |
| `Warning` | `ProvisioningFailed` | Resource provisioning failed |
| `Warning` | `ServiceDeploymentFailed` | Service deployment failed |
| `Warning` | `RolloutFailed` | Deployment rollout failed |
| `Warning` | `BackupFailed` | Backup operation failed |
| `Warning` | `RestoreFailed` | Restore operation failed |

## Metrics

The operator exposes Prometheus metrics for monitoring:

### Operator Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `pishop_operator_prstack_total` | Counter | Total number of PRStacks |
| `pishop_operator_prstack_phase` | Gauge | Number of PRStacks per phase |
| `pishop_operator_reconcile_total` | Counter | Total number of reconciles |
| `pishop_operator_reconcile_duration_seconds` | Histogram | Duration of reconcile operations |
| `pishop_operator_backup_total` | Counter | Total number of backups |
| `pishop_operator_backup_duration_seconds` | Histogram | Duration of backup operations |

### Service Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `pishop_operator_service_deployments_total` | Counter | Total number of service deployments |
| `pishop_operator_service_deployments_ready` | Gauge | Number of ready service deployments |
| `pishop_operator_service_deployments_failed` | Gauge | Number of failed service deployments |

## Webhooks

The operator supports admission webhooks for validation:

### Validation Webhook

- **Path**: `/validate-shop-pilab-hu-v1alpha1-prstack`
- **Operations**: CREATE, UPDATE, DELETE
- **Purpose**: Validate PRStack specifications before admission

### Mutation Webhook

- **Path**: `/mutate-shop-pilab-hu-v1alpha1-prstack`
- **Operations**: CREATE, UPDATE
- **Purpose**: Set default values and mutate PRStack specifications

## RBAC Requirements

The operator requires the following RBAC permissions:

### ClusterRole Permissions

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: pishop-operator-role
rules:
- apiGroups: [""]
  resources: ["pods", "services", "endpoints", "persistentvolumeclaims", "events", "configmaps", "secrets", "namespaces"]
  verbs: ["*"]
- apiGroups: ["apps"]
  resources: ["deployments", "daemonsets", "replicasets", "statefulsets"]
  verbs: ["*"]
- apiGroups: ["networking.k8s.io"]
  resources: ["ingresses"]
  verbs: ["*"]
- apiGroups: ["shop.pilab.hu"]
  resources: ["prstacks"]
  verbs: ["*"]
- apiGroups: ["shop.pilab.hu"]
  resources: ["prstacks/status"]
  verbs: ["*"]
```

## Next Steps

After understanding the API:

1. [Create your first PRStack](Creating-PR-Stacks)
2. [Learn about configuration options](Configuration)
3. [Set up monitoring](Monitoring-Observability)
4. [Explore advanced features](Advanced-Features)
