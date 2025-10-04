package controllers

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	pishopv1alpha1 "go.pilab.hu/shop/pishop-provisioner/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (r *PRStackReconciler) deployService(ctx context.Context, prStack *pishopv1alpha1.PRStack, serviceName string) error {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Deploying service", "service", serviceName, "prNumber", prStack.Spec.PRNumber)

	// Create namespace for this PR if it doesn't exist
	namespaceName := fmt.Sprintf("pr-%s-shop-pilab-hu", prStack.Spec.PRNumber)
	if err := r.createNamespace(ctx, namespaceName); err != nil {
		return err
	}

	// Deploy the service
	switch serviceName {
	case "product-service":
		return r.deployProductService(ctx, prStack, namespaceName)
	case "cart-service":
		return r.deployCartService(ctx, prStack, namespaceName)
	case "order-service":
		return r.deployOrderService(ctx, prStack, namespaceName)
	case "payment-service":
		return r.deployPaymentService(ctx, prStack, namespaceName)
	case "customer-service":
		return r.deployCustomerService(ctx, prStack, namespaceName)
	case "inventory-service":
		return r.deployInventoryService(ctx, prStack, namespaceName)
	case "notification-service":
		return r.deployNotificationService(ctx, prStack, namespaceName)
	case "discount-service":
		return r.deployDiscountService(ctx, prStack, namespaceName)
	case "checkout-service":
		return r.deployCheckoutService(ctx, prStack, namespaceName)
	case "analytics-service":
		return r.deployAnalyticsService(ctx, prStack, namespaceName)
	case "auth-service":
		return r.deployAuthService(ctx, prStack, namespaceName)
	case "graphql-service":
		return r.deployGraphQLService(ctx, prStack, namespaceName)
	default:
		return fmt.Errorf("unknown service: %s", serviceName)
	}
}

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

func (r *PRStackReconciler) deployProductService(ctx context.Context, prStack *pishopv1alpha1.PRStack, namespace string) error {
	// Create ConfigMap for product service
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "product-service-config",
			Namespace: namespace,
		},
		Data: map[string]string{
			"MONGODB_URI":  prStack.Status.MongoDB.ConnectionString,
			"NATS_URL":     prStack.Status.NATS.ConnectionString,
			"REDIS_URL":    prStack.Status.Redis.ConnectionString,
			"SERVICE_NAME": "product-service",
			"ENVIRONMENT":  prStack.Spec.Environment,
			"PR_NUMBER":    prStack.Spec.PRNumber,
		},
	}

	if err := r.CreateOrUpdate(ctx, configMap); err != nil {
		return err
	}

	// Create Deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "product-service",
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "product-service"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "product-service"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "product-service",
							Image: "ghcr.io/pilab-dev/shop-product-service:latest",
							Ports: []corev1.ContainerPort{
								{ContainerPort: 8080},
							},
							EnvFrom: []corev1.EnvFromSource{
								{
									ConfigMapRef: &corev1.ConfigMapEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "product-service-config",
										},
									},
								},
							},
							Resources: r.getResourceRequirements(prStack),
						},
					},
				},
			},
		},
	}

	if err := r.CreateOrUpdate(ctx, deployment); err != nil {
		return err
	}

	// Create Service
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "product-service",
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": "product-service"},
			Ports: []corev1.ServicePort{
				{
					Port:       80,
					TargetPort: intstr.FromInt(8080),
				},
			},
		},
	}

	if err := r.CreateOrUpdate(ctx, service); err != nil {
		return err
	}

	// Create Ingress
	ingress := r.createIngress(prStack, namespace, "product-service", "product")
	if err := r.CreateOrUpdate(ctx, ingress); err != nil {
		return err
	}

	return nil
}

func (r *PRStackReconciler) deployCartService(ctx context.Context, prStack *pishopv1alpha1.PRStack, namespace string) error {
	// Similar implementation for cart service
	// ... (implementation details)
	return nil
}

func (r *PRStackReconciler) deployOrderService(ctx context.Context, prStack *pishopv1alpha1.PRStack, namespace string) error {
	// Similar implementation for order service
	// ... (implementation details)
	return nil
}

func (r *PRStackReconciler) deployPaymentService(ctx context.Context, prStack *pishopv1alpha1.PRStack, namespace string) error {
	// Similar implementation for payment service
	// ... (implementation details)
	return nil
}

func (r *PRStackReconciler) deployCustomerService(ctx context.Context, prStack *pishopv1alpha1.PRStack, namespace string) error {
	// Similar implementation for customer service
	// ... (implementation details)
	return nil
}

func (r *PRStackReconciler) deployInventoryService(ctx context.Context, prStack *pishopv1alpha1.PRStack, namespace string) error {
	// Similar implementation for inventory service
	// ... (implementation details)
	return nil
}

func (r *PRStackReconciler) deployNotificationService(ctx context.Context, prStack *pishopv1alpha1.PRStack, namespace string) error {
	// Similar implementation for notification service
	// ... (implementation details)
	return nil
}

func (r *PRStackReconciler) deployDiscountService(ctx context.Context, prStack *pishopv1alpha1.PRStack, namespace string) error {
	// Similar implementation for discount service
	// ... (implementation details)
	return nil
}

func (r *PRStackReconciler) deployCheckoutService(ctx context.Context, prStack *pishopv1alpha1.PRStack, namespace string) error {
	// Similar implementation for checkout service
	// ... (implementation details)
	return nil
}

func (r *PRStackReconciler) deployAnalyticsService(ctx context.Context, prStack *pishopv1alpha1.PRStack, namespace string) error {
	// Similar implementation for analytics service
	// ... (implementation details)
	return nil
}

func (r *PRStackReconciler) deployAuthService(ctx context.Context, prStack *pishopv1alpha1.PRStack, namespace string) error {
	// Similar implementation for auth service
	// ... (implementation details)
	return nil
}

func (r *PRStackReconciler) deployGraphQLService(ctx context.Context, prStack *pishopv1alpha1.PRStack, namespace string) error {
	// Similar implementation for graphql service
	// ... (implementation details)
	return nil
}

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

func (r *PRStackReconciler) CreateOrUpdate(ctx context.Context, obj client.Object) error {
	if err := r.Create(ctx, obj); err != nil {
		if errors.IsAlreadyExists(err) {
			return r.Update(ctx, obj)
		}
		return err
	}
	return nil
}
