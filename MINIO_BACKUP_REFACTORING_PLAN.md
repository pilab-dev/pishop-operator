# MinIO Backup System Refactoring Plan

## Overview

This document outlines the comprehensive plan to refactor the PiShop Operator's backup system from local PVC storage to MinIO object storage, providing persistent, scalable, and secure backup capabilities.

## Current State Analysis

### Current Issues
1. **Local Storage Only**: Backups stored in local PVCs (`/backup` mount)
2. **No Persistence**: Backups lost when PVCs are deleted
3. **Placeholder Functions**: `ListBackups()` and `CleanupOldBackups()` are not implemented
4. **No Cross-Cluster Access**: Backups can't be accessed from other clusters
5. **No Encryption**: Backups stored in plain text
6. **Limited Scalability**: PVC-based storage doesn't scale well

### Current Architecture
- **Storage**: Local PVC with `/backup` mount
- **Backup Jobs**: Kubernetes Jobs with MongoDB tools
- **Retention**: Not implemented
- **Listing**: Placeholder implementation
- **Cleanup**: Not implemented

## Refactoring Goals

1. **MinIO Integration**: Use MinIO for object storage
2. **Operator Secrets**: Store MinIO credentials in operator secrets
3. **Encryption**: Encrypt backups before storage
4. **Retention Policies**: Implement proper backup retention
5. **Cross-Cluster Access**: Enable backup sharing between clusters
6. **Monitoring**: Add backup status and metrics
7. **Disaster Recovery**: Support backup restoration from any cluster

## Phase 1: MinIO Integration

### 1.1 Update BackupRestoreManager

**Enhanced BackupRestoreManager** (`controllers/backup_restore.go`):
```go
// BackupRestoreManager handles database backup and restore operations with MinIO
type BackupRestoreManager struct {
    client.Client
    MongoURI      string
    MongoUsername string
    MongoPassword string
    MinIOClient   *minio.Client
    MinIOConfig   MinIOConfig
}

type MinIOConfig struct {
    Endpoint        string
    AccessKeyID     string
    SecretAccessKey string
    UseSSL          bool
    BucketName      string
    Region          string
}

// New MinIO client initialization
func NewBackupRestoreManager(client client.Client, mongoURI, mongoUsername, mongoPassword string, minioConfig MinIOConfig) (*BackupRestoreManager, error) {
    minioClient, err := minio.New(minioConfig.Endpoint, &minio.Options{
        Creds:  credentials.NewStaticV4(minioConfig.AccessKeyID, minioConfig.SecretAccessKey, ""),
        Secure: minioConfig.UseSSL,
        Region: minioConfig.Region,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create MinIO client: %v", err)
    }

    return &BackupRestoreManager{
        Client:        client,
        MongoURI:      mongoURI,
        MongoUsername: mongoUsername,
        MongoPassword: mongoPassword,
        MinIOClient:   minioClient,
        MinIOConfig:   minioConfig,
    }, nil
}
```

### 1.2 Add MinIO Secret Management

**New Secret Management** (`controllers/minio_secret.go`):
```go
// MinIOSecretManager handles MinIO credentials from Kubernetes secrets
type MinIOSecretManager struct {
    client.Client
}

type MinIOSecret struct {
    Endpoint        string
    AccessKeyID     string
    SecretAccessKey string
    UseSSL          bool
    BucketName      string
    Region          string
}

func (m *MinIOSecretManager) GetMinIOConfig(ctx context.Context, secretName, namespace string) (*MinIOSecret, error) {
    secret := &corev1.Secret{}
    err := m.Get(ctx, client.ObjectKey{Name: secretName, Namespace: namespace}, secret)
    if err != nil {
        return nil, fmt.Errorf("failed to get MinIO secret: %v", err)
    }

    return &MinIOSecret{
        Endpoint:        string(secret.Data["endpoint"]),
        AccessKeyID:     string(secret.Data["access-key-id"]),
        SecretAccessKey: string(secret.Data["secret-access-key"]),
        UseSSL:          string(secret.Data["use-ssl"]) == "true",
        BucketName:      string(secret.Data["bucket-name"]),
        Region:          string(secret.Data["region"]),
    }, nil
}
```

## Phase 2: Enhanced Backup Operations

### 2.1 Implement MinIO Backup Creation

**Updated CreateBackup** (`controllers/backup_restore.go`):
```go
func (b *BackupRestoreManager) CreateBackup(ctx context.Context, tenant *pishopv1alpha1.Tenant) error {
    log := ctrl.LoggerFrom(ctx)
    log.Info("Creating backup for tenant", "tenantId", tenant.Spec.TenantID)

    if tenant.Status.MongoDB == nil {
        return fmt.Errorf("MongoDB status not available")
    }

    backupName := fmt.Sprintf("tenant-%s-%s", tenant.Spec.TenantID, time.Now().Format("20060102-150405"))
    
    // Create backup job with MinIO upload
    backupJob := b.createMinIOBackupJob(tenant, backupName)
    if err := b.Create(ctx, backupJob); err != nil {
        return fmt.Errorf("failed to create backup job: %v", err)
    }

    log.Info("Backup job created successfully", "jobName", backupJob.Name, "backupName", backupName)
    return nil
}

func (b *BackupRestoreManager) createMinIOBackupJob(tenant *pishopv1alpha1.Tenant, backupName string) *batchv1.Job {
    namespace := b.getNamespaceName(tenant)
    
    return &batchv1.Job{
        ObjectMeta: metav1.ObjectMeta{
            Name:      fmt.Sprintf("backup-%s-%s", tenant.Spec.TenantID, backupName),
            Namespace: namespace,
            Labels: map[string]string{
                "app":        "mongodb-backup",
                "tenant-id":  tenant.Spec.TenantID,
                "backup-name": backupName,
                "component":  "backup",
            },
        },
        Spec: batchv1.JobSpec{
            TTLSecondsAfterFinished: int32Ptr(3600), // 1 hour
            Template: corev1.PodTemplateSpec{
                Spec: corev1.PodSpec{
                    RestartPolicy: corev1.RestartPolicyOnFailure,
                    Containers: []corev1.Container{
                        {
                            Name:  "mongodb-backup",
                            Image: "mongo:6.0",
                            Command: []string{"/bin/bash"},
                            Args:   []string{"-c", b.generateMinIOBackupScript(tenant, backupName)},
                            Env: []corev1.EnvVar{
                                {Name: "MONGO_URI", Value: b.MongoURI},
                                {Name: "MONGO_USERNAME", Value: b.MongoUsername},
                                {Name: "MONGO_PASSWORD", Value: b.MongoPassword},
                                {Name: "BACKUP_NAME", Value: backupName},
                                {Name: "TENANT_ID", Value: tenant.Spec.TenantID},
                                {Name: "MINIO_ENDPOINT", Value: b.MinIOConfig.Endpoint},
                                {Name: "MINIO_ACCESS_KEY", Value: b.MinIOConfig.AccessKeyID},
                                {Name: "MINIO_SECRET_KEY", Value: b.MinIOConfig.SecretAccessKey},
                                {Name: "MINIO_BUCKET", Value: b.MinIOConfig.BucketName},
                                {Name: "MINIO_USE_SSL", Value: fmt.Sprintf("%t", b.MinIOConfig.UseSSL)},
                            },
                            Resources: corev1.ResourceRequirements{
                                Requests: corev1.ResourceList{
                                    corev1.ResourceCPU:    resource.MustParse("100m"),
                                    corev1.ResourceMemory: resource.MustParse("256Mi"),
                                },
                                Limits: corev1.ResourceList{
                                    corev1.ResourceCPU:    resource.MustParse("500m"),
                                    corev1.ResourceMemory: resource.MustParse("1Gi"),
                                },
                            },
                        },
                    },
                },
            },
        },
    }
}
```

### 2.2 Generate MinIO Backup Script

**MinIO Backup Script** (`controllers/backup_restore.go`):
```go
func (b *BackupRestoreManager) generateMinIOBackupScript(tenant *pishopv1alpha1.Tenant, backupName string) string {
    script := `#!/bin/bash
set -e

echo "Starting MongoDB backup with MinIO upload for tenant ${TENANT_ID}"
echo "Backup name: ${BACKUP_NAME}"

# Install MinIO client
wget -q https://dl.min.io/client/mc/release/linux-amd64/mc
chmod +x mc
mv mc /usr/local/bin/

# Configure MinIO client
mc alias set minio ${MINIO_ENDPOINT} ${MINIO_ACCESS_KEY} ${MINIO_SECRET_KEY}

# Create backup directory
BACKUP_DIR="/tmp/${BACKUP_NAME}"
mkdir -p "${BACKUP_DIR}"

# Function to backup a single database
backup_database() {
    local db_name=$1
    echo "Backing up database: ${db_name}"

    mongodump \
        --uri="${MONGO_URI}" \
        --username="${MONGO_USERNAME}" \
        --password="${MONGO_PASSWORD}" \
        --db="${db_name}" \
        --out="${BACKUP_DIR}" \
        --gzip

    if [ $? -eq 0 ]; then
        echo "Successfully backed up database: ${db_name}"
    else
        echo "Failed to backup database: ${db_name}"
        exit 1
    fi
}

# Backup all databases
`

    // Add database backup commands
    for _, db := range tenant.Status.MongoDB.Databases {
        script += fmt.Sprintf("backup_database \"%s\"\n", db)
    }

    script += `
# Create backup metadata
cat > "${BACKUP_DIR}/metadata.json" << EOF
{
    "backup_name": "${BACKUP_NAME}",
    "tenant_id": "${TENANT_ID}",
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "databases": [
`

    // Add database list to metadata
    for i, db := range tenant.Status.MongoDB.Databases {
        if i > 0 {
            script += ","
        }
        script += fmt.Sprintf("        \"%s\"", db)
    }

    script += `
    ]
}
EOF

# Create backup archive
cd /tmp
tar -czf "${BACKUP_NAME}.tar.gz" "${BACKUP_NAME}"

# Encrypt backup (optional)
if [ -n "$BACKUP_ENCRYPTION_KEY" ]; then
    echo "Encrypting backup..."
    openssl enc -aes-256-cbc -salt -in "${BACKUP_NAME}.tar.gz" -out "${BACKUP_NAME}.tar.gz.enc" -k "${BACKUP_ENCRYPTION_KEY}"
    rm "${BACKUP_NAME}.tar.gz"
    mv "${BACKUP_NAME}.tar.gz.enc" "${BACKUP_NAME}.tar.gz"
fi

# Upload to MinIO
echo "Uploading backup to MinIO..."
mc cp "${BACKUP_NAME}.tar.gz" "minio/${MINIO_BUCKET}/backups/${TENANT_ID}/"

# Verify upload
if mc stat "minio/${MINIO_BUCKET}/backups/${TENANT_ID}/${BACKUP_NAME}.tar.gz" > /dev/null 2>&1; then
    echo "Backup uploaded successfully to MinIO"
    echo "Backup size: $(du -h ${BACKUP_NAME}.tar.gz | cut -f1)"
else
    echo "Failed to upload backup to MinIO"
    exit 1
fi

# Clean up local files
rm -rf "${BACKUP_DIR}" "${BACKUP_NAME}.tar.gz"

echo "Backup completed successfully: ${BACKUP_NAME}"
`

    return script
}
```

## Phase 3: Backup Management Operations

### 3.1 Implement Backup Listing

**Real Backup Listing** (`controllers/backup_restore.go`):
```go
func (b *BackupRestoreManager) ListBackups(ctx context.Context, tenantID string) ([]BackupInfo, error) {
    log := ctrl.LoggerFrom(ctx)
    log.Info("Listing backups for tenant", "tenantId", tenantID)

    // List objects in MinIO bucket
    objectCh := b.MinIOClient.ListObjects(ctx, b.MinIOConfig.BucketName, minio.ListObjectsOptions{
        Prefix:    fmt.Sprintf("backups/%s/", tenantID),
        Recursive: true,
    })

    var backups []BackupInfo
    for object := range objectCh {
        if object.Err != nil {
            return nil, fmt.Errorf("failed to list objects: %v", object.Err)
        }

        // Parse backup name from object key
        // Format: backups/{tenantID}/{backupName}.tar.gz
        parts := strings.Split(object.Key, "/")
        if len(parts) >= 3 {
            backupName := strings.TrimSuffix(parts[2], ".tar.gz")
            backups = append(backups, BackupInfo{
                Name:      backupName,
                Size:      object.Size,
                CreatedAt: object.LastModified,
                Key:       object.Key,
            })
        }
    }

    return backups, nil
}

type BackupInfo struct {
    Name      string
    Size      int64
    CreatedAt time.Time
    Key       string
}
```

### 3.2 Implement Backup Cleanup

**Backup Cleanup** (`controllers/backup_restore.go`):
```go
func (b *BackupRestoreManager) CleanupOldBackups(ctx context.Context, tenantID string, retentionDays int) error {
    log := ctrl.LoggerFrom(ctx)
    log.Info("Cleaning up old backups", "tenantId", tenantID, "retentionDays", retentionDays)

    backups, err := b.ListBackups(ctx, tenantID)
    if err != nil {
        return fmt.Errorf("failed to list backups: %v", err)
    }

    cutoffDate := time.Now().AddDate(0, 0, -retentionDays)
    deletedCount := 0

    for _, backup := range backups {
        if backup.CreatedAt.Before(cutoffDate) {
            err := b.MinIOClient.RemoveObject(ctx, b.MinIOConfig.BucketName, backup.Key, minio.RemoveObjectOptions{})
            if err != nil {
                log.Error(err, "Failed to delete backup", "backupName", backup.Name)
                continue
            }
            deletedCount++
            log.Info("Deleted old backup", "backupName", backup.Name, "createdAt", backup.CreatedAt)
        }
    }

    log.Info("Backup cleanup completed", "deletedCount", deletedCount)
    return nil
}
```

### 3.3 Implement Backup Restoration

**MinIO Backup Restoration** (`controllers/backup_restore.go`):
```go
func (b *BackupRestoreManager) RestoreBackup(ctx context.Context, tenant *pishopv1alpha1.Tenant, backupName string) error {
    log := ctrl.LoggerFrom(ctx)
    log.Info("Restoring backup for tenant", "tenantId", tenant.Spec.TenantID, "backupName", backupName)

    if tenant.Status.MongoDB == nil {
        return fmt.Errorf("MongoDB status not available")
    }

    // Create restore job with MinIO download
    restoreJob := b.createMinIORestoreJob(tenant, backupName)
    if err := b.Create(ctx, restoreJob); err != nil {
        return fmt.Errorf("failed to create restore job: %v", err)
    }

    log.Info("Restore job created successfully", "jobName", restoreJob.Name, "backupName", backupName)
    return nil
}

func (b *BackupRestoreManager) createMinIORestoreJob(tenant *pishopv1alpha1.Tenant, backupName string) *batchv1.Job {
    namespace := b.getNamespaceName(tenant)
    
    return &batchv1.Job{
        ObjectMeta: metav1.ObjectMeta{
            Name:      fmt.Sprintf("restore-%s-%s", tenant.Spec.TenantID, backupName),
            Namespace: namespace,
            Labels: map[string]string{
                "app":        "mongodb-restore",
                "tenant-id":  tenant.Spec.TenantID,
                "backup-name": backupName,
                "component":  "restore",
            },
        },
        Spec: batchv1.JobSpec{
            TTLSecondsAfterFinished: int32Ptr(3600), // 1 hour
            Template: corev1.PodTemplateSpec{
                Spec: corev1.PodSpec{
                    RestartPolicy: corev1.RestartPolicyOnFailure,
                    Containers: []corev1.Container{
                        {
                            Name:  "mongodb-restore",
                            Image: "mongo:6.0",
                            Command: []string{"/bin/bash"},
                            Args:   []string{"-c", b.generateMinIORestoreScript(tenant, backupName)},
                            Env: []corev1.EnvVar{
                                {Name: "MONGO_URI", Value: b.MongoURI},
                                {Name: "MONGO_USERNAME", Value: b.MongoUsername},
                                {Name: "MONGO_PASSWORD", Value: b.MongoPassword},
                                {Name: "BACKUP_NAME", Value: backupName},
                                {Name: "TENANT_ID", Value: tenant.Spec.TenantID},
                                {Name: "MINIO_ENDPOINT", Value: b.MinIOConfig.Endpoint},
                                {Name: "MINIO_ACCESS_KEY", Value: b.MinIOConfig.AccessKeyID},
                                {Name: "MINIO_SECRET_KEY", Value: b.MinIOConfig.SecretAccessKey},
                                {Name: "MINIO_BUCKET", Value: b.MinIOConfig.BucketName},
                                {Name: "MINIO_USE_SSL", Value: fmt.Sprintf("%t", b.MinIOConfig.UseSSL)},
                            },
                            Resources: corev1.ResourceRequirements{
                                Requests: corev1.ResourceList{
                                    corev1.ResourceCPU:    resource.MustParse("100m"),
                                    corev1.ResourceMemory: resource.MustParse("256Mi"),
                                },
                                Limits: corev1.ResourceList{
                                    corev1.ResourceCPU:    resource.MustParse("500m"),
                                    corev1.ResourceMemory: resource.MustParse("1Gi"),
                                },
                            },
                        },
                    },
                },
            },
        },
    }
}
```

## Phase 4: Operator Integration

### 4.1 Update Main Operator

**Updated Main** (`operator/main.go`):
```go
func main() {
    // ... existing code ...

    // Get MinIO configuration from secret
    minioSecretManager := &controllers.MinIOSecretManager{Client: mgr.GetClient()}
    minioConfig, err := minioSecretManager.GetMinIOConfig(ctrl.SetupSignalHandler(), "minio-credentials", "pishop-operator-system")
    if err != nil {
        setupLog.Error(err, "unable to get MinIO configuration")
        os.Exit(1)
    }

    // Initialize backup manager with MinIO
    backupManager, err := controllers.NewBackupRestoreManager(
        mgr.GetClient(),
        mongoURI,
        mongoUsername,
        mongoPassword,
        *minioConfig,
    )
    if err != nil {
        setupLog.Error(err, "unable to create backup manager")
        os.Exit(1)
    }

    // ... rest of the code ...
}
```

### 4.2 Add MinIO Secret Template

**MinIO Secret Template** (`config/manager/minio-credentials-secret-template.yaml`):
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: minio-credentials
  namespace: pishop-operator-system
type: Opaque
data:
  endpoint: ${MINIO_ENDPOINT}
  access-key-id: ${MINIO_ACCESS_KEY_ID}
  secret-access-key: ${MINIO_SECRET_ACCESS_KEY}
  use-ssl: ${MINIO_USE_SSL}
  bucket-name: ${MINIO_BUCKET_NAME}
  region: ${MINIO_REGION}
```

## Phase 5: Monitoring and Metrics

### 5.1 Add Backup Metrics

**Backup Metrics** (`controllers/backup_metrics.go`):
```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    backupTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "pishop_backup_total",
            Help: "Total number of backups created",
        },
        []string{"tenant_id", "status"},
    )

    backupSize = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "pishop_backup_size_bytes",
            Help: "Size of the last backup in bytes",
        },
        []string{"tenant_id"},
    )

    backupDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "pishop_backup_duration_seconds",
            Help: "Duration of backup operations in seconds",
        },
        []string{"tenant_id"},
    )
)
```

### 5.2 Update Backup Status

**Enhanced Backup Status** (`api/v1alpha1/tenant_types.go`):
```go
type BackupStatus struct {
    LastBackupTime     *metav1.Time     `json:"lastBackupTime,omitempty"`
    LastBackupName     string           `json:"lastBackupName,omitempty"`
    BackupCount        int              `json:"backupCount,omitempty"`
    LastBackupSize     string           `json:"lastBackupSize,omitempty"`
    LastBackupSizeBytes int64           `json:"lastBackupSizeBytes,omitempty"`
    BackupJobs         []BackupJobStatus `json:"backupJobs,omitempty"`
    MinIOStatus        MinIOStatus      `json:"minioStatus,omitempty"`
}

type MinIOStatus struct {
    BucketName string `json:"bucketName,omitempty"`
    Endpoint   string `json:"endpoint,omitempty"`
    Connected  bool   `json:"connected,omitempty"`
    LastCheck  *metav1.Time `json:"lastCheck,omitempty"`
}
```

## Phase 6: Testing Strategy

### 6.1 Unit Tests

- Test MinIO client initialization
- Test backup creation and upload
- Test backup listing and cleanup
- Test restore operations
- Test error handling

### 6.2 Integration Tests

- Test with real MinIO instance
- Test backup and restore workflows
- Test retention policies
- Test cross-cluster access

### 6.3 Performance Tests

- Test backup performance with large databases
- Test concurrent backup operations
- Test MinIO scalability
- Test network resilience

## Implementation Timeline

### Week 1-2: MinIO Integration
- [ ] Add MinIO client dependency
- [ ] Implement MinIO configuration management
- [ ] Create MinIO secret templates
- [ ] Write unit tests

### Week 3-4: Backup Operations
- [ ] Implement MinIO backup creation
- [ ] Implement backup listing
- [ ] Implement backup cleanup
- [ ] Add encryption support

### Week 5-6: Restore Operations
- [ ] Implement MinIO backup restoration
- [ ] Add restore validation
- [ ] Implement cross-cluster restore
- [ ] Write integration tests

### Week 7-8: Monitoring and Documentation
- [ ] Add backup metrics
- [ ] Update documentation
- [ ] Create migration guide
- [ ] Performance testing

## Benefits of MinIO Integration

1. **Persistent Storage**: Backups survive cluster restarts
2. **Scalability**: MinIO scales horizontally
3. **Cross-Cluster Access**: Backups accessible from any cluster
4. **Encryption**: Optional backup encryption
5. **Retention Policies**: Automated cleanup of old backups
6. **Monitoring**: Better backup status and metrics
7. **Disaster Recovery**: Reliable backup and restore capabilities

## Security Considerations

1. **Credential Management**: Store MinIO credentials in Kubernetes secrets
2. **Encryption**: Support for backup encryption
3. **Access Control**: Use MinIO IAM for fine-grained access control
4. **Network Security**: Use TLS for MinIO connections
5. **Audit Logging**: Log all backup operations

## Conclusion

This refactoring will transform the backup system from a basic PVC-based solution to a robust, scalable, and secure MinIO-based backup system, providing enterprise-grade backup capabilities for the PiShop Operator.