package controllers

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	pishopv1alpha1 "go.pilab.hu/shop/pishop-provisioner/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	// Finalizer name for PRStack resources
	FinalizerName = "shop.pilab.hu/finalizer"

	// Namespace name pattern
	NamespacePattern = "pr-%s-shop-pilab-hu"

	// Stack expiration time
	StackExpirationTime = time.Hour

	// Requeue intervals
	RequeueIntervalShort  = time.Second * 5
	RequeueIntervalMedium = time.Second * 30
	RequeueIntervalLong   = time.Minute * 5

	// Phase constants
	PhaseInitialization = ""
	PhaseProvisioning   = "Provisioning"
	PhaseDeploying      = "Deploying"
	PhaseRunning        = "Running"
	PhaseInactive       = "Inactive"
	PhaseCleaning       = "Cleaning"
	PhaseCleaned        = "Cleaned"

	// Event types
	EventTypeInitializing         = "Initializing"
	EventTypeProvisioning         = "Provisioning"
	EventTypeDeploying            = "Deploying"
	EventTypeRunning              = "Running"
	EventTypeCleaning             = "Cleaning"
	EventTypeActivating           = "Activating"
	EventTypeScalingDown          = "ScalingDown"
	EventTypeInactive             = "Inactive"
	EventTypeRolloutTriggered     = "RolloutTriggered"
	EventTypeDeployed             = "Deployed"
	EventTypeCleanupComplete      = "CleanupComplete"
	EventTypeProvisioningComplete = "ProvisioningComplete"
	EventTypeProvisioningFailed   = "ProvisioningFailed"
	EventTypeRolloutFailed        = "RolloutFailed"
	EventTypeScaleDownFailed      = "ScaleDownFailed"
	EventTypeStackExpired         = "StackExpired"

	// Default services - moved to constants.go

	// MongoDB secret name
	MongoDBSecretName = "mongodb-secret"

	// Restart annotation
	RestartAnnotation = "kubectl.kubernetes.io/restartedAt"
)

// PRStackReconciler reconciles a PRStack object
type PRStackReconciler struct {
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

//+kubebuilder:rbac:groups=shop.pilab.hu,resources=prstacks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=shop.pilab.hu,resources=prstacks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=shop.pilab.hu,resources=prstacks/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;watch;create;update;patch;delete

func (r *PRStackReconciler) getNamespaceName(prNumber string) string {
	return fmt.Sprintf(NamespacePattern, prNumber)
}

// getDomain returns the domain to use for ingress based on CustomDomain or default pattern
func (r *PRStackReconciler) getDomain(prStack *pishopv1alpha1.PRStack) string {
	if prStack.Spec.CustomDomain != "" {
		return prStack.Spec.CustomDomain
	}
	return fmt.Sprintf("pr-%s.%s", prStack.Spec.PRNumber, r.BaseDomain)
}

// Reconcile is part of the main kubernetes reconciliation loop
func (r *PRStackReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	// Fetch the PRStack instance
	var prStack pishopv1alpha1.PRStack
	if err := r.Get(ctx, req.NamespacedName, &prStack); err != nil {
		if errors.IsNotFound(err) {
			log.Info("PRStack resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get PRStack")

		return ctrl.Result{}, err
	}

	// Handle deletion
	if !prStack.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, &prStack)
	}

	// Add finalizer if not present
	if !containsString(prStack.Finalizers, FinalizerName) {
		prStack.Finalizers = append(prStack.Finalizers, FinalizerName)
		if err := r.Update(ctx, &prStack); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Set creation time if not set
	now := metav1.Now()
	if prStack.Status.CreatedAt == nil {
		prStack.Status.CreatedAt = &now
		prStack.Status.LastActiveAt = &now

		if err := r.Status().Update(ctx, &prStack); err != nil {
			return ctrl.Result{}, err
		}
	}

	// If stack is becoming active (transitioning from inactive/expired), update LastActiveAt
	// This prevents immediate re-expiration
	wasReactivated := false
	isExpired := r.isStackExpired(&prStack)
	log.Info("Checking reactivation conditions",
		"prNumber", prStack.Spec.PRNumber,
		"active", prStack.Spec.Active,
		"phase", prStack.Status.Phase,
		"isExpired", isExpired)

	if prStack.Spec.Active && (prStack.Status.Phase == "Inactive" || isExpired) {
		log.Info("Stack being reactivated, updating LastActiveAt", "prNumber", prStack.Spec.PRNumber)
		prStack.Status.LastActiveAt = &now
		if err := r.Status().Update(ctx, &prStack); err != nil {
			return ctrl.Result{}, err
		}
		wasReactivated = true
	}

	// Check if stack is expired and set Active to false
	// Skip if we just reactivated the stack
	if !wasReactivated && r.isStackExpired(&prStack) && prStack.Spec.Active {
		return r.handleStackExpiration(ctx, &prStack)
	}

	// Handle active/inactive state
	if !prStack.Spec.Active {
		return r.handleInactiveStack(ctx, &prStack)
	}

	// Update last active time when active
	prStack.Status.LastActiveAt = &now

	// Check if a deployment rollout is requested
	if r.shouldRolloutDeployments(&prStack) {
		log.Info("Deployment rollout requested", "prNumber", prStack.Spec.PRNumber, "deployedAt", prStack.Spec.DeployedAt)
		if err := r.rolloutDeployments(ctx, &prStack); err != nil {
			log.Error(err, "Failed to rollout deployments")
			r.Recorder.Event(&prStack, corev1.EventTypeWarning, EventTypeRolloutFailed, fmt.Sprintf("Failed to rollout deployments: %v", err))
			prStack.Status.Message = fmt.Sprintf("Rollout failed: %v", err)
			r.Status().Update(ctx, &prStack)
			return ctrl.Result{RequeueAfter: RequeueIntervalMedium}, err
		}
		r.Recorder.Event(&prStack, corev1.EventTypeNormal, EventTypeRolloutTriggered, fmt.Sprintf("PR #%s deployments rolled out successfully", prStack.Spec.PRNumber))
		prStack.Status.LastDeployedAt = prStack.Spec.DeployedAt
		if err := r.Status().Update(ctx, &prStack); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Reconcile based on current phase
	return r.reconcileByPhase(ctx, &prStack)
}

func (r *PRStackReconciler) handleInitialization(ctx context.Context, prStack *pishopv1alpha1.PRStack) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Initializing PR stack", "prNumber", prStack.Spec.PRNumber)

	r.Recorder.Event(prStack, corev1.EventTypeNormal, EventTypeInitializing, fmt.Sprintf("Starting initialization for PR #%s", prStack.Spec.PRNumber))

	// Update status to Provisioning
	prStack.Status.Phase = PhaseProvisioning
	prStack.Status.Message = "Starting PR stack provisioning"
	if err := r.Status().Update(ctx, prStack); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: RequeueIntervalShort}, nil
}

func (r *PRStackReconciler) handleProvisioning(ctx context.Context, prStack *pishopv1alpha1.PRStack) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Provisioning PR stack resources", "prNumber", prStack.Spec.PRNumber)

	r.Recorder.Event(prStack, corev1.EventTypeNormal, EventTypeProvisioning, fmt.Sprintf("Provisioning infrastructure for PR #%s", prStack.Spec.PRNumber))

	// Create namespace first
	namespaceName := r.getNamespaceName(prStack.Spec.PRNumber)
	if err := r.createNamespace(ctx, namespaceName); err != nil {
		return r.recordProvisioningError(ctx, prStack, "Namespace", err)
	}

	// Provision MongoDB databases and users
	if err := r.provisionMongoDB(ctx, prStack); err != nil {
		return r.recordProvisioningError(ctx, prStack, "MongoDB", err)
	}

	// Create secret for MongoDB credentials
	if err := r.createMongoDBSecret(ctx, prStack); err != nil {
		return r.recordProvisioningError(ctx, prStack, "MongoDB secret", err)
	}

	// Provision NATS server and subjects
	if err := r.provisionNATS(ctx, prStack); err != nil {
		return r.recordProvisioningError(ctx, prStack, "NATS", err)
	}

	// Provision Redis keyspaces
	if err := r.provisionRedis(ctx, prStack); err != nil {
		return r.recordProvisioningError(ctx, prStack, "Redis", err)
	}

	r.Recorder.Event(prStack, corev1.EventTypeNormal, EventTypeProvisioningComplete, "Infrastructure provisioning completed successfully")

	// Move to deployment phase
	prStack.Status.Phase = PhaseDeploying
	prStack.Status.Message = "Resources provisioned, starting service deployment"
	if err := r.Status().Update(ctx, prStack); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: RequeueIntervalShort}, nil
}

func (r *PRStackReconciler) createMongoDBSecret(ctx context.Context, prStack *pishopv1alpha1.PRStack) error {
	namespaceName := r.getNamespaceName(prStack.Spec.PRNumber)
	
	// Check if MongoDB credentials are available
	if prStack.Status.MongoDB == nil {
		return fmt.Errorf("MongoDB credentials not available in status")
	}
	
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      MongoDBSecretName,
			Namespace: namespaceName,
		},
		StringData: map[string]string{
			"username":         prStack.Status.MongoDB.User,
			"password":         prStack.Status.MongoDB.Password,
			"connectionString": prStack.Status.MongoDB.ConnectionString,
			"databases":        strings.Join(prStack.Status.MongoDB.Databases, ","),
		},
	}

	if err := r.Create(ctx, secret); err != nil {
		return fmt.Errorf("failed to create MongoDB secret: %w", err)
	}

	return nil
}

func (r *PRStackReconciler) handleDeployment(ctx context.Context, prStack *pishopv1alpha1.PRStack) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Deploying PR stack services", "prNumber", prStack.Spec.PRNumber)

	r.Recorder.Event(prStack, corev1.EventTypeNormal, EventTypeDeploying, fmt.Sprintf("Deploying services for PR #%s", prStack.Spec.PRNumber))

	// Namespace should already exist from provisioning phase
	namespaceName := r.getNamespaceName(prStack.Spec.PRNumber)

	// Create registry secret for image pulls
	if err := r.createRegistrySecret(ctx, namespaceName); err != nil {
		r.updateStatusWithError(ctx, prStack, "Failed to create registry secret", err)
		return ctrl.Result{RequeueAfter: RequeueIntervalMedium}, err
	}

	// Create MongoDB, NATS, and Redis resources
	if err := r.createMongoDBResources(ctx, prStack, namespaceName); err != nil {
		r.updateStatusWithError(ctx, prStack, "Failed to create MongoDB resources", err)
		return ctrl.Result{RequeueAfter: RequeueIntervalMedium}, err
	}

	// Create NATS resources
	if err := r.createNATSResources(ctx, prStack, namespaceName); err != nil {
		r.updateStatusWithError(ctx, prStack, "Failed to create NATS resources", err)
		return ctrl.Result{RequeueAfter: RequeueIntervalMedium}, err
	}

	// Create Redis resources
	if err := r.createRedisResources(ctx, prStack, namespaceName); err != nil {
		r.updateStatusWithError(ctx, prStack, "Failed to create Redis resources", err)
		return ctrl.Result{RequeueAfter: RequeueIntervalMedium}, err
	}

	// Ensure backup storage is available
	if err := r.ensureBackupStorage(ctx, prStack, namespaceName); err != nil {
		r.updateStatusWithError(ctx, prStack, "Failed to create backup storage", err)
		return ctrl.Result{RequeueAfter: RequeueIntervalMedium}, err
	}

	// Deploy services based on the services list
	services := prStack.Spec.Services
	if len(services) == 0 {
		// Default services if none specified
		services = strings.Split(DefaultServicesString, ",")
	}

	// Deploy each service
	var serviceStatuses []pishopv1alpha1.ServiceStatus
	for _, serviceName := range services {
		if err := r.createServiceDeployment(ctx, prStack, namespaceName, serviceName); err != nil {
			log.Error(err, "Failed to deploy service", "service", serviceName)
			serviceStatuses = append(serviceStatuses, pishopv1alpha1.ServiceStatus{
				Name:    serviceName,
				Status:  "Failed",
				Message: err.Error(),
			})
		} else {
			serviceStatuses = append(serviceStatuses, pishopv1alpha1.ServiceStatus{
				Name:    serviceName,
				Status:  "Running",
				Message: "Service deployed successfully",
			})
		}
	}

	// Update status with service information
	prStack.Status.Services = serviceStatuses

	// Move to running phase FIRST to avoid reconciliation loop
	prStack.Status.Phase = PhaseRunning
	prStack.Status.Message = "PR stack is running"
	if err := r.Status().Update(ctx, prStack); err != nil {
		return ctrl.Result{}, err
	}

	r.Recorder.Event(prStack, corev1.EventTypeNormal, EventTypeDeployed, fmt.Sprintf("PR #%s stack is now running with %d services", prStack.Spec.PRNumber, len(serviceStatuses)))

	return ctrl.Result{RequeueAfter: RequeueIntervalLong}, nil
}

func (r *PRStackReconciler) handleRunning(ctx context.Context, prStack *pishopv1alpha1.PRStack) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("PR stack is running", "prNumber", prStack.Spec.PRNumber)

	// Check if we need to scale based on Active flag
	namespaceName := r.getNamespaceName(prStack.Spec.PRNumber)

	// Get current deployment state
	var deployments appsv1.DeploymentList
	if err := r.List(ctx, &deployments, client.InNamespace(namespaceName)); err == nil && len(deployments.Items) > 0 {
		// Check if replicas match desired state (max 1 replica)
		desiredReplicas := int32(1)
		if !prStack.Spec.Active {
			desiredReplicas = 0
		}

		currentReplicas := int32(-1)
		if len(deployments.Items) > 0 && deployments.Items[0].Spec.Replicas != nil {
			currentReplicas = *deployments.Items[0].Spec.Replicas
		}

		// Only scale if there's a mismatch
		if currentReplicas != desiredReplicas {
			log.Info("Scaling deployments to match active state", "current", currentReplicas, "desired", desiredReplicas)
			if err := r.scaleDeployments(ctx, namespaceName, desiredReplicas); err != nil {
				log.Error(err, "Failed to scale deployments")
			}
		}
	}

	// Check service health
	allHealthy := true
	for _, service := range prStack.Status.Services {
		if service.Status != "Running" {
			allHealthy = false
			break
		}
	}

	if !allHealthy {
		prStack.Status.Message = "Some services are not healthy"
		r.Status().Update(ctx, prStack)
	}

	return ctrl.Result{RequeueAfter: RequeueIntervalLong}, nil
}

func (r *PRStackReconciler) handleCleaning(ctx context.Context, prStack *pishopv1alpha1.PRStack) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Cleaning up PR stack", "prNumber", prStack.Spec.PRNumber)

	r.Recorder.Event(prStack, corev1.EventTypeNormal, EventTypeCleaning, fmt.Sprintf("Cleaning up all resources for PR #%s", prStack.Spec.PRNumber))

	// Create final backup before cleanup if enabled
	if prStack.Spec.BackupConfig != nil && prStack.Spec.BackupConfig.Enabled && prStack.Status.MongoDB != nil {
		log.Info("Creating final backup before cleanup", "prNumber", prStack.Spec.PRNumber)
		if err := r.BackupManager.CreateBackup(ctx, prStack); err != nil {
			log.Error(err, "Failed to create final backup")
			// Continue with cleanup even if backup fails
		}
	}

	// FIRST: Clean up Kubernetes resources (deployments, services, etc.)
	// This ensures deployments are removed before database cleanup
	if err := r.cleanupAllResources(ctx, prStack); err != nil {
		r.updateStatusWithError(ctx, prStack, "Kubernetes cleanup failed", err)
		return ctrl.Result{RequeueAfter: RequeueIntervalMedium}, err
	}

	// SECOND: Clean up MongoDB databases and users
	if err := r.cleanupMongoDB(ctx, prStack); err != nil {
		r.updateStatusWithError(ctx, prStack, "MongoDB cleanup failed", err)
		return ctrl.Result{RequeueAfter: RequeueIntervalMedium}, err
	}

	// THIRD: Clean up NATS subjects
	if err := r.cleanupNATS(ctx, prStack); err != nil {
		r.updateStatusWithError(ctx, prStack, "NATS cleanup failed", err)
		return ctrl.Result{RequeueAfter: RequeueIntervalMedium}, err
	}

	// FOURTH: Clean up Redis keyspaces
	if err := r.cleanupRedis(ctx, prStack); err != nil {
		r.updateStatusWithError(ctx, prStack, "Redis cleanup failed", err)
		return ctrl.Result{RequeueAfter: RequeueIntervalMedium}, err
	}

	// Update status to indicate cleanup is complete BEFORE removing finalizer
	prStack.Status.Phase = PhaseCleaned
	prStack.Status.Message = "All resources have been cleaned up"
	if err := r.Status().Update(ctx, prStack); err != nil {
		return ctrl.Result{}, err
	}

	r.Recorder.Event(prStack, corev1.EventTypeNormal, EventTypeCleanupComplete, fmt.Sprintf("All resources for PR #%s cleaned up successfully", prStack.Spec.PRNumber))

	// Remove finalizer to allow deletion (this triggers immediate deletion by K8s)
	prStack.Finalizers = removeString(prStack.Finalizers, FinalizerName)
	if err := r.Update(ctx, prStack); err != nil {
		return ctrl.Result{}, err
	}

	log.Info("Successfully completed cleanup and removed finalizer", "prNumber", prStack.Spec.PRNumber)
	return ctrl.Result{}, nil
}

func (r *PRStackReconciler) handleDeletion(ctx context.Context, prStack *pishopv1alpha1.PRStack) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Handling PR stack deletion", "prNumber", prStack.Spec.PRNumber, "phase", prStack.Status.Phase)

	// If already in cleaning phase, proceed with cleanup
	if prStack.Status.Phase == PhaseCleaning {
		log.Info("PRStack is already in cleaning phase, proceeding with cleanup", "prNumber", prStack.Spec.PRNumber)
		return r.handleCleaning(ctx, prStack)
	}

	// Set phase to cleaning
	log.Info("Setting PRStack phase to Cleaning", "prNumber", prStack.Spec.PRNumber)
	prStack.Status.Phase = PhaseCleaning
	prStack.Status.Message = "Cleaning up PR stack resources"
	if err := r.Status().Update(ctx, prStack); err != nil {
		log.Error(err, "Failed to update PRStack status to Cleaning")
		return ctrl.Result{}, err
	}

	log.Info("Successfully updated PRStack status to Cleaning", "prNumber", prStack.Spec.PRNumber)
	return ctrl.Result{RequeueAfter: RequeueIntervalShort}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PRStackReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&pishopv1alpha1.PRStack{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}

// New handler functions for enhanced lifecycle management

func (r *PRStackReconciler) isStackExpired(prStack *pishopv1alpha1.PRStack) bool {
	if prStack.Status.LastActiveAt == nil {
		// Fallback to CreatedAt if LastActiveAt is not set
		if prStack.Status.CreatedAt == nil {
			return false
		}
		return time.Since(prStack.Status.CreatedAt.Time) > StackExpirationTime
	}
	return time.Since(prStack.Status.LastActiveAt.Time) > StackExpirationTime
}

func (r *PRStackReconciler) handleInactiveStack(ctx context.Context, prStack *pishopv1alpha1.PRStack) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	// Check if this is due to expiration
	isExpired := r.isStackExpired(prStack)
	reason := "marked inactive"
	if isExpired {
		reason = "expired (older than 1 hour)"
	}

	log.Info("Handling inactive stack - scaling down deployments", "prNumber", prStack.Spec.PRNumber, "reason", reason)

	r.Recorder.Event(prStack, corev1.EventTypeNormal, EventTypeScalingDown, fmt.Sprintf("PR #%s %s, scaling down all deployments to 0", prStack.Spec.PRNumber, reason))

	namespaceName := r.getNamespaceName(prStack.Spec.PRNumber)

	// Scale down all deployments to 0 replicas
	if err := r.scaleDeployments(ctx, namespaceName, 0); err != nil {
		log.Error(err, "Failed to scale down deployments")
		r.Recorder.Event(prStack, corev1.EventTypeWarning, EventTypeScaleDownFailed, fmt.Sprintf("Failed to scale down: %v", err))
		r.updateStatusWithError(ctx, prStack, "Failed to scale down", err)
		return ctrl.Result{RequeueAfter: RequeueIntervalMedium}, err
	}

	prStack.Status.Phase = PhaseInactive
	if isExpired {
		prStack.Status.Message = fmt.Sprintf("Stack expired (age: %v) - all deployments scaled to 0", time.Since(prStack.Status.CreatedAt.Time).Round(time.Minute))
	} else {
		prStack.Status.Message = "Stack is inactive - all deployments scaled to 0"
	}

	if err := r.Status().Update(ctx, prStack); err != nil {
		return ctrl.Result{}, err
	}

	r.Recorder.Event(prStack, corev1.EventTypeNormal, EventTypeInactive, fmt.Sprintf("PR #%s stack scaled down to 0 replicas", prStack.Spec.PRNumber))

	return ctrl.Result{RequeueAfter: RequeueIntervalLong}, nil
}

func (r *PRStackReconciler) scaleDeployments(ctx context.Context, namespace string, replicas int32) error {
	log := ctrl.LoggerFrom(ctx)

	// List all deployments in the namespace
	var deployments appsv1.DeploymentList
	if err := r.List(ctx, &deployments, client.InNamespace(namespace)); err != nil {
		return fmt.Errorf("failed to list deployments: %v", err)
	}

	// Scale each deployment
	scaledCount := 0
	for _, deployment := range deployments.Items {
		if deployment.Spec.Replicas != nil && *deployment.Spec.Replicas == replicas {
			continue // Already at desired replica count
		}

		log.Info("Scaling deployment", "name", deployment.Name, "namespace", namespace, "replicas", replicas)

		// Update the deployment
		deployment.Spec.Replicas = &replicas
		if err := r.Update(ctx, &deployment); err != nil {
			log.Error(err, "Failed to scale deployment", "name", deployment.Name)
			return fmt.Errorf("failed to scale deployment %s: %v", deployment.Name, err)
		}
		scaledCount++
	}

	if scaledCount > 0 {
		log.Info("Scaled deployments", "count", scaledCount, "namespace", namespace, "replicas", replicas)
	}

	return nil
}

// func (r *PRStackReconciler) forceRestartDeployments(ctx context.Context, namespace string) error {
// 	log := ctrl.LoggerFrom(ctx)

// 	// List all deployments in the namespace
// 	var deployments appsv1.DeploymentList
// 	if err := r.List(ctx, &deployments, client.InNamespace(namespace)); err != nil {
// 		return fmt.Errorf("failed to list deployments: %v", err)
// 	}

// 	// Force restart each deployment by updating the pod template
// 	for _, deployment := range deployments.Items {
// 		log.Info("Force restarting deployment", "name", deployment.Name, "namespace", namespace)

// 		// Add or update an annotation to force pod restart
// 		if deployment.Spec.Template.Annotations == nil {
// 			deployment.Spec.Template.Annotations = make(map[string]string)
// 		}
// 		deployment.Spec.Template.Annotations[RestartAnnotation] = time.Now().Format(time.RFC3339)

// 		// Set image pull policy to Always to force image re-pull
// 		for i := range deployment.Spec.Template.Spec.Containers {
// 			deployment.Spec.Template.Spec.Containers[i].ImagePullPolicy = corev1.PullAlways
// 		}

// 		if err := r.Update(ctx, &deployment); err != nil {
// 			log.Error(err, "Failed to restart deployment", "name", deployment.Name)
// 			return fmt.Errorf("failed to restart deployment %s: %v", deployment.Name, err)
// 		}
// 	}

// 	return nil
// }

// shouldRolloutDeployments checks if deployments should be rolled out based on deployedAt timestamp
func (r *PRStackReconciler) shouldRolloutDeployments(prStack *pishopv1alpha1.PRStack) bool {
	// If deployedAt is not set in spec, no rollout needed
	if prStack.Spec.DeployedAt == nil {
		return false
	}

	// If lastDeployedAt is not set in status, rollout is needed
	if prStack.Status.LastDeployedAt == nil {
		return true
	}

	// Rollout if deployedAt is different from lastDeployedAt
	return !prStack.Spec.DeployedAt.Equal(prStack.Status.LastDeployedAt)
}

// rolloutDeployments triggers a rollout of all deployments in the namespace
func (r *PRStackReconciler) rolloutDeployments(ctx context.Context, prStack *pishopv1alpha1.PRStack) error {
	log := ctrl.LoggerFrom(ctx)
	namespaceName := r.getNamespaceName(prStack.Spec.PRNumber)

	log.Info("Rolling out all deployments", "namespace", namespaceName)

	// List all deployments in the namespace
	deployments := &appsv1.DeploymentList{}
	if err := r.List(ctx, deployments, client.InNamespace(namespaceName)); err != nil {
		return fmt.Errorf("failed to list deployments: %w", err)
	}

	if len(deployments.Items) == 0 {
		log.Info("No deployments found to rollout", "namespace", namespaceName)
		return nil
	}

	// Trigger rollout for each deployment by updating the restartedAt annotation
	for i := range deployments.Items {
		deployment := &deployments.Items[i]
		log.Info("Rolling out deployment", "name", deployment.Name, "namespace", deployment.Namespace)

		// Add/update the restartedAt annotation to trigger a rollout
		if deployment.Spec.Template.Annotations == nil {
			deployment.Spec.Template.Annotations = make(map[string]string)
		}
		deployment.Spec.Template.Annotations[RestartAnnotation] = prStack.Spec.DeployedAt.Format(time.RFC3339)

		if err := r.Update(ctx, deployment); err != nil {
			log.Error(err, "Failed to rollout deployment", "name", deployment.Name)
			return fmt.Errorf("failed to rollout deployment %s: %w", deployment.Name, err)
		}

		log.Info("Successfully triggered rollout", "deployment", deployment.Name)
	}

	log.Info("All deployments rolled out successfully", "count", len(deployments.Items))
	return nil
}

// Helper functions
func containsString(slice []string, s string) bool {
	return slices.Contains(slice, s)
}

func removeString(slice []string, s string) []string {
	index := slices.Index(slice, s)
	if index == -1 {
		return slice
	}
	// Create a new slice without the element at index
	result := make([]string, 0, len(slice)-1)
	result = append(result, slice[:index]...)
	result = append(result, slice[index+1:]...)
	return result
}

// handleStackExpiration handles stack expiration logic
func (r *PRStackReconciler) handleStackExpiration(ctx context.Context, prStack *pishopv1alpha1.PRStack) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	age := time.Since(prStack.Status.LastActiveAt.Time)

	log.Info("Stack expired, setting Active to false", "prNumber", prStack.Spec.PRNumber, "age", age)
	r.Recorder.Event(prStack, corev1.EventTypeWarning, EventTypeStackExpired,
		fmt.Sprintf("PR #%s stack expired (age: %v), deactivating", prStack.Spec.PRNumber, age.Round(time.Minute)))

	prStack.Spec.Active = false
	if err := r.Update(ctx, prStack); err != nil {
		return ctrl.Result{}, err
	}

	// Requeue immediately to handle the inactive state
	return ctrl.Result{Requeue: true}, nil
}

// reconcileByPhase handles reconciliation based on current phase
func (r *PRStackReconciler) reconcileByPhase(ctx context.Context, prStack *pishopv1alpha1.PRStack) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	switch prStack.Status.Phase {
	case PhaseInitialization:
		return r.handleInitialization(ctx, prStack)
	case PhaseInactive:
		// Transition from Inactive to Initialization when stack becomes active
		log.Info("Transitioning from Inactive to Initialization", "prNumber", prStack.Spec.PRNumber)
		r.Recorder.Event(prStack, corev1.EventTypeNormal, EventTypeActivating,
			fmt.Sprintf("PR #%s stack becoming active, scaling up services", prStack.Spec.PRNumber))
		prStack.Status.Phase = PhaseInitialization
		if err := r.Status().Update(ctx, prStack); err != nil {
			return ctrl.Result{}, err
		}
		return r.handleInitialization(ctx, prStack)
	case PhaseProvisioning:
		return r.handleProvisioning(ctx, prStack)
	case PhaseDeploying:
		return r.handleDeployment(ctx, prStack)
	case PhaseRunning:
		return r.handleRunning(ctx, prStack)
	case PhaseCleaning:
		return r.handleCleaning(ctx, prStack)
	default:
		log.Info("Unknown phase", "phase", prStack.Status.Phase)
		return ctrl.Result{}, nil
	}
}

// updateStatusWithError updates PRStack status with error information
func (r *PRStackReconciler) updateStatusWithError(ctx context.Context, prStack *pishopv1alpha1.PRStack, message string, err error) {
	prStack.Status.Message = fmt.Sprintf("%s: %v", message, err)
	r.Status().Update(ctx, prStack)
}

// recordProvisioningError records a provisioning error event and updates status
func (r *PRStackReconciler) recordProvisioningError(ctx context.Context, prStack *pishopv1alpha1.PRStack, component string, err error) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	log.Error(err, fmt.Sprintf("Failed to provision %s", component))
	r.Recorder.Event(prStack, corev1.EventTypeWarning, EventTypeProvisioningFailed,
		fmt.Sprintf("%s provisioning failed: %v", component, err))
	r.updateStatusWithError(ctx, prStack, fmt.Sprintf("%s provisioning failed", component), err)
	return ctrl.Result{RequeueAfter: RequeueIntervalMedium}, err
}
