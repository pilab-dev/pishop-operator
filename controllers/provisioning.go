package controllers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	pishopv1alpha1 "go.pilab.hu/shop/pishop-provisioner/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (r *PRStackReconciler) provisionMongoDB(ctx context.Context, prStack *pishopv1alpha1.PRStack) error {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Provisioning MongoDB for PR", "prNumber", prStack.Spec.PRNumber)

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

	// Generate PR user credentials
	prUser := fmt.Sprintf("pishop_pr_%s", prStack.Spec.PRNumber)
	prPassword, err := generateSecurePassword()
	if err != nil {
		return fmt.Errorf("failed to generate secure password: %v", err)
	}

	// Create PR user with limited permissions
	adminDB := client.Database("admin")
	
	// Build roles dynamically from DefaultServices
	services := DefaultServices
	var roles []bson.M
	for _, service := range services {
		dbName := getDatabaseName(service, prStack.Spec.PRNumber)
		roles = append(roles, bson.M{"role": "readWrite", "db": dbName})
	}

	// Drop user if exists (ignore errors)
	adminDB.RunCommand(ctx, bson.D{{Key: "dropUser", Value: prUser}})

	// Create user
	if err := adminDB.RunCommand(ctx, bson.D{
		{Key: "createUser", Value: prUser},
		{Key: "pwd", Value: prPassword},
		{Key: "roles", Value: roles},
	}).Err(); err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}

	// Create databases and collections
	var databases []string

	for _, service := range services {
		dbName := getDatabaseName(service, prStack.Spec.PRNumber)
		databases = append(databases, dbName)

		database := client.Database(dbName)

		// Create collections based on service
		if err := r.createServiceCollections(ctx, database, service); err != nil {
			return fmt.Errorf("failed to create collections for %s: %v", service, err)
		}
	}

	// Update status
	prStack.Status.MongoDB = &pishopv1alpha1.MongoDBCredentials{
		User:             prUser,
		Password:         prPassword,
		ConnectionString: fmt.Sprintf("mongodb://%s:%s@mongodb.pishop-base.svc.cluster.local:27017", prUser, prPassword),
		Databases:        databases,
	}

	// Update the PRStack status
	if err := r.Status().Update(ctx, prStack); err != nil {
		return fmt.Errorf("failed to update PRStack status: %v", err)
	}

	return nil
}

func (r *PRStackReconciler) provisionNATS(ctx context.Context, prStack *pishopv1alpha1.PRStack) error {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Provisioning NATS for PR", "prNumber", prStack.Spec.PRNumber)

	// Set up subject prefixes and connection details
	// NATS resources will be created in the deployment phase
	subjectPrefix := fmt.Sprintf("pishop.pr.%s", prStack.Spec.PRNumber)
	namespaceName := fmt.Sprintf("pr-%s-shop-pilab-hu", prStack.Spec.PRNumber)
	natsURL := fmt.Sprintf("nats://nats.%s.svc.cluster.local:4222", namespaceName)

	// Update status
	prStack.Status.NATS = &pishopv1alpha1.NATSConfig{
		SubjectPrefix:    subjectPrefix,
		ConnectionString: natsURL,
	}

	log.Info("NATS configuration provisioned", "subjectPrefix", subjectPrefix, "connectionString", natsURL)
	return nil
}

func (r *PRStackReconciler) provisionRedis(ctx context.Context, prStack *pishopv1alpha1.PRStack) error {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Provisioning Redis for PR", "prNumber", prStack.Spec.PRNumber)

	// Set up key prefixes and connection details
	// Redis resources will be created in the deployment phase
	keyPrefix := fmt.Sprintf("pishop:pr:%s:", prStack.Spec.PRNumber)
	namespaceName := fmt.Sprintf("pr-%s-shop-pilab-hu", prStack.Spec.PRNumber)
	redisURL := fmt.Sprintf("redis://redis.%s.svc.cluster.local:6379", namespaceName)

	// Update status
	prStack.Status.Redis = &pishopv1alpha1.RedisConfig{
		KeyPrefix:        keyPrefix,
		ConnectionString: redisURL,
	}

	log.Info("Redis configuration provisioned", "keyPrefix", keyPrefix, "connectionString", redisURL)
	return nil
}

func (r *PRStackReconciler) createNATSResources(ctx context.Context, prStack *pishopv1alpha1.PRStack, namespace string) error {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Creating NATS resources", "namespace", namespace)

	// Create NATS deployment
	natsDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nats",
			Namespace: namespace,
			Labels: map[string]string{
				"app": "nats",
				"pr":  prStack.Spec.PRNumber,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "nats",
					"pr":  prStack.Spec.PRNumber,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "nats",
						"pr":  prStack.Spec.PRNumber,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nats",
							Image: "nats:2.10-alpine",
							Args: []string{
								"--jetstream",
								"--store_dir=/data",
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 4222,
									Name:          "client",
								},
								{
									ContainerPort: 8222,
									Name:          "monitor",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data",
									MountPath: "/data",
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("0m"),
									corev1.ResourceMemory: resource.MustParse("0Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("512Mi"),
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "data",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}

	if err := r.CreateOrUpdate(ctx, natsDeployment); err != nil {
		return fmt.Errorf("failed to create NATS deployment: %v", err)
	}

	// Create NATS service
	natsService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nats",
			Namespace: namespace,
			Labels: map[string]string{
				"app": "nats",
				"pr":  prStack.Spec.PRNumber,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "nats",
				"pr":  prStack.Spec.PRNumber,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "client",
					Port:       4222,
					TargetPort: intstr.FromInt(4222),
					Protocol:   corev1.ProtocolTCP,
				},
				{
					Name:       "monitor",
					Port:       8222,
					TargetPort: intstr.FromInt(8222),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}

	if err := r.CreateOrUpdate(ctx, natsService); err != nil {
		return fmt.Errorf("failed to create NATS service: %v", err)
	}

	log.Info("NATS resources created successfully", "namespace", namespace)
	return nil
}

func (r *PRStackReconciler) createRedisResources(ctx context.Context, prStack *pishopv1alpha1.PRStack, namespace string) error {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Creating Redis resources", "namespace", namespace)

	// Create Redis deployment
	redisDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "redis",
			Namespace: namespace,
			Labels: map[string]string{
				"app": "redis",
				"pr":  prStack.Spec.PRNumber,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "redis",
					"pr":  prStack.Spec.PRNumber,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "redis",
						"pr":  prStack.Spec.PRNumber,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "redis",
							Image: "redis:7-alpine",
							Args: []string{
								"redis-server",
								"--appendonly",
								"yes",
								"--maxmemory",
								"256mb",
								"--maxmemory-policy",
								"allkeys-lru",
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 6379,
									Name:          "redis",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data",
									MountPath: "/data",
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("0m"),
									corev1.ResourceMemory: resource.MustParse("0Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("512Mi"),
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "data",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}

	if err := r.CreateOrUpdate(ctx, redisDeployment); err != nil {
		return fmt.Errorf("failed to create Redis deployment: %v", err)
	}

	// Create Redis service
	redisService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "redis",
			Namespace: namespace,
			Labels: map[string]string{
				"app": "redis",
				"pr":  prStack.Spec.PRNumber,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "redis",
				"pr":  prStack.Spec.PRNumber,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "redis",
					Port:       6379,
					TargetPort: intstr.FromInt(6379),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}

	if err := r.CreateOrUpdate(ctx, redisService); err != nil {
		return fmt.Errorf("failed to create Redis service: %v", err)
	}

	log.Info("Redis resources created successfully", "namespace", namespace)
	return nil
}

func (r *PRStackReconciler) createServiceCollections(ctx context.Context, database *mongo.Database, service string) error {
	// Convert service name to collection name by removing "-service" suffix and handling special cases
	var collectionName string
	switch service {
	case "product-service":
		collectionName = "products"
	case "cart-service":
		collectionName = "cart"
	case "order-service":
		collectionName = "orders"
	case "payment-service":
		collectionName = "payments"
	case "customer-service":
		collectionName = "customers"
	case "inventory-service":
		collectionName = "inventory"
	case "notification-service":
		collectionName = "notifications"
	case "discount-service":
		collectionName = "discounts"
	case "checkout-service":
		collectionName = "checkout"
	case "analytics-service":
		collectionName = "analytics"
	case "auth-service":
		collectionName = "auth"
	case "graphql-service":
		collectionName = "graphql"
	default:
		return fmt.Errorf("unknown service: %s", service)
	}

	// Call the appropriate collection creation function
	switch collectionName {
	case "products":
		return r.createProductCollections(ctx, database)
	case "cart":
		return r.createCartCollections(ctx, database)
	case "orders":
		return r.createOrderCollections(ctx, database)
	case "payments":
		return r.createPaymentCollections(ctx, database)
	case "customers":
		return r.createCustomerCollections(ctx, database)
	case "inventory":
		return r.createInventoryCollections(ctx, database)
	case "notifications":
		return r.createNotificationCollections(ctx, database)
	case "discounts":
		return r.createDiscountCollections(ctx, database)
	case "checkout":
		return r.createCheckoutCollections(ctx, database)
	case "analytics":
		return r.createAnalyticsCollections(ctx, database)
	case "auth":
		return r.createAuthCollections(ctx, database)
	case "graphql":
		return r.createGraphQLCollections(ctx, database)
	default:
		return fmt.Errorf("unknown collection: %s", collectionName)
	}
}

func (r *PRStackReconciler) createProductCollections(ctx context.Context, database *mongo.Database) error {
	// Create products collection with indexes
	productsCollection := database.Collection("products")

	// Create indexes
	indexes := []mongo.IndexModel{
		{Keys: bson.M{"slug": 1}, Options: options.Index().SetUnique(true)},
		{Keys: bson.M{"sku": 1}, Options: options.Index().SetUnique(true)},
		{Keys: bson.M{"category_id": 1}},
		{Keys: bson.M{"is_active": 1}},
	}

	if _, err := productsCollection.Indexes().CreateMany(ctx, indexes); err != nil {
		return err
	}

	// Create categories collection
	categoriesCollection := database.Collection("categories")
	categoriesCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.M{"slug": 1},
		Options: options.Index().SetUnique(true),
	})

	// Create collections collection
	collectionsCollection := database.Collection("collections")
	collectionsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.M{"slug": 1},
		Options: options.Index().SetUnique(true),
	})

	return nil
}

func (r *PRStackReconciler) createCartCollections(ctx context.Context, database *mongo.Database) error {
	// Create carts collection
	cartsCollection := database.Collection("carts")
	indexes := []mongo.IndexModel{
		{Keys: bson.M{"user_id": 1}, Options: options.Index().SetUnique(true)},
		{Keys: bson.M{"session_id": 1}},
	}
	cartsCollection.Indexes().CreateMany(ctx, indexes)

	// Create cart_items collection
	cartItemsCollection := database.Collection("cart_items")
	cartItemsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{"cart_id": 1},
	})

	return nil
}

func (r *PRStackReconciler) createOrderCollections(ctx context.Context, database *mongo.Database) error {
	// Create orders collection
	ordersCollection := database.Collection("orders")
	indexes := []mongo.IndexModel{
		{Keys: bson.M{"order_number": 1}, Options: options.Index().SetUnique(true)},
		{Keys: bson.M{"user_id": 1}},
		{Keys: bson.M{"status": 1}},
		{Keys: bson.M{"created_at": 1}},
	}
	ordersCollection.Indexes().CreateMany(ctx, indexes)

	// Create order_items collection
	orderItemsCollection := database.Collection("order_items")
	orderItemsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{"order_id": 1},
	})

	// Create order_status_history collection
	orderStatusHistoryCollection := database.Collection("order_status_history")
	orderStatusHistoryCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{"order_id": 1},
	})

	return nil
}

func (r *PRStackReconciler) createPaymentCollections(ctx context.Context, database *mongo.Database) error {
	// Create payments collection
	paymentsCollection := database.Collection("payments")
	indexes := []mongo.IndexModel{
		{Keys: bson.M{"order_id": 1}},
		{Keys: bson.M{"user_id": 1}},
		{Keys: bson.M{"status": 1}},
	}
	paymentsCollection.Indexes().CreateMany(ctx, indexes)

	// Create payment_methods collection
	paymentMethodsCollection := database.Collection("payment_methods")
	paymentMethodsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{"user_id": 1},
	})

	// Create payment_transactions collection
	paymentTransactionsCollection := database.Collection("payment_transactions")
	paymentTransactionsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{"payment_id": 1},
	})

	return nil
}

func (r *PRStackReconciler) createCustomerCollections(ctx context.Context, database *mongo.Database) error {
	// Create customers collection
	customersCollection := database.Collection("customers")
	indexes := []mongo.IndexModel{
		{Keys: bson.M{"email": 1}, Options: options.Index().SetUnique(true)},
		{Keys: bson.M{"user_id": 1}, Options: options.Index().SetUnique(true)},
	}
	customersCollection.Indexes().CreateMany(ctx, indexes)

	// Create customer_addresses collection
	customerAddressesCollection := database.Collection("customer_addresses")
	customerAddressesCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{"customer_id": 1},
	})

	// Create customer_preferences collection
	customerPreferencesCollection := database.Collection("customer_preferences")
	customerPreferencesCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{"customer_id": 1},
	})

	return nil
}

func (r *PRStackReconciler) createInventoryCollections(ctx context.Context, database *mongo.Database) error {
	// Create inventory_items collection
	inventoryItemsCollection := database.Collection("inventory_items")
	indexes := []mongo.IndexModel{
		{Keys: bson.M{"product_id": 1}, Options: options.Index().SetUnique(true)},
		{Keys: bson.M{"sku": 1}},
	}
	inventoryItemsCollection.Indexes().CreateMany(ctx, indexes)

	// Create stock_movements collection
	stockMovementsCollection := database.Collection("stock_movements")
	stockMovementsCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.M{"product_id": 1}},
		{Keys: bson.M{"created_at": 1}},
	})

	// Create reservations collection
	reservationsCollection := database.Collection("reservations")
	reservationsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{"product_id": 1},
	})

	return nil
}

func (r *PRStackReconciler) createNotificationCollections(ctx context.Context, database *mongo.Database) error {
	// Create notifications collection
	notificationsCollection := database.Collection("notifications")
	indexes := []mongo.IndexModel{
		{Keys: bson.M{"user_id": 1}},
		{Keys: bson.M{"type": 1}},
		{Keys: bson.M{"status": 1}},
		{Keys: bson.M{"created_at": 1}},
	}
	notificationsCollection.Indexes().CreateMany(ctx, indexes)

	// Create notification_templates collection
	notificationTemplatesCollection := database.Collection("notification_templates")
	notificationTemplatesCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{"type": 1},
	})

	// Create notification_preferences collection
	notificationPreferencesCollection := database.Collection("notification_preferences")
	notificationPreferencesCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{"user_id": 1},
	})

	return nil
}

func (r *PRStackReconciler) createDiscountCollections(ctx context.Context, database *mongo.Database) error {
	// Create discounts collection
	discountsCollection := database.Collection("discounts")
	indexes := []mongo.IndexModel{
		{Keys: bson.M{"code": 1}, Options: options.Index().SetUnique(true)},
		{Keys: bson.M{"is_active": 1}},
		{Keys: bson.M{"valid_from": 1}},
		{Keys: bson.M{"valid_until": 1}},
	}
	discountsCollection.Indexes().CreateMany(ctx, indexes)

	// Create discount_usage collection
	discountUsageCollection := database.Collection("discount_usage")
	discountUsageCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{"discount_id": 1},
	})

	// Create promotion_codes collection
	promotionCodesCollection := database.Collection("promotion_codes")
	promotionCodesCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.M{"code": 1},
		Options: options.Index().SetUnique(true),
	})

	return nil
}

func (r *PRStackReconciler) createCheckoutCollections(ctx context.Context, database *mongo.Database) error {
	// Create checkout_sessions collection
	checkoutSessionsCollection := database.Collection("checkout_sessions")
	indexes := []mongo.IndexModel{
		{Keys: bson.M{"session_id": 1}, Options: options.Index().SetUnique(true)},
		{Keys: bson.M{"user_id": 1}},
		{Keys: bson.M{"status": 1}},
	}
	checkoutSessionsCollection.Indexes().CreateMany(ctx, indexes)

	// Create checkout_steps collection
	checkoutStepsCollection := database.Collection("checkout_steps")
	checkoutStepsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{"session_id": 1},
	})

	// Create shipping_options collection
	shippingOptionsCollection := database.Collection("shipping_options")
	shippingOptionsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{"is_active": 1},
	})

	return nil
}

func generateSecurePassword() (string, error) {
	// Generate a random password using crypto/rand for security.
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes for password: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (r *PRStackReconciler) createAnalyticsCollections(ctx context.Context, database *mongo.Database) error {
	// Create analytics collection with indexes
	analyticsCollection := database.Collection("analytics")

	// Create indexes
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "created_at", Value: 1}}},
		{Keys: bson.D{{Key: "event_type", Value: 1}}},
		{Keys: bson.D{{Key: "user_id", Value: 1}}},
	}

	if _, err := analyticsCollection.Indexes().CreateMany(ctx, indexes); err != nil {
		return fmt.Errorf("failed to create analytics indexes: %v", err)
	}

	return nil
}

func (r *PRStackReconciler) createAuthCollections(ctx context.Context, database *mongo.Database) error {
	// Create auth collection with indexes
	authCollection := database.Collection("users")

	// Create indexes
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "email", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "username", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "created_at", Value: 1}}},
	}

	if _, err := authCollection.Indexes().CreateMany(ctx, indexes); err != nil {
		return fmt.Errorf("failed to create auth indexes: %v", err)
	}

	return nil
}

func (r *PRStackReconciler) createGraphQLCollections(ctx context.Context, database *mongo.Database) error {
	// Create graphql collection with indexes
	graphqlCollection := database.Collection("queries")

	// Create indexes
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "created_at", Value: 1}}},
		{Keys: bson.D{{Key: "operation", Value: 1}}},
	}

	if _, err := graphqlCollection.Indexes().CreateMany(ctx, indexes); err != nil {
		return fmt.Errorf("failed to create graphql indexes: %v", err)
	}

	return nil
}

