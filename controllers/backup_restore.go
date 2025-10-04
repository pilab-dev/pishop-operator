package controllers

import (
	"context"
	"fmt"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	pishopv1alpha1 "go.pilab.hu/shop/pishop-provisioner/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BackupRestoreManager handles database backup and restore operations
type BackupRestoreManager struct {
	client.Client
	MongoURI      string
	MongoUsername string
	MongoPassword string
	BackupPath    string
}

// BackupSpec defines backup configuration
type BackupSpec struct {
	PRNumber    string
	Databases   []string
	BackupName  string
	Compression bool
}

// RestoreSpec defines restore configuration
type RestoreSpec struct {
	PRNumber   string
	BackupName string
	Databases  []string
}

// CreateBackup creates a backup of all databases for a PR stack
func (b *BackupRestoreManager) CreateBackup(ctx context.Context, prStack *pishopv1alpha1.PRStack) error {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Creating backup for PR stack", "prNumber", prStack.Spec.PRNumber)

	backupSpec := &BackupSpec{
		PRNumber:    prStack.Spec.PRNumber,
		Databases:   prStack.Status.MongoDB.Databases,
		BackupName:  fmt.Sprintf("pr-%s-%s", prStack.Spec.PRNumber, time.Now().Format("20060102-150405")),
		Compression: true,
	}

	// Create backup job
	backupJob := b.createBackupJob(prStack, backupSpec)
	if err := b.Create(ctx, backupJob); err != nil {
		return fmt.Errorf("failed to create backup job: %v", err)
	}

	log.Info("Backup job created successfully", "jobName", backupJob.Name, "backupName", backupSpec.BackupName)
	return nil
}

// RestoreBackup restores databases from a backup
func (b *BackupRestoreManager) RestoreBackup(ctx context.Context, prStack *pishopv1alpha1.PRStack, backupName string) error {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Restoring backup for PR stack", "prNumber", prStack.Spec.PRNumber, "backupName", backupName)

	restoreSpec := &RestoreSpec{
		PRNumber:   prStack.Spec.PRNumber,
		BackupName: backupName,
		Databases:  prStack.Status.MongoDB.Databases,
	}

	// Create restore job
	restoreJob := b.createRestoreJob(prStack, restoreSpec)
	if err := b.Create(ctx, restoreJob); err != nil {
		return fmt.Errorf("failed to create restore job: %v", err)
	}

	log.Info("Restore job created successfully", "jobName", restoreJob.Name, "backupName", backupName)
	return nil
}

// createBackupJob creates a Kubernetes Job for database backup
func (b *BackupRestoreManager) createBackupJob(prStack *pishopv1alpha1.PRStack, spec *BackupSpec) *batchv1.Job {
	namespace := fmt.Sprintf("pr-%s-shop-pilab-hu", prStack.Spec.PRNumber)
	jobName := fmt.Sprintf("backup-%s", spec.BackupName)

	// Create backup script
	backupScript := b.generateBackupScript(spec)

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: namespace,
			Labels: map[string]string{
				"app":       "mongodb-backup",
				"pr-number": prStack.Spec.PRNumber,
				"backup":    spec.BackupName,
			},
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: int32Ptr(3600), // Clean up after 1 hour
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":       "mongodb-backup",
						"pr-number": prStack.Spec.PRNumber,
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:    "mongodb-backup",
							Image:   "mongo:7.0",
							Command: []string{"/bin/bash", "-c"},
							Args:    []string{backupScript},
							Env: []corev1.EnvVar{
								{
									Name: "MONGO_URI",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{Name: "mongodb-secret"},
											Key:                  "uri",
										},
									},
								},
								{
									Name: "MONGO_USERNAME",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{Name: "mongodb-secret"},
											Key:                  "username",
										},
									},
								},
								{
									Name: "MONGO_PASSWORD",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{Name: "mongodb-secret"},
											Key:                  "password",
										},
									},
								},
								{
									Name:  "BACKUP_NAME",
									Value: spec.BackupName,
								},
								{
									Name:  "PR_NUMBER",
									Value: spec.PRNumber,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "backup-storage",
									MountPath: "/backup",
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("0m"),
									corev1.ResourceMemory: resource.MustParse("0Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("1Gi"),
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "backup-storage",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "mongodb-backup-pvc",
								},
							},
						},
					},
				},
			},
		},
	}

	return job
}

// createRestoreJob creates a Kubernetes Job for database restore
func (b *BackupRestoreManager) createRestoreJob(prStack *pishopv1alpha1.PRStack, spec *RestoreSpec) *batchv1.Job {
	namespace := fmt.Sprintf("pr-%s-shop-pilab-hu", prStack.Spec.PRNumber)
	jobName := fmt.Sprintf("restore-%s", spec.BackupName)

	// Create restore script
	restoreScript := b.generateRestoreScript(spec)

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: namespace,
			Labels: map[string]string{
				"app":       "mongodb-restore",
				"pr-number": prStack.Spec.PRNumber,
				"backup":    spec.BackupName,
			},
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: int32Ptr(3600), // Clean up after 1 hour
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":       "mongodb-restore",
						"pr-number": prStack.Spec.PRNumber,
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:    "mongodb-restore",
							Image:   "mongo:7.0",
							Command: []string{"/bin/bash", "-c"},
							Args:    []string{restoreScript},
							Env: []corev1.EnvVar{
								{
									Name: "MONGO_URI",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{Name: "mongodb-secret"},
											Key:                  "uri",
										},
									},
								},
								{
									Name: "MONGO_USERNAME",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{Name: "mongodb-secret"},
											Key:                  "username",
										},
									},
								},
								{
									Name: "MONGO_PASSWORD",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{Name: "mongodb-secret"},
											Key:                  "password",
										},
									},
								},
								{
									Name:  "BACKUP_NAME",
									Value: spec.BackupName,
								},
								{
									Name:  "PR_NUMBER",
									Value: spec.PRNumber,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "backup-storage",
									MountPath: "/backup",
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("0m"),
									corev1.ResourceMemory: resource.MustParse("0Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("1Gi"),
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "backup-storage",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "mongodb-backup-pvc",
								},
							},
						},
					},
				},
			},
		},
	}

	return job
}

// generateBackupScript creates the backup script for mongodump
func (b *BackupRestoreManager) generateBackupScript(spec *BackupSpec) string {
	script := `#!/bin/bash
set -e

echo "Starting MongoDB backup for PR ${PR_NUMBER}"
echo "Backup name: ${BACKUP_NAME}"

# Create backup directory
BACKUP_DIR="/backup/${BACKUP_NAME}"
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
	for _, db := range spec.Databases {
		script += fmt.Sprintf("backup_database \"%s\"\n", db)
	}

	script += `
# Create backup metadata
cat > "${BACKUP_DIR}/metadata.json" << EOF
{
    "backup_name": "${BACKUP_NAME}",
    "pr_number": "${PR_NUMBER}",
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "databases": [
`

	// Add database list to metadata
	for i, db := range spec.Databases {
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
cd /backup
tar -czf "${BACKUP_NAME}.tar.gz" "${BACKUP_NAME}"
rm -rf "${BACKUP_NAME}"

echo "Backup completed successfully: ${BACKUP_NAME}.tar.gz"
echo "Backup size: $(du -h ${BACKUP_NAME}.tar.gz | cut -f1)"
`

	return script
}

// generateRestoreScript creates the restore script for mongorestore
func (b *BackupRestoreManager) generateRestoreScript(spec *RestoreSpec) string {
	script := `#!/bin/bash
set -e

echo "Starting MongoDB restore for PR ${PR_NUMBER}"
echo "Backup name: ${BACKUP_NAME}"

# Extract backup archive
BACKUP_DIR="/backup/${BACKUP_NAME}"
cd /backup

if [ ! -f "${BACKUP_NAME}.tar.gz" ]; then
    echo "Backup file not found: ${BACKUP_NAME}.tar.gz"
    exit 1
fi

echo "Extracting backup archive..."
tar -xzf "${BACKUP_NAME}.tar.gz"

if [ ! -d "${BACKUP_DIR}" ]; then
    echo "Backup directory not found after extraction: ${BACKUP_DIR}"
    exit 1
fi

# Function to restore a single database
restore_database() {
    local db_name=$1
    echo "Restoring database: ${db_name}"

    # Drop existing database first
    echo "Dropping existing database: ${db_name}"
    mongo "${MONGO_URI}" --username="${MONGO_USERNAME}" --password="${MONGO_PASSWORD}" --eval "db.getSiblingDB('${db_name}').dropDatabase()"

    # Restore database
    mongorestore \
        --uri="${MONGO_URI}" \
        --username="${MONGO_USERNAME}" \
        --password="${MONGO_PASSWORD}" \
        --db="${db_name}" \
        --gzip \
        "${BACKUP_DIR}/${db_name}"

    if [ $? -eq 0 ]; then
        echo "Successfully restored database: ${db_name}"
    else
        echo "Failed to restore database: ${db_name}"
        exit 1
    fi
}

# Restore all databases
`

	// Add database restore commands
	for _, db := range spec.Databases {
		script += fmt.Sprintf("restore_database \"%s\"\n", db)
	}

	script += `
# Clean up extracted files
rm -rf "${BACKUP_DIR}"

echo "Restore completed successfully"
`

	return script
}

// ListBackups lists available backups for a PR
func (b *BackupRestoreManager) ListBackups(ctx context.Context, prNumber string) ([]string, error) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Listing backups for PR", "prNumber", prNumber)

	// This would typically query a storage backend or PVC
	// For now, we'll return a placeholder implementation
	backups := []string{
		fmt.Sprintf("pr-%s-20240101-120000", prNumber),
		fmt.Sprintf("pr-%s-20240102-120000", prNumber),
	}

	return backups, nil
}

// CleanupOldBackups removes old backups based on retention policy
func (b *BackupRestoreManager) CleanupOldBackups(ctx context.Context, prNumber string, retentionDays int) error {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Cleaning up old backups", "prNumber", prNumber, "retentionDays", retentionDays)

	// Implementation would clean up backups older than retention period
	// This is a placeholder for the actual cleanup logic

	return nil
}

// Helper functions
func mustParseQuantity(s string) corev1.ResourceList {
	// This is a simplified version - in production, use resource.MustParse
	return corev1.ResourceList{}
}
