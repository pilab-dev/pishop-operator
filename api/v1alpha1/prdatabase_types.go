//go:generate controller-gen object:headerFile="../../hack/boilerplate.go.txt" paths="."
//go:generate controller-gen crd paths="./..." output:crd:dir=../../config/crd/bases

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PRStackSpec defines the desired state of PRStack
type PRStackSpec struct {
	// PRNumber is the pull request number
	PRNumber string `json:"prNumber"`

	// ImageTag is the Docker image tag to use for services (e.g., pr-33-abc123, v1.2.3, latest)
	// If not specified, defaults to pr-{prNumber}
	ImageTag string `json:"imageTag,omitempty"`

	// CustomDomain is a custom domain for the ingress (e.g., gyurushop.hu, magicshop.hu)
	// If not specified, defaults to pr-{prNumber}.shop.pilab.hu
	CustomDomain string `json:"customDomain,omitempty"`

	// IngressTlsSecretName is the name of the Kubernetes secret containing the TLS certificate
	// for the custom domain. The secret should contain 'tls.crt' and 'tls.key' keys.
	// If not specified, no TLS configuration will be added to the ingress.
	IngressTlsSecretName string `json:"ingressTlsSecretName,omitempty"`

	// Active controls whether the stack is active (replicas > 0) or inactive (replicas = 0)
	// When false, all deployments are scaled to 0 replicas
	// When true, all deployments are scaled to 1 replica
	// +kubebuilder:default=true
	Active bool `json:"active,omitempty"`

	// DeployedAt is a timestamp that triggers a rollout of all deployments when changed
	// Update this field to force a re-deployment of all services
	DeployedAt *metav1.Time `json:"deployedAt,omitempty"`

	// MongoDB connection details
	MongoURI      string `json:"mongoURI,omitempty"`
	MongoUsername string `json:"mongoUsername,omitempty"`
	MongoPassword string `json:"mongoPassword,omitempty"`

	// NATS connection details
	NatsURL string `json:"natsURL,omitempty"`

	// Redis connection details
	RedisURL string `json:"redisURL,omitempty"`

	// Services to provision for this PR
	Services []string `json:"services,omitempty"`

	// Environment configuration
	Environment string `json:"environment,omitempty"`

	// Resource limits for the PR environment
	ResourceLimits *ResourceLimits `json:"resourceLimits,omitempty"`

	// Backup configuration
	BackupConfig *BackupConfig `json:"backupConfig,omitempty"`
}

// ResourceLimits defines resource constraints for the PR environment
type ResourceLimits struct {
	// CPU limit per service
	CPULimit string `json:"cpuLimit,omitempty"`
	// Memory limit per service
	MemoryLimit string `json:"memoryLimit,omitempty"`
	// Storage limit for databases
	StorageLimit string `json:"storageLimit,omitempty"`
}

// BackupConfig defines backup configuration for the PR environment
type BackupConfig struct {
	// Enabled controls whether automatic backups are enabled
	Enabled bool `json:"enabled,omitempty"`
	// Schedule defines the backup schedule in cron format
	Schedule string `json:"schedule,omitempty"`
	// RetentionDays defines how many days to keep backups
	RetentionDays int `json:"retentionDays,omitempty"`
	// StorageClass defines the storage class for backup PVCs
	StorageClass string `json:"storageClass,omitempty"`
	// StorageSize defines the size of backup storage
	StorageSize string `json:"storageSize,omitempty"`
}

// PRStackStatus defines the observed state of PRStack
type PRStackStatus struct {
	// Phase represents the current phase of the PR stack
	Phase string `json:"phase,omitempty"`

	// Message provides additional information about the current status
	Message string `json:"message,omitempty"`

	// CreatedAt is the timestamp when the stack was first created
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`

	// LastActiveAt is the timestamp of the last activity on the stack
	LastActiveAt *metav1.Time `json:"lastActiveAt,omitempty"`

	// LastDeployedAt is the timestamp when deployments were last rolled out
	LastDeployedAt *metav1.Time `json:"lastDeployedAt,omitempty"`

	// MongoDB credentials for this PR
	MongoDB *MongoDBCredentials `json:"mongodb,omitempty"`

	// NATS configuration for this PR
	NATS *NATSConfig `json:"nats,omitempty"`

	// Redis configuration for this PR
	Redis *RedisConfig `json:"redis,omitempty"`

	// Deployed services
	Services []ServiceStatus `json:"services,omitempty"`

	// Conditions represent the latest available observations of the object's state
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Backup status
	Backup *BackupStatus `json:"backup,omitempty"`
}

// MongoDBCredentials contains MongoDB connection details for the PR
type MongoDBCredentials struct {
	// PRUser is the created MongoDB user for this PR
	User string `json:"user,omitempty"`
	// PRPassword is the generated password for the PR user
	Password string `json:"password,omitempty"`
	// Connection string for the PR databases
	ConnectionString string `json:"connectionString,omitempty"`
	// Databases lists the created databases
	Databases []string `json:"databases,omitempty"`
}

// NATSConfig contains NATS configuration for the PR
type NATSConfig struct {
	// Subject prefix for this PR
	SubjectPrefix string `json:"subjectPrefix,omitempty"`
	// Connection string for NATS
	ConnectionString string `json:"connectionString,omitempty"`
}

// RedisConfig contains Redis configuration for the PR
type RedisConfig struct {
	// Key prefix for this PR
	KeyPrefix string `json:"keyPrefix,omitempty"`
	// Connection string for Redis
	ConnectionString string `json:"connectionString,omitempty"`
}

// ServiceStatus represents the status of a deployed service
type ServiceStatus struct {
	// Name of the service
	Name string `json:"name"`
	// Status of the service (Running, Failed, Pending)
	Status string `json:"status"`
	// URL of the service
	URL string `json:"url,omitempty"`
	// Message about the service status
	Message string `json:"message,omitempty"`
}

// BackupStatus represents the backup status for the PR stack
type BackupStatus struct {
	// LastBackupTime is the timestamp of the last successful backup
	LastBackupTime *metav1.Time `json:"lastBackupTime,omitempty"`
	// LastBackupName is the name of the last successful backup
	LastBackupName string `json:"lastBackupName,omitempty"`
	// BackupCount is the total number of backups available
	BackupCount int `json:"backupCount,omitempty"`
	// LastBackupSize is the size of the last backup
	LastBackupSize string `json:"lastBackupSize,omitempty"`
	// BackupJobs tracks running backup/restore jobs
	BackupJobs []BackupJobStatus `json:"backupJobs,omitempty"`
}

// BackupJobStatus represents the status of a backup or restore job
type BackupJobStatus struct {
	// Name of the job
	Name string `json:"name"`
	// Type of job (backup, restore)
	Type string `json:"type"`
	// Status of the job (Running, Completed, Failed)
	Status string `json:"status"`
	// StartTime is when the job started
	StartTime *metav1.Time `json:"startTime,omitempty"`
	// CompletionTime is when the job completed
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`
	// Message about the job status
	Message string `json:"message,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
//+kubebuilder:printcolumn:name="PR Number",type="string",JSONPath=".spec.prNumber"
//+kubebuilder:printcolumn:name="Environment",type="string",JSONPath=".spec.environment"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// PRStack is the Schema for the prstacks API
type PRStack struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PRStackSpec   `json:"spec,omitempty"`
	Status PRStackStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PRStackList contains a list of PRStack
type PRStackList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PRStack `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PRStack{}, &PRStackList{})
}
