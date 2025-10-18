# Tenant CRD Feature Implementation

## Overview

The PiShop Operator currently manages PR-based environments through the `PRStack` CRD. This feature request proposes implementing a new `Tenant` CRD that provides similar functionality but is designed for production customer shops with their own isolated infrastructure.

## Current State

The operator currently handles:
- **PRStack CRD**: Manages PR-based environments with isolated MongoDB, NATS, Redis
- **Lifecycle Management**: Initialization → Provisioning → Deployment → Running → Cleaning
- **Auto-scaling**: Stacks can be marked inactive (0 replicas) or active (1 replica)
- **Backup/Restore**: Automated backup creation and restoration
- **GitHub Integration**: Pulls images from GHCR with automatic secret management
- **Namespace Isolation**: Creates `pr-{number}-shop-pilab-hu` namespaces
- **Resource Management**: CPU, memory, and storage limits per environment

## Proposed Tenant CRD

### API Structure

```go
// TenantSpec defines the desired state of Tenant
type TenantSpec struct {
    // TenantID is the unique identifier for the tenant
    TenantID string `json:"tenantId"`
    
    // TenantType specifies the type of tenant (production, staging, development)
    TenantType TenantType `json:"tenantType"`
    
    // ImageTag is the Docker image tag to use for services
    ImageTag string `json:"imageTag,omitempty"`
    
    // CustomDomain is a custom domain for the ingress
    CustomDomain string `json:"customDomain,omitempty"`
    
    // IngressTlsSecretName is the name of the Kubernetes secret containing the TLS certificate
    IngressTlsSecretName string `json:"ingressTlsSecretName,omitempty"`
    
    // Active controls whether the tenant is active
    Active bool `json:"active,omitempty"`
    
    // DeployedAt triggers a rollout when changed
    DeployedAt *metav1.Time `json:"deployedAt,omitempty"`
    
    // MongoDB connection details
    MongoURI      string `json:"mongoURI,omitempty"`
    MongoUsername string `json:"mongoUsername,omitempty"`
    MongoPassword string `json:"mongoPassword,omitempty"`
    
    // NATS connection details
    NatsURL string `json:"natsURL,omitempty"`
    
    // Redis connection details
    RedisURL string `json:"redisURL,omitempty"`
    
    // Services to provision for this tenant
    Services []string `json:"services,omitempty"`
    
    // Environment configuration
    Environment string `json:"environment,omitempty"`
    
    // Resource limits for the tenant
    ResourceLimits *ResourceLimits `json:"resourceLimits,omitempty"`
    
    // Backup configuration
    BackupConfig *BackupConfig `json:"backupConfig,omitempty"`
    
    // Namespace configuration
    NamespaceConfig *NamespaceConfig `json:"namespaceConfig,omitempty"`
    
    // Tenant-specific configuration
    TenantConfig *TenantConfig `json:"tenantConfig,omitempty"`
}

type TenantType string

const (
    TenantTypeProduction  TenantType = "production"
    TenantTypeStaging     TenantType = "staging"
    TenantTypeDevelopment TenantType = "development"
)

type NamespaceConfig struct {
    // Pattern defines the namespace naming pattern
    // Default: "tenant-{tenantId}-{environment}"
    Pattern string `json:"pattern,omitempty"`
    
    // Prefix for the namespace (e.g., "shop", "pishop")
    Prefix string `json:"prefix,omitempty"`
    
    // Suffix for the namespace (e.g., "pilab-hu")
    Suffix string `json:"suffix,omitempty"`
}

type TenantConfig struct {
    // Customer information
    CustomerName string `json:"customerName,omitempty"`
    CustomerEmail string `json:"customerEmail,omitempty"`
    
    // Billing information
    BillingTier string `json:"billingTier,omitempty"`
    
    // Tenant-specific features
    Features []string `json:"features,omitempty"`
    
    // Custom labels and annotations
    Labels      map[string]string `json:"labels,omitempty"`
    Annotations map[string]string `json:"annotations,omitempty"`
}
```

## Key Features

### 1. Multi-Tenant Architecture
- **Isolated Namespaces**: Each tenant gets its own namespace (e.g., `tenant-{id}-{environment}`)
- **Resource Isolation**: Separate MongoDB databases, NATS subjects, Redis keyspaces
- **Network Isolation**: Tenant-specific ingress and service configurations
- **Security Isolation**: Separate secrets and RBAC for each tenant

### 2. Flexible Namespace Management
- **Configurable Patterns**: Support different namespace naming conventions
- **Environment Support**: Separate staging, development, and production environments
- **Custom Domains**: Support for customer-specific domains with TLS
- **Multi-Region**: Support for deploying tenants across different regions

### 3. Production-Ready Features
- **High Availability**: Support for multiple replicas and failover
- **Monitoring**: Built-in metrics and health checks
- **Backup/Restore**: Automated backup strategies for production data
- **Scaling**: Horizontal and vertical scaling capabilities
- **Upgrades**: Rolling updates and canary deployments

### 4. Customer Management
- **Customer Information**: Store customer details and billing information
- **Feature Flags**: Enable/disable features per tenant
- **Resource Limits**: Enforce resource quotas per tenant
- **Billing Integration**: Support for different billing tiers

### 5. Lifecycle Management
- **Provisioning**: Automated infrastructure setup
- **Deployment**: Service deployment and configuration
- **Monitoring**: Health checks and status reporting
- **Maintenance**: Automated updates and maintenance windows
- **Cleanup**: Graceful tenant decommissioning

## Implementation Plan

### Phase 1: Core CRD Implementation
1. **API Types**: Create `tenant_types.go` with all CRD definitions
2. **CRD Generation**: Generate Kubernetes CRD manifests
3. **Validation**: Add validation rules and webhooks
4. **Documentation**: Create API documentation and examples

### Phase 2: Controller Implementation
1. **Tenant Controller**: Create `tenant_controller.go` with reconciliation logic
2. **Provisioning**: Implement MongoDB, NATS, Redis provisioning
3. **Deployment**: Add service deployment and configuration
4. **Cleanup**: Implement tenant cleanup and resource removal

### Phase 3: Advanced Features
1. **Backup/Restore**: Implement tenant-specific backup strategies
2. **Monitoring**: Add metrics and health checks
3. **Scaling**: Implement auto-scaling and resource management
4. **Security**: Add RBAC and network policies

### Phase 4: Integration and Testing
1. **Unit Tests**: Comprehensive test coverage
2. **Integration Tests**: End-to-end testing
3. **Performance Tests**: Load and stress testing
4. **Documentation**: Complete user and operator documentation

## Example Usage

### Basic Tenant Creation

```yaml
apiVersion: shop.pilab.hu/v1alpha1
kind: Tenant
metadata:
  name: customer-shop-1
  namespace: default
spec:
  tenantId: "customer-shop-1"
  tenantType: "production"
  imageTag: "v1.2.3"
  customDomain: "shop.customer.com"
  ingressTlsSecretName: "customer-shop-tls"
  active: true
  services:
    - "api"
    - "frontend"
    - "worker"
    - "scheduler"
  environment: "production"
  resourceLimits:
    cpuLimit: "2"
    memoryLimit: "4Gi"
    storageLimit: "20Gi"
  backupConfig:
    enabled: true
    schedule: "0 2 * * *"
    retention: "30d"
  tenantConfig:
    customerName: "Customer Shop Inc."
    customerEmail: "admin@customer.com"
    billingTier: "premium"
    features:
      - "advanced-analytics"
      - "custom-themes"
      - "api-access"
```

## Benefits

1. **Production Ready**: Designed for production customer deployments
2. **Multi-Tenant**: Support for multiple customers with isolated resources
3. **Scalable**: Horizontal and vertical scaling capabilities
4. **Flexible**: Configurable namespace patterns and tenant-specific settings
5. **Secure**: Isolated resources and network policies
6. **Maintainable**: Clear separation of concerns and modular design
7. **Extensible**: Easy to add new features and tenant types

## Migration from PRStack

The Tenant CRD will coexist with PRStack during a transition period:

1. **Dual Support**: Both CRDs will be supported simultaneously
2. **Migration Tools**: Scripts to convert PRStack resources to Tenant
3. **Gradual Migration**: Customers can migrate at their own pace
4. **Backward Compatibility**: Existing PRStack resources continue to work

## Success Criteria

- [ ] Tenant CRD is fully implemented and tested
- [ ] Controller handles all lifecycle phases correctly
- [ ] Multi-tenant isolation is properly enforced
- [ ] Backup/restore functionality works for tenants
- [ ] Monitoring and health checks are implemented
- [ ] Documentation is complete and accurate
- [ ] Migration tools are available and tested
- [ ] Performance meets production requirements

## Future Enhancements

1. **Multi-Region Support**: Deploy tenants across different regions
2. **Advanced Monitoring**: Custom dashboards and alerting
3. **API Management**: Tenant-specific API rate limiting and quotas
4. **Cost Management**: Resource usage tracking and cost optimization
5. **Compliance**: GDPR, SOC2, and other compliance features
6. **Marketplace**: Self-service tenant provisioning and management
