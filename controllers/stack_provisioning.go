
package controllers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	pishopv1alpha1 "go.pilab.hu/shop/pishop-provisioner/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// createMongoDBResources creates the MongoDB ConfigMap and Secret for the PR namespace
func (r *PRStackReconciler) createMongoDBResources(ctx context.Context, prStack *pishopv1alpha1.PRStack, namespace string) error {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Creating MongoDB resources for PR", "prNumber", prStack.Spec.PRNumber, "namespace", namespace)

	// Create MongoDB ConfigMap (based on k8s/base/mongodb-external.yaml)
	mongodbConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mongodb-config",
			Namespace: namespace,
		},
		Data: func() map[string]string {
			data := getDatabaseConfigMapData(prStack.Spec.PRNumber)
			data["uri"] = prStack.Status.MongoDB.ConnectionString
			return data
		}(),
	}

	if err := r.CreateOrUpdate(ctx, mongodbConfigMap); err != nil {
		return fmt.Errorf("failed to create MongoDB ConfigMap: %v", err)
	}

	// Create MongoDB Secret (based on k8s/base/mongodb-external.yaml)
	mongodbSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mongodb-secret",
			Namespace: namespace,
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"uri":                 prStack.Status.MongoDB.ConnectionString,
			"username":            prStack.Status.MongoDB.User,
			"password":            prStack.Status.MongoDB.Password,
			"product-db-uri":      fmt.Sprintf("%s/pishop_products_pr_%s", prStack.Status.MongoDB.ConnectionString, prStack.Spec.PRNumber),
			"cart-db-uri":         fmt.Sprintf("%s/pishop_cart_pr_%s", prStack.Status.MongoDB.ConnectionString, prStack.Spec.PRNumber),
			"order-db-uri":        fmt.Sprintf("%s/pishop_orders_pr_%s", prStack.Status.MongoDB.ConnectionString, prStack.Spec.PRNumber),
			"payment-db-uri":      fmt.Sprintf("%s/pishop_payments_pr_%s", prStack.Status.MongoDB.ConnectionString, prStack.Spec.PRNumber),
			"customer-db-uri":     fmt.Sprintf("%s/pishop_customers_pr_%s", prStack.Status.MongoDB.ConnectionString, prStack.Spec.PRNumber),
			"inventory-db-uri":    fmt.Sprintf("%s/pishop_inventory_pr_%s", prStack.Status.MongoDB.ConnectionString, prStack.Spec.PRNumber),
			"notification-db-uri": fmt.Sprintf("%s/pishop_notifications_pr_%s", prStack.Status.MongoDB.ConnectionString, prStack.Spec.PRNumber),
			"discount-db-uri":     fmt.Sprintf("%s/pishop_discounts_pr_%s", prStack.Status.MongoDB.ConnectionString, prStack.Spec.PRNumber),
			"checkout-db-uri":     fmt.Sprintf("%s/pishop_checkout_pr_%s", prStack.Status.MongoDB.ConnectionString, prStack.Spec.PRNumber),
			"analytics-db-uri":    fmt.Sprintf("%s/pishop_analytics_pr_%s", prStack.Status.MongoDB.ConnectionString, prStack.Spec.PRNumber),
			"auth-db-uri":         fmt.Sprintf("%s/pishop_auth_pr_%s", prStack.Status.MongoDB.ConnectionString, prStack.Spec.PRNumber),
			"graphql-db-uri":      fmt.Sprintf("%s/pishop_graphql_pr_%s", prStack.Status.MongoDB.ConnectionString, prStack.Spec.PRNumber),
		},
	}

	if err := r.CreateOrUpdate(ctx, mongodbSecret); err != nil {
		return fmt.Errorf("failed to create MongoDB Secret: %v", err)
	}

	// Create NATS ConfigMap
	natsConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nats-config",
			Namespace: namespace,
		},
		Data: map[string]string{
			"url": prStack.Status.NATS.ConnectionString,
		},
	}

	if err := r.CreateOrUpdate(ctx, natsConfigMap); err != nil {
		return fmt.Errorf("failed to create NATS ConfigMap: %v", err)
	}

	// Create Redis ConfigMap
	redisConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "redis-config",
			Namespace: namespace,
		},
		Data: map[string]string{
			"url": prStack.Status.Redis.ConnectionString,
		},
	}

	if err := r.CreateOrUpdate(ctx, redisConfigMap); err != nil {
		return fmt.Errorf("failed to create Redis ConfigMap: %v", err)
	}

	log.Info("Successfully created MongoDB, NATS, and Redis resources", "namespace", namespace)
	return nil
}

// createRegistrySecret creates a docker-registry secret for pulling images from GitHub Container Registry
func (r *PRStackReconciler) createRegistrySecret(ctx context.Context, namespace string) error {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Creating registry secret for namespace", "namespace", namespace)

	if r.GitHubUsername == "" || r.GitHubToken == "" {
		log.Info("GitHub credentials not configured, skipping registry secret creation")
		return nil
	}

	// Create docker config JSON
	type dockerConfigEntry struct {
		Username string `json:"username"`
		Email    string `json:"email,omitempty"`
		Password string `json:"password"`
		Auth     string `json:"auth"`
	}

	type dockerConfig struct {
		Auths map[string]dockerConfigEntry `json:"auths"`
	}

	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", r.GitHubUsername, r.GitHubToken)))
	dockerCfg := dockerConfig{
		Auths: map[string]dockerConfigEntry{
			"ghcr.io": {
				Username: r.GitHubUsername,
				Password: r.GitHubToken,
				Email:    r.GitHubEmail,
				Auth:     auth,
			},
		},
	}

	dockerCfgJSON, err := json.Marshal(dockerCfg)
	if err != nil {
		return fmt.Errorf("failed to marshal docker config: %v", err)
	}

	// Create the secret
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ghcr-secret",
			Namespace: namespace,
		},
		Type: corev1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			corev1.DockerConfigJsonKey: dockerCfgJSON,
		},
	}

	if err := r.CreateOrUpdate(ctx, secret); err != nil {
		return fmt.Errorf("failed to create registry secret: %v", err)
	}

	log.Info("Successfully created registry secret", "namespace", namespace)
	return nil
}

// createServiceDeployment creates a deployment for a specific service based on the K8s templates
func (r *PRStackReconciler) createServiceDeployment(ctx context.Context, prStack *pishopv1alpha1.PRStack, namespace string, serviceName string) error {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Creating deployment for service", "service", serviceName, "namespace", namespace)

	// Get resource requirements
	resourceRequirements := r.getResourceRequirements(prStack)

	// Determine replica count based on Active flag (max 1 replica)
	replicas := int32(1)
	if !prStack.Spec.Active {
		replicas = 0
	}

	// Create deployment based on the service
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: namespace,
			Labels: map[string]string{
				"app":         serviceName,
				"component":   "microservice",
				"environment": "pr",
				"tier":        "pr",
				"pr-number":   prStack.Spec.PRNumber,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": serviceName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":         serviceName,
						"component":   "microservice",
						"environment": "pr",
						"tier":        "pr",
						"pr-number":   prStack.Spec.PRNumber,
					},
				},
				Spec: corev1.PodSpec{
					ImagePullSecrets: []corev1.LocalObjectReference{
						{Name: "ghcr-secret"},
					},
					Containers: []corev1.Container{
						{
							Name:      serviceName,
							Image:     getImageTag(prStack, serviceName),
							Ports:     GetServicePorts(),
							Env:       GetServiceConfig(serviceName, prStack.Spec.PRNumber).ToEnvVars(),
							Resources: resourceRequirements,
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/health",
										Port: intstr.FromInt(8080),
									},
								},
								InitialDelaySeconds: 30,
								PeriodSeconds:       10,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/ready",
										Port: intstr.FromInt(8080),
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:       5,
							},
						},
					},
				},
			},
		},
	}

	if err := r.CreateOrUpdate(ctx, deployment); err != nil {
		return fmt.Errorf("failed to create deployment for %s: %v", serviceName, err)
	}

	// Create service
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: namespace,
			Labels: map[string]string{
				"app":         serviceName,
				"environment": "pr",
				"tier":        "pr",
				"pr-number":   prStack.Spec.PRNumber,
			},
			Annotations: map[string]string{
				"prometheus.io/scrape": "true",
				"prometheus.io/port":   "8080",
				"prometheus.io/path":   "/metrics",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": serviceName},
			Ports: []corev1.ServicePort{
				{Name: "http", Port: 8080, TargetPort: intstr.FromInt(8080)},
			},
		},
	}

	if err := r.CreateOrUpdate(ctx, service); err != nil {
		return fmt.Errorf("failed to create service for %s: %v", serviceName, err)
	}

	// Create ingress only for GraphQL service
	if serviceName == "graphql-service" {
		ingress := r.createIngress(prStack, namespace, serviceName, "graphql")
		if err := r.CreateOrUpdate(ctx, ingress); err != nil {
			return fmt.Errorf("failed to create ingress for %s: %v", serviceName, err)
		}
	}

	log.Info("Successfully created deployment and service", "service", serviceName)
	return nil
}

// getResourceRequirements returns resource requirements for a service based on PRStack configuration
func (r *PRStackReconciler) getResourceRequirements(prStack *pishopv1alpha1.PRStack) corev1.ResourceRequirements {
	limits := corev1.ResourceList{}
	requests := corev1.ResourceList{}

	if prStack.Spec.ResourceLimits != nil {
		if prStack.Spec.ResourceLimits.CPULimit != "" {
			limits["cpu"] = resource.MustParse(prStack.Spec.ResourceLimits.CPULimit)
		}
		if prStack.Spec.ResourceLimits.MemoryLimit != "" {
			limits["memory"] = resource.MustParse(prStack.Spec.ResourceLimits.MemoryLimit)
		}
	} else {
		// Default limits
		limits["cpu"] = resource.MustParse("500m")
		limits["memory"] = resource.MustParse("512Mi")
	}

	// Set zero requests to prevent cluster over-allocation
	requests["cpu"] = resource.MustParse("0m")
	requests["memory"] = resource.MustParse("0Mi")

	return corev1.ResourceRequirements{
		Limits:   limits,
		Requests: requests,
	}
}

// CreateOrUpdate creates or updates a Kubernetes resource
func (r *PRStackReconciler) CreateOrUpdate(ctx context.Context, obj client.Object) error {
	if err := r.Create(ctx, obj); err != nil {
		if errors.IsAlreadyExists(err) {
			return r.Update(ctx, obj)
		}
		return err
	}
	return nil
}

// createNamespace creates a namespace if it doesn't exist
func (r *PRStackReconciler) createNamespace(ctx context.Context, namespaceName string) error {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
		},
	}

	if err := r.Get(ctx, types.NamespacedName{Name: namespaceName}, namespace); err != nil {
		if errors.IsNotFound(err) {
			if err := r.Create(ctx, namespace); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

// getImageTag returns the Docker image tag to use for a specific service
// If ImageTag is specified in the PRStack spec, use it
// Otherwise, default to pr-{prNumber}
func getImageTag(prStack *pishopv1alpha1.PRStack, serviceName string) string {
	if prStack.Spec.ImageTag != "" {
		return fmt.Sprintf("ghcr.io/pilab-dev/%s:%s", serviceName, prStack.Spec.ImageTag)
	}
	return fmt.Sprintf("ghcr.io/pilab-dev/%s:pr-%s", serviceName, prStack.Spec.PRNumber)
}

// createIngress creates an ingress resource for the GraphQL service
func (r *PRStackReconciler) createIngress(prStack *pishopv1alpha1.PRStack, namespace, serviceName, pathPrefix string) *networkingv1.Ingress {
	hostname := r.getDomain(prStack)

	// Set default values if not configured
	ingressClassName := r.IngressClassName
	if ingressClassName == "" {
		ingressClassName = "traefik"
	}
	certManagerIssuer := r.CertManagerIssuer
	traefikEntrypoints := r.TraefikEntrypoints
	if traefikEntrypoints == "" {
		traefikEntrypoints = "websecure"
	}
	traefikTLSEnabled := r.TraefikTLSEnabled
	if traefikTLSEnabled == "" {
		traefikTLSEnabled = "true"
	}

	// Determine if we should enable TLS
	enableTLS := prStack.Spec.IngressTlsSecretName != "" || certManagerIssuer != ""
	
	// Determine the TLS secret name
	var tlsSecretName string
	if prStack.Spec.IngressTlsSecretName != "" {
		tlsSecretName = prStack.Spec.IngressTlsSecretName
	} else if certManagerIssuer != "" {
		tlsSecretName = fmt.Sprintf("%s-tls", serviceName)
	}

	// Build annotations
	annotations := make(map[string]string)
	if certManagerIssuer != "" {
		annotations["cert-manager.io/cluster-issuer"] = certManagerIssuer
	}
	if traefikEntrypoints != "" {
		annotations["traefik.ingress.kubernetes.io/router.entrypoints"] = traefikEntrypoints
	}
	if traefikTLSEnabled != "" {
		annotations["traefik.ingress.kubernetes.io/router.tls"] = traefikTLSEnabled
	}

	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        serviceName,
			Namespace:   namespace,
			Annotations: annotations,
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: &ingressClassName,
			Rules: []networkingv1.IngressRule{
				{
					Host: hostname,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     fmt.Sprintf("/%s", pathPrefix),
									PathType: func() *networkingv1.PathType { pt := networkingv1.PathTypePrefix; return &pt }(),
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: serviceName,
											Port: networkingv1.ServiceBackendPort{
												Number: 8080,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Add TLS configuration if enabled
	if enableTLS {
		ingress.Spec.TLS = []networkingv1.IngressTLS{
			{
				Hosts:      []string{hostname},
				SecretName: tlsSecretName,
			},
		}
	}

	return ingress
}
