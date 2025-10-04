# Advanced Features

This guide covers advanced features and configurations for the PiShop Operator, including custom domains, backup strategies, monitoring, and production deployment patterns.

## Custom Domains and TLS

### Setting Up Custom Domains

Custom domains allow you to use your own domain names instead of the default `pr-{number}.shop.pilab.hu` pattern.

#### Basic Custom Domain

```yaml
apiVersion: shop.pilab.hu/v1alpha1
kind: PRStack
metadata:
  name: production-tenant
spec:
  prNumber: "tenant1"
  customDomain: "magicshop.hu"
  active: true
  services:
    - product-service
    - cart-service
    - order-service
    - payment-service
```

#### Custom Domain with TLS

```yaml
apiVersion: shop.pilab.hu/v1alpha1
kind: PRStack
metadata:
  name: production-tenant
spec:
  prNumber: "tenant1"
  customDomain: "magicshop.hu"
  ingressTlsSecretName: "magicshop-tls"
  active: true
  services:
    - product-service
    - cart-service
    - order-service
    - payment-service
```

### TLS Certificate Management

#### Manual TLS Secret Creation

```bash
# Create TLS secret from certificate files
kubectl create secret tls magicshop-tls \
  --cert=path/to/cert.crt \
  --key=path/to/private.key \
  --namespace=default
```

#### Using Cert-Manager for Automatic Certificates

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
  - www.magicshop.hu
```

#### Let's Encrypt with DNS Challenge

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-dns
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@magicshop.hu
    privateKeySecretRef:
      name: letsencrypt-dns
    solvers:
    - dns01:
        cloudflare:
          email: admin@magicshop.hu
          apiTokenSecretRef:
            name: cloudflare-api-token
            key: api-token
```

### DNS Configuration

#### A Record Configuration

```bash
# Point domain to ingress controller IP
dig magicshop.hu
# Should return the ingress controller external IP

# Verify DNS propagation
nslookup magicshop.hu 8.8.8.8
```

#### CNAME Configuration

```bash
# For subdomains, use CNAME
# www.magicshop.hu CNAME magicshop.hu
```

## Backup and Restore Strategies

### Automated Backup Configuration

#### Daily Backups

```yaml
apiVersion: shop.pilab.hu/v1alpha1
kind: PRStack
metadata:
  name: production-tenant
spec:
  prNumber: "tenant1"
  backupConfig:
    enabled: true
    schedule: "0 2 * * *"  # Daily at 2 AM
    retentionDays: 30
    storageClass: "fast-ssd"
    storageSize: "100Gi"
```

#### Multiple Backup Schedules

```yaml
apiVersion: shop.pilab.hu/v1alpha1
kind: PRStack
metadata:
  name: production-tenant
spec:
  prNumber: "tenant1"
  backupConfig:
    enabled: true
    schedule: "0 */6 * * *"  # Every 6 hours
    retentionDays: 7
    storageClass: "fast-ssd"
    storageSize: "50Gi"
```

### Manual Backup Operations

#### Create One-Time Backup

```bash
# Trigger manual backup
kubectl patch prstack production-tenant --type=merge -p '{
  "spec": {
    "backupConfig": {
      "enabled": true,
      "schedule": "manual-backup-'$(date +%s)'"
    }
  }
}'
```

#### Restore from Backup

```bash
# List available backups
kubectl get pvc -n pr-tenant1-shop-pilab-hu | grep backup

# Trigger restore (implementation depends on your restore strategy)
kubectl patch prstack production-tenant --type=merge -p '{
  "spec": {
    "restoreFromBackup": "backup-pr-tenant1-20240115"
  }
}'
```

### Backup Storage Strategies

#### Local Storage

```yaml
apiVersion: v1
kind: StorageClass
metadata:
  name: fast-ssd
provisioner: kubernetes.io/no-provisioner
volumeBindingMode: WaitForFirstConsumer
```

#### Cloud Storage

```yaml
apiVersion: v1
kind: StorageClass
metadata:
  name: aws-ebs-fast
provisioner: ebs.csi.aws.com
parameters:
  type: gp3
  iops: "4000"
  throughput: "250"
```

## Monitoring and Observability

### Prometheus Metrics

The operator exposes metrics on port 8080:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: pishop-operator-metrics
  namespace: pishop-operator-system
spec:
  ports:
  - name: metrics
    port: 8080
    targetPort: 8080
  selector:
    control-plane: controller-manager
```

#### ServiceMonitor for Prometheus

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: pishop-operator
  namespace: pishop-operator-system
spec:
  selector:
    matchLabels:
      name: pishop-operator-metrics
  endpoints:
  - port: metrics
    interval: 30s
```

### Grafana Dashboards

#### Operator Dashboard

```json
{
  "dashboard": {
    "title": "PiShop Operator",
    "panels": [
      {
        "title": "PRStack Count",
        "type": "stat",
        "targets": [
          {
            "expr": "count(pishop_operator_prstack_total)",
            "legendFormat": "Total PRStacks"
          }
        ]
      },
      {
        "title": "Reconcile Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(pishop_operator_reconcile_total[5m])",
            "legendFormat": "Reconciles/sec"
          }
        ]
      }
    ]
  }
}
```

### Alerting Rules

#### Prometheus Alerting Rules

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: pishop-operator-alerts
  namespace: pishop-operator-system
spec:
  groups:
  - name: pishop-operator
    rules:
    - alert: PRStackStuck
      expr: pishop_operator_prstack_phase{phase="Pending"} > 0
      for: 5m
      labels:
        severity: warning
      annotations:
        summary: "PRStack stuck in Pending phase"
        description: "PRStack {{ $labels.name }} has been in Pending phase for more than 5 minutes"
    
    - alert: OperatorDown
      expr: up{job="pishop-operator"} == 0
      for: 1m
      labels:
        severity: critical
      annotations:
        summary: "PiShop Operator is down"
        description: "PiShop Operator has been down for more than 1 minute"
```

## Multi-Tenant Deployment

### Production Tenant Management

#### Tenant Isolation

```yaml
# Tenant 1
apiVersion: shop.pilab.hu/v1alpha1
kind: PRStack
metadata:
  name: tenant1-production
spec:
  prNumber: "tenant1"
  customDomain: "tenant1.shop.pilab.hu"
  ingressTlsSecretName: "tenant1-tls"
  active: true
  environment: "production"
  services:
    - product-service
    - cart-service
    - order-service
    - payment-service
  resourceLimits:
    cpuLimit: "2000m"
    memoryLimit: "4Gi"
    storageLimit: "100Gi"

---
# Tenant 2
apiVersion: shop.pilab.hu/v1alpha1
kind: PRStack
metadata:
  name: tenant2-production
spec:
  prNumber: "tenant2"
  customDomain: "tenant2.shop.pilab.hu"
  ingressTlsSecretName: "tenant2-tls"
  active: true
  environment: "production"
  services:
    - product-service
    - cart-service
    - order-service
    - payment-service
  resourceLimits:
    cpuLimit: "2000m"
    memoryLimit: "4Gi"
    storageLimit: "100Gi"
```

### Resource Quotas

#### Namespace Resource Quotas

```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: tenant1-quota
  namespace: pr-tenant1-shop-pilab-hu
spec:
  hard:
    requests.cpu: "2"
    requests.memory: 4Gi
    limits.cpu: "4"
    limits.memory: 8Gi
    persistentvolumeclaims: "10"
    pods: "20"
    services: "10"
```

#### Limit Ranges

```yaml
apiVersion: v1
kind: LimitRange
metadata:
  name: tenant1-limits
  namespace: pr-tenant1-shop-pilab-hu
spec:
  limits:
  - default:
      cpu: "500m"
      memory: "1Gi"
    defaultRequest:
      cpu: "100m"
      memory: "256Mi"
    type: Container
```

## Service Mesh Integration

### Istio Integration

#### Enable Istio Sidecar Injection

```yaml
apiVersion: shop.pilab.hu/v1alpha1
kind: PRStack
metadata:
  name: pr-123
  annotations:
    sidecar.istio.io/inject: "true"
spec:
  prNumber: "123"
  active: true
  services:
    - product-service
    - cart-service
    - order-service
    - payment-service
```

#### Istio Gateway Configuration

```yaml
apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: pr-123-gateway
  namespace: pr-123-shop-pilab-hu
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      number: 80
      name: http
      protocol: HTTP
    hosts:
    - pr-123.shop.pilab.hu
  - port:
      number: 443
      name: https
      protocol: HTTPS
    tls:
      mode: SIMPLE
      credentialName: pr-123-tls
    hosts:
    - pr-123.shop.pilab.hu
```

#### Virtual Service

```yaml
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: pr-123-vs
  namespace: pr-123-shop-pilab-hu
spec:
  hosts:
  - pr-123.shop.pilab.hu
  gateways:
  - pr-123-gateway
  http:
  - match:
    - uri:
        prefix: /api/products
    route:
    - destination:
        host: product-service
        port:
          number: 8080
  - match:
    - uri:
        prefix: /api/cart
    route:
    - destination:
        host: cart-service
        port:
          number: 8080
```

## Security Hardening

### Network Policies

#### Default Deny All

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-all
  namespace: pr-123-shop-pilab-hu
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
```

#### Allow Internal Communication

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-internal
  namespace: pr-123-shop-pilab-hu
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: pr-123-shop-pilab-hu
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: pr-123-shop-pilab-hu
  - to: []  # Allow egress to external services
    ports:
    - protocol: TCP
      port: 53
    - protocol: UDP
      port: 53
```

### Pod Security Standards

#### Pod Security Policy

```yaml
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: pishop-operator-psp
spec:
  privileged: false
  allowPrivilegeEscalation: false
  requiredDropCapabilities:
    - ALL
  volumes:
    - 'configMap'
    - 'emptyDir'
    - 'projected'
    - 'secret'
    - 'downwardAPI'
    - 'persistentVolumeClaim'
  runAsUser:
    rule: 'MustRunAsNonRoot'
  seLinux:
    rule: 'RunAsAny'
  fsGroup:
    rule: 'RunAsAny'
```

### RBAC Configuration

#### Operator RBAC

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: pishop-operator-role
rules:
- apiGroups: [""]
  resources: ["pods", "services", "endpoints", "persistentvolumeclaims", "events", "configmaps", "secrets"]
  verbs: ["*"]
- apiGroups: ["apps"]
  resources: ["deployments", "daemonsets", "replicasets", "statefulsets"]
  verbs: ["*"]
- apiGroups: ["monitoring.coreos.com"]
  resources: ["servicemonitors"]
  verbs: ["get", "create"]
- apiGroups: ["shop.pilab.hu"]
  resources: ["prstacks"]
  verbs: ["*"]
- apiGroups: ["shop.pilab.hu"]
  resources: ["prstacks/status"]
  verbs: ["*"]
```

## Performance Optimization

### Resource Optimization

#### Horizontal Pod Autoscaling

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: product-service-hpa
  namespace: pr-123-shop-pilab-hu
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: product-service
  minReplicas: 1
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

#### Vertical Pod Autoscaling

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: product-service-vpa
  namespace: pr-123-shop-pilab-hu
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: product-service
  updatePolicy:
    updateMode: "Auto"
  resourcePolicy:
    containerPolicies:
    - containerName: product-service
      minAllowed:
        cpu: 100m
        memory: 128Mi
      maxAllowed:
        cpu: 1000m
        memory: 2Gi
```

### Caching Strategies

#### Redis Configuration

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config
  namespace: pr-123-shop-pilab-hu
data:
  redis.conf: |
    maxmemory 512mb
    maxmemory-policy allkeys-lru
    save 900 1
    save 300 10
    save 60 10000
```

#### MongoDB Optimization

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: mongodb-config
  namespace: pr-123-shop-pilab-hu
data:
  mongod.conf: |
    storage:
      wiredTiger:
        engineConfig:
          cacheSizeGB: 1
    operationProfiling:
      slowOpThresholdMs: 100
      mode: slowOp
```

## Disaster Recovery

### Backup Strategies

#### Cross-Region Backup

```yaml
apiVersion: shop.pilab.hu/v1alpha1
kind: PRStack
metadata:
  name: production-tenant
spec:
  prNumber: "tenant1"
  backupConfig:
    enabled: true
    schedule: "0 2 * * *"
    retentionDays: 30
    storageClass: "cross-region-backup"
    storageSize: "100Gi"
    backupLocation: "s3://backup-bucket/tenant1"
```

#### Point-in-Time Recovery

```yaml
apiVersion: shop.pilab.hu/v1alpha1
kind: PRStack
metadata:
  name: production-tenant
spec:
  prNumber: "tenant1"
  backupConfig:
    enabled: true
    schedule: "0 */6 * * *"  # Every 6 hours
    retentionDays: 30
    pointInTimeRecovery: true
    backupLocation: "s3://backup-bucket/tenant1"
```

### Failover Procedures

#### Active-Passive Setup

```yaml
# Primary region
apiVersion: shop.pilab.hu/v1alpha1
kind: PRStack
metadata:
  name: tenant1-primary
spec:
  prNumber: "tenant1"
  customDomain: "magicshop.hu"
  active: true
  environment: "production"
  backupConfig:
    enabled: true
    schedule: "0 2 * * *"
    crossRegionReplication: true
    replicationTarget: "secondary-region"

---
# Secondary region
apiVersion: shop.pilab.hu/v1alpha1
kind: PRStack
metadata:
  name: tenant1-secondary
spec:
  prNumber: "tenant1"
  customDomain: "magicshop-backup.hu"
  active: false
  environment: "production"
  backupConfig:
    enabled: true
    restoreFromBackup: "latest"
```

## Next Steps

After implementing advanced features:

1. [Set up monitoring and alerting](Monitoring-Observability)
2. [Implement disaster recovery procedures](Disaster-Recovery)
3. [Optimize performance based on metrics](Performance-Tuning)
4. [Review security configurations](Security)
5. [Document custom configurations](Documentation)
