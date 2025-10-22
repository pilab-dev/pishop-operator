package controllers

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	pishopv1alpha1 "go.pilab.hu/shop/pishop-provisioner/api/v1alpha1"
)

func (r *PRStackReconciler) cleanupServices(ctx context.Context, prStack *pishopv1alpha1.PRStack) error {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Cleaning up services for PR", "prNumber", prStack.Spec.PRNumber)

	// Delete the PR namespace
	namespaceName := r.getNamespaceName(prStack.Spec.PRNumber)
	namespace := &corev1.Namespace{}
	if err := r.Get(ctx, client.ObjectKey{Name: namespaceName}, namespace); err == nil {
		if err := r.Delete(ctx, namespace); err != nil {
			log.Error(err, "Failed to delete namespace", "namespace", namespaceName)
			return err
		}
		log.Info("Successfully deleted namespace", "namespace", namespaceName)
	}

	return nil
}

func (r *PRStackReconciler) cleanupMongoDB(ctx context.Context, prStack *pishopv1alpha1.PRStack) error {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Cleaning up MongoDB for PR", "prNumber", prStack.Spec.PRNumber)

	// Connect to MongoDB
	mongoURI := r.MongoURI
	if prStack.Spec.MongoURI != "" {
		mongoURI = prStack.Spec.MongoURI
	}

	client, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	// Test connection
	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %v", err)
	}

	// Drop databases for each service (same logic as main.go)
	services := DefaultServices

	for _, service := range services {
		dbName := fmt.Sprintf("pishop_%s_pr_%s", service, prStack.Spec.PRNumber)
		log.Info("Dropping database", "database", dbName)

		database := client.Database(dbName)
		if err := database.Drop(ctx); err != nil {
			// Check if the error is because the database doesn't exist using MongoDB error codes
			if serverErr, ok := err.(mongo.ServerError); ok && serverErr.HasErrorCode(26) { // NamespaceNotFound
				// Database not found is expected during cleanup, log as info
				log.Info("Database not found during cleanup (expected)", "database", dbName)
			} else {
				// Other errors should be logged as errors
				log.Error(err, "Failed to drop database", "database", dbName)
			}
			// Continue with other databases even if one fails
		} else {
			log.Info("Successfully dropped database", "database", dbName)
		}
	}

	// Clean up the PR user (same logic as main.go)
	prUser := fmt.Sprintf("pishop_pr_%s", prStack.Spec.PRNumber)
	log.Info("Cleaning up PR user", "user", prUser)

	adminDB := client.Database("admin")
	if err := adminDB.RunCommand(ctx, bson.M{"dropUser": prUser}).Err(); err != nil {
		// Check if the error is because the user doesn't exist using MongoDB error codes
		if serverErr, ok := err.(mongo.ServerError); ok && serverErr.HasErrorCode(11) { // UserNotFound
			// User not found is expected during cleanup, log as info
			log.Info("User not found during cleanup (expected)", "user", prUser)
		} else {
			// Other errors should be logged as errors
			log.Error(err, "Failed to drop user", "user", prUser)
		}
		// Don't return error as user might not exist
	} else {
		log.Info("Successfully dropped user", "user", prUser)
	}

	return nil
}

func (r *PRStackReconciler) cleanupNATS(ctx context.Context, prStack *pishopv1alpha1.PRStack) error {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Cleaning up NATS for PR", "prNumber", prStack.Spec.PRNumber)

	namespaceName := fmt.Sprintf("pr-%s-shop-pilab-hu", prStack.Spec.PRNumber)

	// Delete NATS deployment
	natsDeployment := &appsv1.Deployment{}
	if err := r.Get(ctx, client.ObjectKey{Name: "nats", Namespace: namespaceName}, natsDeployment); err == nil {
		if err := r.Delete(ctx, natsDeployment); err != nil {
			log.Error(err, "Failed to delete NATS deployment")
			return err
		}
		log.Info("Successfully deleted NATS deployment")
	}

	// Delete NATS service
	natsService := &corev1.Service{}
	if err := r.Get(ctx, client.ObjectKey{Name: "nats", Namespace: namespaceName}, natsService); err == nil {
		if err := r.Delete(ctx, natsService); err != nil {
			log.Error(err, "Failed to delete NATS service")
			return err
		}
		log.Info("Successfully deleted NATS service")
	}

	subjectPrefix := fmt.Sprintf("pishop.pr.%s", prStack.Spec.PRNumber)
	log.Info("NATS cleanup completed for subject prefix", "prefix", subjectPrefix)

	return nil
}

func (r *PRStackReconciler) cleanupRedis(ctx context.Context, prStack *pishopv1alpha1.PRStack) error {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Cleaning up Redis for PR", "prNumber", prStack.Spec.PRNumber)

	namespaceName := fmt.Sprintf("pr-%s-shop-pilab-hu", prStack.Spec.PRNumber)

	// Delete Redis deployment
	redisDeployment := &appsv1.Deployment{}
	if err := r.Get(ctx, client.ObjectKey{Name: "redis", Namespace: namespaceName}, redisDeployment); err == nil {
		if err := r.Delete(ctx, redisDeployment); err != nil {
			log.Error(err, "Failed to delete Redis deployment")
			return err
		}
		log.Info("Successfully deleted Redis deployment")
	}

	// Delete Redis service
	redisService := &corev1.Service{}
	if err := r.Get(ctx, client.ObjectKey{Name: "redis", Namespace: namespaceName}, redisService); err == nil {
		if err := r.Delete(ctx, redisService); err != nil {
			log.Error(err, "Failed to delete Redis service")
			return err
		}
		log.Info("Successfully deleted Redis service")
	}

	keyPrefix := fmt.Sprintf("pishop:pr:%s:", prStack.Spec.PRNumber)
	log.Info("Redis cleanup completed for key prefix", "prefix", keyPrefix)

	return nil
}

// cleanupAllResources performs complete cleanup of all PR resources
func (r *PRStackReconciler) cleanupAllResources(ctx context.Context, prStack *pishopv1alpha1.PRStack) error {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Starting complete cleanup for PR", "prNumber", prStack.Spec.PRNumber)

	// Clean up MongoDB databases and users
	if err := r.cleanupMongoDB(ctx, prStack); err != nil {
		log.Error(err, "Failed to cleanup MongoDB")
		// Continue with other cleanup even if MongoDB cleanup fails
	}

	// Clean up NATS
	if err := r.cleanupNATS(ctx, prStack); err != nil {
		log.Error(err, "Failed to cleanup NATS")
		// Continue with other cleanup even if NATS cleanup fails
	}

	// Clean up Redis
	if err := r.cleanupRedis(ctx, prStack); err != nil {
		log.Error(err, "Failed to cleanup Redis")
		// Continue with other cleanup even if Redis cleanup fails
	}

	// Clean up Kubernetes services and namespace
	if err := r.cleanupServices(ctx, prStack); err != nil {
		log.Error(err, "Failed to cleanup services")
		return err // This is critical, so we return the error
	}

	log.Info("Successfully completed cleanup for PR", "prNumber", prStack.Spec.PRNumber)
	return nil
}
