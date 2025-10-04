package controllers

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	pishopv1alpha1 "go.pilab.hu/shop/pishop-provisioner/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// createBackupPVC creates a PersistentVolumeClaim for backup storage
func (r *PRStackReconciler) createBackupPVC(ctx context.Context, prStack *pishopv1alpha1.PRStack, namespace string) error {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Creating backup PVC", "namespace", namespace, "prNumber", prStack.Spec.PRNumber)

	// Default backup configuration if not specified
	backupConfig := prStack.Spec.BackupConfig
	if backupConfig == nil {
		backupConfig = &pishopv1alpha1.BackupConfig{
			Enabled:      false,
			StorageClass: "standard",
			StorageSize:  "10Gi",
		}
	}

	// Skip if backup is not enabled
	if !backupConfig.Enabled {
		log.Info("Backup not enabled, skipping PVC creation")
		return nil
	}

	storageSize := backupConfig.StorageSize
	if storageSize == "" {
		storageSize = "10Gi"
	}

	storageClass := backupConfig.StorageClass
	if storageClass == "" {
		storageClass = "standard"
	}

	// Parse storage size
	quantity, err := resource.ParseQuantity(storageSize)
	if err != nil {
		return fmt.Errorf("invalid storage size %s: %v", storageSize, err)
	}

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mongodb-backup-pvc",
			Namespace: namespace,
			Labels: map[string]string{
				"app":       "mongodb-backup",
				"pr-number": prStack.Spec.PRNumber,
				"component": "backup-storage",
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: quantity,
				},
			},
			StorageClassName: &storageClass,
		},
	}

	// Check if PVC already exists
	existingPVC := &corev1.PersistentVolumeClaim{}
	err = r.Get(ctx, client.ObjectKey{Name: pvc.Name, Namespace: namespace}, existingPVC)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("failed to check existing PVC: %v", err)
		}
		// PVC doesn't exist, create it
		if err := r.Create(ctx, pvc); err != nil {
			return fmt.Errorf("failed to create backup PVC: %v", err)
		}
		log.Info("Successfully created backup PVC", "name", pvc.Name, "size", storageSize)
	} else {
		log.Info("Backup PVC already exists", "name", pvc.Name)
	}

	return nil
}

// cleanupBackupPVC removes the backup PVC when cleaning up the PR stack
func (r *PRStackReconciler) cleanupBackupPVC(ctx context.Context, prStack *pishopv1alpha1.PRStack, namespace string) error {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Cleaning up backup PVC", "namespace", namespace, "prNumber", prStack.Spec.PRNumber)

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mongodb-backup-pvc",
			Namespace: namespace,
		},
	}

	err := r.Delete(ctx, pvc)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("failed to delete backup PVC: %v", err)
		}
		log.Info("Backup PVC not found, already deleted")
	} else {
		log.Info("Successfully deleted backup PVC")
	}

	return nil
}

// ensureBackupStorage ensures backup storage is available for the PR stack
func (r *PRStackReconciler) ensureBackupStorage(ctx context.Context, prStack *pishopv1alpha1.PRStack, namespace string) error {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Ensuring backup storage", "namespace", namespace, "prNumber", prStack.Spec.PRNumber)

	// Create backup PVC if backup is enabled
	if err := r.createBackupPVC(ctx, prStack, namespace); err != nil {
		return fmt.Errorf("failed to create backup PVC: %v", err)
	}

	return nil
}
