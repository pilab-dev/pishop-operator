# PRStack to Tenant Refactoring Plan

## Overview

This document outlines the comprehensive plan to refactor the PiShop Operator from PRStack-based to Tenant-based operations, making it more suitable for both PR environments and production multi-tenant deployments.

## Current State Analysis

### Current Issues
1. **Naming Inconsistency**: API uses `PRStack` but supports tenant use cases
2. **Mixed Terminology**: Documentation mixes PR and tenant concepts
3. **Namespace Pattern**: Hardcoded `pr-{number}-shop-pilab-hu` pattern
4. **Limited Flexibility**: Designed primarily for PR environments

### Current Architecture
- **CRD**: `PRStack` with `prNumber` field
- **Namespace Pattern**: `pr-{prNumber}-shop-pilab-hu`
- **Controller**: `PRStackReconciler`
- **API Group**: `shop.pilab.hu/v1alpha1`

## Refactoring Goals

1. **Unified API**: Single CRD that works for both PR and tenant environments
2. **Flexible Naming**: Configurable namespace patterns
3. **Clear Terminology**: Consistent tenant-focused language
4. **Backward Compatibility**: Support existing PRStack resources during transition
5. **Enhanced Features**: Better support for production tenant deployments

## Phase 1: API Refactoring

### 1.1 Create New Tenant CRD

**New API Types** (`api/v1alpha1/tenant_types.go`):
```go
// TenantSpec defines the desired state of Tenant
type TenantSpec struct {
    // TenantID is the unique identifier for the tenant
    TenantID string `json:"tenantId"`
    
    // TenantType specifies the type of tenant (pr, production, staging)
    TenantType TenantType `json:"tenantType"`
    
    // PRNumber is the pull request number (only for PR tenants)
    PRNumber *string `json:"prNumber,omitempty"`
    
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
    
    // Connection details
    MongoURI      string `json:"mongoURI,omitempty"`
    MongoUsername string `json:"mongoUsername,omitempty"`
    MongoPassword string `json:"mongoPassword,omitempty"`
    NatsURL       string `json:"natsURL,omitempty"`
    RedisURL      string `json:"redisURL,omitempty"`
    
    // Services to provision
    Services []string `json:"services,omitempty"`
    
    // Environment configuration
    Environment string `json:"environment,omitempty"`
    
    // Resource limits
    ResourceLimits *ResourceLimits `json:"resourceLimits,omitempty"`
    
    // Backup configuration
    BackupConfig *BackupConfig `json:"backupConfig,omitempty"`
    
    // Namespace configuration
    NamespaceConfig *NamespaceConfig `json:"namespaceConfig,omitempty"`
}

type TenantType string

const (
    TenantTypePR         TenantType = "pr"
    TenantTypeProduction TenantType = "production"
    TenantTypeStaging    TenantType = "staging"
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
```

### 1.2 Update Status Types

```go
// TenantStatus defines the observed state of Tenant
type TenantStatus struct {
    // Phase represents the current phase
    Phase string `json:"phase,omitempty"`
    
    // Message provides additional information
    Message string `json:"message,omitempty"`
    
    // Timestamps
    CreatedAt      *metav1.Time `json:"createdAt,omitempty"`
    LastActiveAt   *metav1.Time `json:"lastActiveAt,omitempty"`
    LastDeployedAt *metav1.Time `json:"lastDeployedAt,omitempty"`
    
    // Connection details
    MongoDB *MongoDBCredentials `json:"mongodb,omitempty"`
    NATS    *NATSConfig         `json:"nats,omitempty"`
    Redis   *RedisConfig        `json:"redis,omitempty"`
    
    // Deployed services
    Services []ServiceStatus `json:"services,omitempty"`
    
    // Conditions
    Conditions []metav1.Condition `json:"conditions,omitempty"`
    
    // Backup status
    Backup *BackupStatus `json:"backup,omitempty"`
    
    // Namespace information
    Namespace string `json:"namespace,omitempty"`
}
```

## Phase 2: Controller Refactoring

### 2.1 Create Tenant Controller

**New Controller** (`controllers/tenant_controller.go`):
```go
// TenantReconciler reconciles a Tenant object
type TenantReconciler struct {
    client.Client
    Scheme         *runtime.Scheme
    Recorder       record.EventRecorder
    MongoURI       string
    MongoUsername  string
    MongoPassword  string
    BackupManager  *BackupRestoreManager
    GitHubToken    string
    GitHubUsername string
    GitHubEmail    string
    BaseDomain     string
}

// New namespace generation logic
func (r *TenantReconciler) getNamespaceName(tenant *pishopv1alpha1.Tenant) string {
    config := tenant.Spec.NamespaceConfig
    if config == nil {
        config = &pishopv1alpha1.NamespaceConfig{}
    }
    
    // Default pattern: tenant-{tenantId}-{environment}
    pattern := config.Pattern
    if pattern == "" {
        environment := tenant.Spec.Environment
        if environment == "" {
            environment = "default"
        }
        pattern = fmt.Sprintf("tenant-%s-%s", tenant.Spec.TenantID, environment)
    }
    
    // Apply prefix and suffix if specified
    if config.Prefix != "" {
        pattern = fmt.Sprintf("%s-%s", config.Prefix, pattern)
    }
    if config.Suffix != "" {
        pattern = fmt.Sprintf("%s-%s", pattern, config.Suffix)
    }
    
    return pattern
}
```

### 2.2 Update Constants

```go
const (
    // Finalizer name for Tenant resources
    FinalizerName = "shop.pilab.hu/tenant-finalizer"
    
    // Default namespace pattern
    DefaultNamespacePattern = "tenant-%s-%s"
    
    // Tenant expiration time (configurable per tenant type)
    DefaultExpirationTime = time.Hour
    
    // Phase constants
    PhaseInitialization = ""
    PhaseProvisioning   = "Provisioning"
    PhaseDeploying      = "Deploying"
    PhaseRunning        = "Running"
    PhaseInactive       = "Inactive"
    PhaseCleaning       = "Cleaning"
    PhaseCleaned        = "Cleaned"
)
```

## Phase 3: Migration Strategy

### 3.1 Backward Compatibility

1. **Keep PRStack CRD**: Maintain existing PRStack resources
2. **Create Adapter**: Create a conversion layer from PRStack to Tenant
3. **Dual Controller**: Run both controllers during transition period
4. **Migration Tool**: Provide tool to convert PRStack to Tenant

### 3.2 Migration Tool

**Migration Script** (`tools/migrate-prstack-to-tenant.go`):
```go
func MigratePRStackToTenant(prStack *pishopv1alpha1.PRStack) *pishopv1alpha1.Tenant {
    tenant := &pishopv1alpha1.Tenant{
        ObjectMeta: metav1.ObjectMeta{
            Name:      fmt.Sprintf("tenant-%s", prStack.Spec.PRNumber),
            Namespace: prStack.Namespace,
        },
        Spec: pishopv1alpha1.TenantSpec{
            TenantID:   prStack.Spec.PRNumber,
            TenantType: pishopv1alpha1.TenantTypePR,
            PRNumber:   &prStack.Spec.PRNumber,
            // ... map other fields
        },
    }
    return tenant
}
```

## Phase 4: Documentation Updates

### 4.1 Update README

- Replace "PR-based environments" with "tenant-based environments"
- Update architecture diagram to show Tenant CRD
- Add examples for both PR and production tenants
- Update installation and usage instructions

### 4.2 Update Configuration Documentation

- Document new Tenant CRD fields
- Add namespace configuration options
- Update examples for different tenant types
- Add migration guide

## Phase 5: Testing Strategy

### 5.1 Unit Tests

- Test Tenant controller logic
- Test namespace generation
- Test migration tool
- Test backward compatibility

### 5.2 Integration Tests

- Test Tenant CRD creation
- Test namespace provisioning
- Test service deployment
- Test cleanup operations

### 5.3 Migration Tests

- Test PRStack to Tenant conversion
- Test dual controller operation
- Test rollback scenarios

## Implementation Timeline

### Week 1-2: API Design and Implementation
- [ ] Create new Tenant CRD
- [ ] Implement API types
- [ ] Generate CRD manifests
- [ ] Write unit tests

### Week 3-4: Controller Implementation
- [ ] Create Tenant controller
- [ ] Implement reconciliation logic
- [ ] Add namespace generation
- [ ] Write controller tests

### Week 5-6: Migration Tools
- [ ] Create migration script
- [ ] Implement backward compatibility
- [ ] Test migration scenarios
- [ ] Create migration documentation

### Week 7-8: Documentation and Testing
- [ ] Update all documentation
- [ ] Write integration tests
- [ ] Create migration guide
- [ ] Performance testing

### Week 9-10: Deployment and Validation
- [ ] Deploy to staging
- [ ] Test with real workloads
- [ ] Validate migration tools
- [ ] Prepare production deployment

## Benefits of Refactoring

1. **Unified API**: Single CRD for all tenant types
2. **Flexible Naming**: Configurable namespace patterns
3. **Better Production Support**: Optimized for production tenants
4. **Clearer Terminology**: Consistent tenant-focused language
5. **Enhanced Features**: Better resource management and monitoring
6. **Future-Proof**: Easier to add new tenant types and features

## Risks and Mitigation

### Risks
1. **Breaking Changes**: Existing PRStack resources may not work
2. **Migration Complexity**: Converting existing resources
3. **Testing Overhead**: Need to test both old and new systems
4. **Documentation Updates**: Extensive documentation changes

### Mitigation
1. **Backward Compatibility**: Keep PRStack support during transition
2. **Migration Tools**: Provide automated migration scripts
3. **Gradual Rollout**: Deploy in stages with validation
4. **Comprehensive Testing**: Thorough testing before production

## Conclusion

This refactoring will transform the PiShop Operator from a PR-focused tool to a comprehensive tenant management platform, supporting both development and production use cases while maintaining backward compatibility during the transition period.