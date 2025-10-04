package controllers

import (
	"strconv"

	corev1 "k8s.io/api/core/v1"
)

// ServiceConfig defines the configuration requirements for each service
type ServiceConfig struct {
	// Service identification
	ServiceName string
	Environment string
	LogLevel    string
	Version     string

	// Server configuration
	ServerPort   int
	ReadTimeout  string
	WriteTimeout string
	IdleTimeout  string

	// Database configuration
	MongoDBURI         string
	MongoDBDatabase    string
	MongoDBUsername    string
	MongoDBPassword    string
	MongoDBTimeout     string
	MongoDBMaxPoolSize string
	MongoDBMinPoolSize string

	// Cache configuration (Redis)
	RedisURL      string
	RedisPassword string
	RedisDatabase string
	RedisTimeout  string

	// NATS configuration
	NATSURL      string
	NATSUsername string
	NATSPassword string
	NATSToken    string
	NATSTimeout  string

	// Observability configuration
	MetricsEnabled     string
	MetricsPort        string
	MetricsPath        string
	TracingEnabled     string
	TracingServiceName string
	TracingEndpoint    string
	TracingSampleRate  string
	HealthEnabled      string
	HealthPort         string
	HealthPath         string

	// Event configuration
	EventsPublishInterval string
	EventsBatchSize       string
	EventsRetryAttempts   string
	EventsRetryDelay      string

	// Performance configuration
	ConnectionPoolMaxConnections    string
	ConnectionPoolMinConnections    string
	ConnectionPoolConnectionTimeout string
	ConnectionPoolIdleTimeout       string
	ConnectionPoolMaxLifetime       string
	RateLimitingEnabled             string
	RateLimitingRequestsPer         string
	RateLimitingWindowSize          string
	TimeoutRequestTimeout           string
	TimeoutResponseTimeout          string
	TimeoutGracefulShutdown         string

	// Service-specific configurations
	AnalyticsConfig    *AnalyticsServiceConfig
	SecurityConfig     *SecurityServiceConfig
	CartConfig         *CartServiceConfig
	OrderConfig        *OrderServiceConfig
	PaymentConfig      *PaymentServiceConfig
	CustomerConfig     *CustomerServiceConfig
	InventoryConfig    *InventoryServiceConfig
	NotificationConfig *NotificationServiceConfig
	DiscountConfig     *DiscountServiceConfig
	CheckoutConfig     *CheckoutServiceConfig
	AuthConfig         *AuthServiceConfig
	GraphQLConfig      *GraphQLServiceConfig
	InvoiceConfig      *InvoiceServiceConfig
	MonolithConfig     *MonolithServiceConfig
}

// AnalyticsServiceConfig defines analytics-specific configuration
type AnalyticsServiceConfig struct {
	DataRetentionDays       string
	BatchProcessingEnabled  string
	BatchSize               string
	BatchInterval           string
	RecommendationEnabled   string
	RecommendationCacheSize string
	RecommendationCacheTTL  string
	SimilarityThreshold     string
	RealTimeEnabled         string
	RealTimeWindow          string
	ReportGenerationEnabled string
	ReportSchedule          string
	ReportRetentionDays     string
}

// SecurityServiceConfig defines security-specific configuration
type SecurityServiceConfig struct {
	TLSEnabled  string
	TLSCertFile string
	TLSKeyFile  string
	CORSEnabled string
}

// CartServiceConfig defines cart-specific configuration
type CartServiceConfig struct {
	GuestExpiration         string
	UserExpiration          string
	SharedExpiration        string
	MaxItems                string
	MaxQuantity             string
	AutoMerge               string
	CrossDeviceSync         string
	PricingEnabled          string
	TaxEnabled              string
	DiscountEnabled         string
	Currency                string
	TaxRate                 string
	InventoryValidation     string
	PriceValidation         string
	QuantityValidation      string
	OptionValidation        string
	ExpirationCheckInterval string
	CleanupInterval         string
	AbandonmentTracking     string
	AbandonmentThreshold    string
}

// OrderServiceConfig defines order-specific configuration
type OrderServiceConfig struct {
	PaymentTimeout      string
	InventoryTimeout    string
	ShippingTimeout     string
	NotificationTimeout string
	OrderExpirationDays string
	MaxOrderItems       string
	AutoCancelTimeout   string
	FulfillmentTimeout  string
}

// PaymentServiceConfig defines payment-specific configuration
type PaymentServiceConfig struct {
	StripeEnabled            string
	StripeSecretKey          string
	StripePublishableKey     string
	StripeWebhookSecret      string
	PayPalEnabled            string
	PayPalClientID           string
	PayPalClientSecret       string
	PayPalEnvironment        string
	PaymentTimeout           string
	AuthorizationTimeout     string
	CaptureTimeout           string
	MaxRetries               string
	Currency                 string
	MinAmount                string
	MaxAmount                string
	NATSClusterID            string
	NATSClientID             string
	NATSReconnectWait        string
	NATSMaxReconnectAttempts string
}

// CustomerServiceConfig defines customer-specific configuration
type CustomerServiceConfig struct {
	// Registration
	RegistrationRequireEmailVerification string
	RegistrationDefaultMarketingOptIn    string
	RegistrationTimeout                  string
	RegistrationMaxAttempts              string

	// Profile
	ProfileAllowUpdates      string
	ProfileRequireValidation string
	ProfileUpdateTimeout     string
	ProfileMaxUpdateAttempts string

	// Address
	AddressMaxPerCustomer      string
	AddressRequireValidation   string
	AddressValidationTimeout   string
	AddressGeocodingEnabled    string
	AddressGeocodingServiceURL string

	// Preferences
	PreferencesAllowUpdates      string
	PreferencesUpdateTimeout     string
	PreferencesMaxUpdateAttempts string

	// Analytics
	AnalyticsEnableActivityTracking  string
	AnalyticsActivityRetentionPeriod string
	AnalyticsMetricsUpdateInterval   string
	AnalyticsSegmentUpdateInterval   string
	AnalyticsEnableMLSegmentation    string
	AnalyticsMLServiceURL            string

	// Privacy
	PrivacyEnableDataExport        string
	PrivacyEnableDataDeletion      string
	PrivacyDataExportTimeout       string
	PrivacyDataDeletionTimeout     string
	PrivacyRequireConsentTracking  string
	PrivacyConsentRetentionPeriod  string
	PrivacyEnableOptOutRequests    string
	PrivacyOptOutProcessingTimeout string
}

// InventoryServiceConfig defines inventory-specific configuration
type InventoryServiceConfig struct {
	// Low Stock Alert
	LowStockAlertEnabled          string
	LowStockAlertCheckInterval    string
	LowStockAlertDefaultThreshold string
	LowStockAlertBatchSize        string

	// Reservation
	ReservationDefaultCartTTL  string
	ReservationDefaultOrderTTL string
	ReservationMaxRetries      string

	// Cleanup
	CleanupEnabled   string
	CleanupInterval  string
	CleanupBatchSize string
}

// NotificationServiceConfig defines notification-specific configuration
type NotificationServiceConfig struct {
	EmailProvider           string
	SMTPHost                string
	SMTPPort                string
	SMTPUsername            string
	SMTPPassword            string
	SMSProvider             string
	TwilioAccountSID        string
	TwilioAuthToken         string
	PushNotificationEnabled string
	EmailEnabled            string
	SMSEnabled              string
	TemplatePath            string
}

// DiscountServiceConfig defines discount-specific configuration
type DiscountServiceConfig struct {
	MaxDiscountPercentage   string
	MinOrderAmount          string
	MaxUsagePerCustomer     string
	MaxUsageTotal           string
	ExpirationCheckInterval string
	AutoExpireEnabled       string
	StackableDiscounts      string
	ExclusiveDiscounts      string
}

// CheckoutServiceConfig defines checkout-specific configuration
type CheckoutServiceConfig struct {
	CheckoutTimeout            string
	PaymentTimeout             string
	InventoryTimeout           string
	ShippingTimeout            string
	TaxCalculationEnabled      string
	ShippingCalculationEnabled string
	GuestCheckoutEnabled       string
	AutoApplyDiscounts         string
	RequireShippingAddress     string
}

// AuthServiceConfig defines auth-specific configuration
type AuthServiceConfig struct {
	JWTSecret                string
	JWTExpiration            string
	RefreshTokenExpiration   string
	PasswordMinLength        string
	RequireEmailVerification string
	MaxLoginAttempts         string
	LockoutDuration          string
	OAuthProviders           string
	SessionTimeout           string
	TokenRefreshEnabled      string
}

// GraphQLServiceConfig defines GraphQL-specific configuration
type GraphQLServiceConfig struct {
	PlaygroundEnabled    string
	IntrospectionEnabled string
	ComplexityLimit      string
	QueryTimeout         string
	MaxQueryDepth        string
	MaxQueryComplexity   string
}

// InvoiceServiceConfig defines invoice-specific configuration
type InvoiceServiceConfig struct {
	DefaultCurrency     string
	InvoicePrefix       string
	DueDateDays         string
	LateFeePercentage   string
	TaxRate             string
	AutoGenerateEnabled string
	PDFTemplatePath     string
	EmailEnabled        string
}

// MonolithServiceConfig defines monolith-specific configuration
type MonolithServiceConfig struct {
	EmbeddedNATSPort      string
	EmbeddedNATSJetStream string
	EmbeddedNATSDataDir   string
	EmbeddedNATSLogLevel  string
	EmbeddedNATSDebug     string
	EmbeddedNATSTrace     string
}

// GetServiceConfig returns the standardized configuration for a service
func GetServiceConfig(serviceName, prNumber string) *ServiceConfig {
	config := &ServiceConfig{
		// Service identification
		ServiceName: serviceName,
		Environment: "pr",
		LogLevel:    "info",
		Version:     "1.0.0",

		// Server configuration
		ServerPort:   8080,
		ReadTimeout:  "30s",
		WriteTimeout: "30s",
		IdleTimeout:  "60s",

		// Database configuration
		MongoDBURI:         "mongodb://mongodb-service:27017",
		MongoDBDatabase:    getDatabaseName(serviceName, prNumber),
		MongoDBUsername:    "admin",
		MongoDBPassword:    "", // Will be set from secret
		MongoDBTimeout:     "10s",
		MongoDBMaxPoolSize: "100",
		MongoDBMinPoolSize: "10",

		// Cache configuration (Redis)
		RedisURL:      "redis://redis-service:6379",
		RedisPassword: "", // Will be set from secret if needed
		RedisDatabase: "0",
		RedisTimeout:  "10s",

		// NATS configuration
		NATSURL:      "nats://nats-service:4222",
		NATSUsername: "",
		NATSPassword: "",
		NATSToken:    "",
		NATSTimeout:  "10s",

		// Observability configuration
		MetricsEnabled:     "true",
		MetricsPort:        "8080",
		MetricsPath:        "/metrics",
		TracingEnabled:     "true",
		TracingServiceName: serviceName,
		TracingEndpoint:    "http://jaeger-service:14268/api/traces",
		TracingSampleRate:  "0.1",
		HealthEnabled:      "true",
		HealthPort:         "8080",
		HealthPath:         "/health",

		// Event configuration
		EventsPublishInterval: "30s",
		EventsBatchSize:       "100",
		EventsRetryAttempts:   "3",
		EventsRetryDelay:      "5s",

		// Performance configuration
		ConnectionPoolMaxConnections:    "100",
		ConnectionPoolMinConnections:    "10",
		ConnectionPoolConnectionTimeout: "30s",
		ConnectionPoolIdleTimeout:       "5m",
		ConnectionPoolMaxLifetime:       "1h",
		RateLimitingEnabled:             "true",
		RateLimitingRequestsPer:         "1000",
		RateLimitingWindowSize:          "60",
		TimeoutRequestTimeout:           "30s",
		TimeoutResponseTimeout:          "30s",
		TimeoutGracefulShutdown:         "5s",
	}

	// Add service-specific configurations
	switch serviceName {
	case "analytics-service":
		config.AnalyticsConfig = &AnalyticsServiceConfig{
			DataRetentionDays:       "365",
			BatchProcessingEnabled:  "true",
			BatchSize:               "1000",
			BatchInterval:           "5m",
			RecommendationEnabled:   "true",
			RecommendationCacheSize: "10000",
			RecommendationCacheTTL:  "1h",
			SimilarityThreshold:     "0.7",
			RealTimeEnabled:         "true",
			RealTimeWindow:          "1m",
			ReportGenerationEnabled: "true",
			ReportSchedule:          "0 0 * * *",
			ReportRetentionDays:     "90",
		}
	case "product-service":
		config.SecurityConfig = &SecurityServiceConfig{
			TLSEnabled:  "false",
			TLSCertFile: "",
			TLSKeyFile:  "",
			CORSEnabled: "true",
		}
	case "cart-service":
		config.CartConfig = &CartServiceConfig{
			GuestExpiration:         "24h",
			UserExpiration:          "30d",
			SharedExpiration:        "7d",
			MaxItems:                "100",
			MaxQuantity:             "99",
			AutoMerge:               "true",
			CrossDeviceSync:         "true",
			PricingEnabled:          "true",
			TaxEnabled:              "true",
			DiscountEnabled:         "true",
			Currency:                "USD",
			TaxRate:                 "0.08",
			InventoryValidation:     "true",
			PriceValidation:         "true",
			QuantityValidation:      "true",
			OptionValidation:        "true",
			ExpirationCheckInterval: "1h",
			CleanupInterval:         "6h",
			AbandonmentTracking:     "true",
			AbandonmentThreshold:    "2h",
		}
	case "order-service":
		config.OrderConfig = &OrderServiceConfig{
			PaymentTimeout:      "30s",
			InventoryTimeout:    "10s",
			ShippingTimeout:     "15s",
			NotificationTimeout: "5s",
			OrderExpirationDays: "30",
			MaxOrderItems:       "50",
			AutoCancelTimeout:   "24h",
			FulfillmentTimeout:  "7d",
		}
	case "payment-service":
		config.PaymentConfig = &PaymentServiceConfig{
			StripeEnabled:            "false",
			StripeSecretKey:          "", // Will be set from secret
			StripePublishableKey:     "", // Will be set from secret
			StripeWebhookSecret:      "", // Will be set from secret
			PayPalEnabled:            "false",
			PayPalClientID:           "", // Will be set from secret
			PayPalClientSecret:       "", // Will be set from secret
			PayPalEnvironment:        "sandbox",
			PaymentTimeout:           "300s",
			AuthorizationTimeout:     "60s",
			CaptureTimeout:           "168h",
			MaxRetries:               "3",
			Currency:                 "USD",
			MinAmount:                "1",
			MaxAmount:                "99999999",
			NATSClusterID:            "test-cluster",
			NATSClientID:             "payment-service",
			NATSReconnectWait:        "2s",
			NATSMaxReconnectAttempts: "10",
		}
	case "customer-service":
		config.CustomerConfig = &CustomerServiceConfig{
			// Registration
			RegistrationRequireEmailVerification: "true",
			RegistrationDefaultMarketingOptIn:    "false",
			RegistrationTimeout:                  "30s",
			RegistrationMaxAttempts:              "3",

			// Profile
			ProfileAllowUpdates:      "true",
			ProfileRequireValidation: "true",
			ProfileUpdateTimeout:     "30s",
			ProfileMaxUpdateAttempts: "3",

			// Address
			AddressMaxPerCustomer:      "10",
			AddressRequireValidation:   "true",
			AddressValidationTimeout:   "30s",
			AddressGeocodingEnabled:    "true",
			AddressGeocodingServiceURL: "http://localhost:8082",

			// Preferences
			PreferencesAllowUpdates:      "true",
			PreferencesUpdateTimeout:     "30s",
			PreferencesMaxUpdateAttempts: "3",

			// Analytics
			AnalyticsEnableActivityTracking:  "true",
			AnalyticsActivityRetentionPeriod: "2y",
			AnalyticsMetricsUpdateInterval:   "1h",
			AnalyticsSegmentUpdateInterval:   "24h",
			AnalyticsEnableMLSegmentation:    "true",
			AnalyticsMLServiceURL:            "http://localhost:8083",

			// Privacy
			PrivacyEnableDataExport:        "true",
			PrivacyEnableDataDeletion:      "true",
			PrivacyDataExportTimeout:       "5m",
			PrivacyDataDeletionTimeout:     "5m",
			PrivacyRequireConsentTracking:  "true",
			PrivacyConsentRetentionPeriod:  "7y",
			PrivacyEnableOptOutRequests:    "true",
			PrivacyOptOutProcessingTimeout: "5m",
		}
	case "inventory-service":
		config.InventoryConfig = &InventoryServiceConfig{
			// Low Stock Alert
			LowStockAlertEnabled:          "true",
			LowStockAlertCheckInterval:    "1h",
			LowStockAlertDefaultThreshold: "10",
			LowStockAlertBatchSize:        "100",

			// Reservation
			ReservationDefaultCartTTL:  "15m",
			ReservationDefaultOrderTTL: "30m",
			ReservationMaxRetries:      "3",

			// Cleanup
			CleanupEnabled:   "true",
			CleanupInterval:  "1h",
			CleanupBatchSize: "1000",
		}
	case "notification-service":
		config.NotificationConfig = &NotificationServiceConfig{
			EmailProvider:           "smtp",
			SMTPHost:                "smtp.gmail.com",
			SMTPPort:                "587",
			SMTPUsername:            "", // Will be set from secret
			SMTPPassword:            "", // Will be set from secret
			SMSProvider:             "twilio",
			TwilioAccountSID:        "", // Will be set from secret
			TwilioAuthToken:         "", // Will be set from secret
			PushNotificationEnabled: "true",
			EmailEnabled:            "true",
			SMSEnabled:              "true",
			TemplatePath:            "/templates",
		}
	case "discount-service":
		config.DiscountConfig = &DiscountServiceConfig{
			MaxDiscountPercentage:   "50",
			MinOrderAmount:          "10.00",
			MaxUsagePerCustomer:     "1",
			MaxUsageTotal:           "1000",
			ExpirationCheckInterval: "1h",
			AutoExpireEnabled:       "true",
			StackableDiscounts:      "false",
			ExclusiveDiscounts:      "true",
		}
	case "checkout-service":
		config.CheckoutConfig = &CheckoutServiceConfig{
			CheckoutTimeout:            "10m",
			PaymentTimeout:             "30s",
			InventoryTimeout:           "10s",
			ShippingTimeout:            "15s",
			TaxCalculationEnabled:      "true",
			ShippingCalculationEnabled: "true",
			GuestCheckoutEnabled:       "true",
			AutoApplyDiscounts:         "true",
			RequireShippingAddress:     "true",
		}
	case "auth-service":
		config.AuthConfig = &AuthServiceConfig{
			JWTSecret:                "", // Will be set from secret
			JWTExpiration:            "1h",
			RefreshTokenExpiration:   "7d",
			PasswordMinLength:        "8",
			RequireEmailVerification: "true",
			MaxLoginAttempts:         "5",
			LockoutDuration:          "30m",
			OAuthProviders:           "google,github",
			SessionTimeout:           "24h",
			TokenRefreshEnabled:      "true",
		}
	case "graphql-service":
		config.GraphQLConfig = &GraphQLServiceConfig{
			PlaygroundEnabled:    "true",
			IntrospectionEnabled: "true",
			ComplexityLimit:      "1000",
			QueryTimeout:         "30s",
			MaxQueryDepth:        "10",
			MaxQueryComplexity:   "1000",
		}
	case "invoice-service":
		config.InvoiceConfig = &InvoiceServiceConfig{
			DefaultCurrency:     "USD",
			InvoicePrefix:       "INV",
			DueDateDays:         "30",
			LateFeePercentage:   "2.5",
			TaxRate:             "8.0",
			AutoGenerateEnabled: "true",
			PDFTemplatePath:     "/templates/invoice.pdf",
			EmailEnabled:        "true",
		}
	case "monolith-service":
		config.MonolithConfig = &MonolithServiceConfig{
			EmbeddedNATSPort:      "4222",
			EmbeddedNATSJetStream: "true",
			EmbeddedNATSDataDir:   "/data/nats",
			EmbeddedNATSLogLevel:  "info",
			EmbeddedNATSDebug:     "false",
			EmbeddedNATSTrace:     "false",
		}
	}

	return config
}

// ToEnvVars converts the service configuration to Kubernetes environment variables
func (sc *ServiceConfig) ToEnvVars() []corev1.EnvVar {
	envVars := []corev1.EnvVar{
		// Service identification
		{Name: "SERVICE_NAME", Value: sc.ServiceName},
		{Name: "ENVIRONMENT", Value: sc.Environment},
		{Name: "LOG_LEVEL", Value: sc.LogLevel},
		{Name: "VERSION", Value: sc.Version},

		// Server configuration - Fixed naming to match service expectations
		{Name: "SERVICE_PORT", Value: strconv.Itoa(sc.ServerPort)},
		{Name: "SERVER_READ_TIMEOUT", Value: sc.ReadTimeout},
		{Name: "SERVER_WRITE_TIMEOUT", Value: sc.WriteTimeout},
		{Name: "SERVER_IDLE_TIMEOUT", Value: sc.IdleTimeout},

		// Database configuration - Fixed naming to match service expectations
		{Name: "MONGODB_URI", Value: sc.MongoDBURI},
		{Name: "MONGODB_DATABASE", Value: sc.MongoDBDatabase},
		{Name: "MONGODB_TIMEOUT", Value: sc.MongoDBTimeout},
		{Name: "MONGODB_MAX_POOL_SIZE", Value: sc.MongoDBMaxPoolSize},
		{Name: "MONGODB_MIN_POOL_SIZE", Value: sc.MongoDBMinPoolSize},

		// Cache configuration (Redis) - Fixed naming to match service expectations
		{Name: "REDIS_URL", Value: sc.RedisURL},
		{Name: "REDIS_DATABASE", Value: sc.RedisDatabase},
		{Name: "REDIS_TIMEOUT", Value: sc.RedisTimeout},

		// NATS configuration
		{Name: "NATS_URL", Value: sc.NATSURL},
		{Name: "NATS_TIMEOUT", Value: sc.NATSTimeout},

		// Observability configuration - Fixed naming to match service expectations
		{Name: "METRICS_ENABLED", Value: sc.MetricsEnabled},
		{Name: "METRICS_PORT", Value: sc.MetricsPort},
		{Name: "METRICS_PATH", Value: sc.MetricsPath},
		{Name: "TRACING_ENABLED", Value: sc.TracingEnabled},
		{Name: "TRACING_SERVICE_NAME", Value: sc.TracingServiceName},
		{Name: "TRACING_ENDPOINT", Value: sc.TracingEndpoint},
		{Name: "TRACING_SAMPLE_RATE", Value: sc.TracingSampleRate},
		{Name: "HEALTH_ENABLED", Value: sc.HealthEnabled},
		{Name: "HEALTH_PORT", Value: sc.HealthPort},
		{Name: "HEALTH_PATH", Value: sc.HealthPath},

		// Event configuration
		{Name: "EVENTS_PUBLISH_INTERVAL", Value: sc.EventsPublishInterval},
		{Name: "EVENTS_BATCH_SIZE", Value: sc.EventsBatchSize},
		{Name: "EVENTS_RETRY_ATTEMPTS", Value: sc.EventsRetryAttempts},
		{Name: "EVENTS_RETRY_DELAY", Value: sc.EventsRetryDelay},

		// Performance configuration
		{Name: "CONNECTION_POOL_MAX_CONNECTIONS", Value: sc.ConnectionPoolMaxConnections},
		{Name: "CONNECTION_POOL_MIN_CONNECTIONS", Value: sc.ConnectionPoolMinConnections},
		{Name: "CONNECTION_POOL_CONNECTION_TIMEOUT", Value: sc.ConnectionPoolConnectionTimeout},
		{Name: "CONNECTION_POOL_IDLE_TIMEOUT", Value: sc.ConnectionPoolIdleTimeout},
		{Name: "CONNECTION_POOL_MAX_LIFETIME", Value: sc.ConnectionPoolMaxLifetime},
		{Name: "RATE_LIMITING_ENABLED", Value: sc.RateLimitingEnabled},
		{Name: "RATE_LIMITING_REQUESTS_PER", Value: sc.RateLimitingRequestsPer},
		{Name: "RATE_LIMITING_WINDOW_SIZE", Value: sc.RateLimitingWindowSize},
		{Name: "REQUEST_TIMEOUT", Value: sc.TimeoutRequestTimeout},
		{Name: "RESPONSE_TIMEOUT", Value: sc.TimeoutResponseTimeout},
		{Name: "GRACEFUL_SHUTDOWN", Value: sc.TimeoutGracefulShutdown},
	}

	// Add MongoDB credentials from secret - Fixed naming to match service expectations
	envVars = append(envVars, corev1.EnvVar{
		Name: "MONGODB_USERNAME",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: "mongodb-secret"},
				Key:                  "username",
			},
		},
	})

	envVars = append(envVars, corev1.EnvVar{
		Name: "MONGODB_PASSWORD",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: "mongodb-secret"},
				Key:                  "password",
			},
		},
	})

	// Add Redis password from secret if needed - Fixed naming to match service expectations
	if sc.RedisPassword != "" {
		envVars = append(envVars, corev1.EnvVar{
			Name: "REDIS_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: "redis-secret"},
					Key:                  "password",
				},
			},
		})
	}

	// Add NATS credentials from secret if needed
	if sc.NATSUsername != "" {
		envVars = append(envVars, corev1.EnvVar{
			Name: "NATS_USERNAME",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: "nats-secret"},
					Key:                  "username",
				},
			},
		})
	}

	if sc.NATSPassword != "" {
		envVars = append(envVars, corev1.EnvVar{
			Name: "NATS_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: "nats-secret"},
					Key:                  "password",
				},
			},
		})
	}

	// Add service-specific environment variables
	if sc.AnalyticsConfig != nil {
		analyticsEnvVars := []corev1.EnvVar{
			{Name: "ANALYTICS_DATA_RETENTION_DAYS", Value: sc.AnalyticsConfig.DataRetentionDays},
			{Name: "ANALYTICS_BATCH_PROCESSING_ENABLED", Value: sc.AnalyticsConfig.BatchProcessingEnabled},
			{Name: "ANALYTICS_BATCH_SIZE", Value: sc.AnalyticsConfig.BatchSize},
			{Name: "ANALYTICS_BATCH_INTERVAL", Value: sc.AnalyticsConfig.BatchInterval},
			{Name: "ANALYTICS_RECOMMENDATION_ENABLED", Value: sc.AnalyticsConfig.RecommendationEnabled},
			{Name: "ANALYTICS_RECOMMENDATION_CACHE_SIZE", Value: sc.AnalyticsConfig.RecommendationCacheSize},
			{Name: "ANALYTICS_RECOMMENDATION_CACHE_TTL", Value: sc.AnalyticsConfig.RecommendationCacheTTL},
			{Name: "ANALYTICS_SIMILARITY_THRESHOLD", Value: sc.AnalyticsConfig.SimilarityThreshold}, // String value will be converted to float64 in service
			{Name: "ANALYTICS_REAL_TIME_ENABLED", Value: sc.AnalyticsConfig.RealTimeEnabled},
			{Name: "ANALYTICS_REAL_TIME_WINDOW", Value: sc.AnalyticsConfig.RealTimeWindow},
			{Name: "ANALYTICS_REPORT_GENERATION_ENABLED", Value: sc.AnalyticsConfig.ReportGenerationEnabled},
			{Name: "ANALYTICS_REPORT_SCHEDULE", Value: sc.AnalyticsConfig.ReportSchedule},
			{Name: "ANALYTICS_REPORT_RETENTION_DAYS", Value: sc.AnalyticsConfig.ReportRetentionDays},
		}
		envVars = append(envVars, analyticsEnvVars...)
	}

	if sc.SecurityConfig != nil {
		securityEnvVars := []corev1.EnvVar{
			{Name: "SECURITY_TLS_ENABLED", Value: sc.SecurityConfig.TLSEnabled},
			{Name: "SECURITY_TLS_CERT_FILE", Value: sc.SecurityConfig.TLSCertFile},
			{Name: "SECURITY_TLS_KEY_FILE", Value: sc.SecurityConfig.TLSKeyFile},
			{Name: "SECURITY_CORS_ENABLED", Value: sc.SecurityConfig.CORSEnabled},
		}
		envVars = append(envVars, securityEnvVars...)
	}

	if sc.CartConfig != nil {
		cartEnvVars := []corev1.EnvVar{
			{Name: "CART_GUEST_EXPIRATION", Value: sc.CartConfig.GuestExpiration},
			{Name: "CART_USER_EXPIRATION", Value: sc.CartConfig.UserExpiration},
			{Name: "CART_SHARED_EXPIRATION", Value: sc.CartConfig.SharedExpiration},
			{Name: "CART_MAX_ITEMS", Value: sc.CartConfig.MaxItems},
			{Name: "CART_MAX_QUANTITY", Value: sc.CartConfig.MaxQuantity},
			{Name: "CART_AUTO_MERGE", Value: sc.CartConfig.AutoMerge},
			{Name: "CART_CROSS_DEVICE_SYNC", Value: sc.CartConfig.CrossDeviceSync},
			{Name: "CART_PRICING_ENABLED", Value: sc.CartConfig.PricingEnabled},
			{Name: "CART_TAX_ENABLED", Value: sc.CartConfig.TaxEnabled},
			{Name: "CART_DISCOUNT_ENABLED", Value: sc.CartConfig.DiscountEnabled},
			{Name: "CART_CURRENCY", Value: sc.CartConfig.Currency},
			{Name: "CART_TAX_RATE", Value: sc.CartConfig.TaxRate},
			{Name: "CART_INVENTORY_VALIDATION", Value: sc.CartConfig.InventoryValidation},
			{Name: "CART_PRICE_VALIDATION", Value: sc.CartConfig.PriceValidation},
			{Name: "CART_QUANTITY_VALIDATION", Value: sc.CartConfig.QuantityValidation},
			{Name: "CART_OPTION_VALIDATION", Value: sc.CartConfig.OptionValidation},
			{Name: "CART_EXPIRATION_CHECK_INTERVAL", Value: sc.CartConfig.ExpirationCheckInterval},
			{Name: "CART_CLEANUP_INTERVAL", Value: sc.CartConfig.CleanupInterval},
			{Name: "CART_ABANDONMENT_TRACKING", Value: sc.CartConfig.AbandonmentTracking},
			{Name: "CART_ABANDONMENT_THRESHOLD", Value: sc.CartConfig.AbandonmentThreshold},
		}
		envVars = append(envVars, cartEnvVars...)
	}

	if sc.OrderConfig != nil {
		orderEnvVars := []corev1.EnvVar{
			// Order service URLs and timeouts
			{Name: "PAYMENT_TIMEOUT", Value: sc.OrderConfig.PaymentTimeout},
			{Name: "INVENTORY_TIMEOUT", Value: sc.OrderConfig.InventoryTimeout},
			{Name: "SHIPPING_TIMEOUT", Value: sc.OrderConfig.ShippingTimeout},
			{Name: "NOTIFICATION_TIMEOUT", Value: sc.OrderConfig.NotificationTimeout},
			// Order-specific settings
			{Name: "ORDER_EXPIRATION_DAYS", Value: sc.OrderConfig.OrderExpirationDays},
			{Name: "ORDER_MAX_ITEMS", Value: sc.OrderConfig.MaxOrderItems},
			{Name: "ORDER_AUTO_CANCEL_TIMEOUT", Value: sc.OrderConfig.AutoCancelTimeout},
			{Name: "ORDER_FULFILLMENT_TIMEOUT", Value: sc.OrderConfig.FulfillmentTimeout},
		}
		envVars = append(envVars, orderEnvVars...)
	}

	if sc.PaymentConfig != nil {
		paymentEnvVars := []corev1.EnvVar{
			{Name: "STRIPE_ENABLED", Value: sc.PaymentConfig.StripeEnabled},
			{Name: "PAYPAL_ENABLED", Value: sc.PaymentConfig.PayPalEnabled},
			{Name: "PAYPAL_ENVIRONMENT", Value: sc.PaymentConfig.PayPalEnvironment},
			{Name: "PAYMENT_TIMEOUT", Value: sc.PaymentConfig.PaymentTimeout},
			{Name: "PAYMENT_AUTHORIZATION_TIMEOUT", Value: sc.PaymentConfig.AuthorizationTimeout},
			{Name: "PAYMENT_CAPTURE_TIMEOUT", Value: sc.PaymentConfig.CaptureTimeout},
			{Name: "PAYMENT_MAX_RETRIES", Value: sc.PaymentConfig.MaxRetries},
			{Name: "PAYMENT_CURRENCY", Value: sc.PaymentConfig.Currency},
			{Name: "PAYMENT_MIN_AMOUNT", Value: sc.PaymentConfig.MinAmount},
			{Name: "PAYMENT_MAX_AMOUNT", Value: sc.PaymentConfig.MaxAmount},
			{Name: "NATS_CLUSTER_ID", Value: sc.PaymentConfig.NATSClusterID},
			{Name: "NATS_CLIENT_ID", Value: sc.PaymentConfig.NATSClientID},
			{Name: "NATS_RECONNECT_WAIT", Value: sc.PaymentConfig.NATSReconnectWait},
			{Name: "NATS_MAX_RECONNECT_ATTEMPTS", Value: sc.PaymentConfig.NATSMaxReconnectAttempts},
		}
		envVars = append(envVars, paymentEnvVars...)

		// Add Stripe credentials from secret
		envVars = append(envVars, corev1.EnvVar{
			Name: "STRIPE_SECRET_KEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: "payment-secret"},
					Key:                  "stripe-secret-key",
				},
			},
		})

		envVars = append(envVars, corev1.EnvVar{
			Name: "STRIPE_PUBLISHABLE_KEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: "payment-secret"},
					Key:                  "stripe-publishable-key",
				},
			},
		})

		envVars = append(envVars, corev1.EnvVar{
			Name: "STRIPE_WEBHOOK_SECRET",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: "payment-secret"},
					Key:                  "stripe-webhook-secret",
				},
			},
		})

		// Add PayPal credentials from secret
		envVars = append(envVars, corev1.EnvVar{
			Name: "PAYPAL_CLIENT_ID",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: "payment-secret"},
					Key:                  "paypal-client-id",
				},
			},
		})

		envVars = append(envVars, corev1.EnvVar{
			Name: "PAYPAL_CLIENT_SECRET",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: "payment-secret"},
					Key:                  "paypal-client-secret",
				},
			},
		})
	}

	if sc.CustomerConfig != nil {
		customerEnvVars := []corev1.EnvVar{
			// Registration settings
			{Name: "REGISTRATION_REQUIRE_EMAIL_VERIFICATION", Value: sc.CustomerConfig.RegistrationRequireEmailVerification},
			{Name: "REGISTRATION_DEFAULT_MARKETING_OPT_IN", Value: sc.CustomerConfig.RegistrationDefaultMarketingOptIn},
			{Name: "REGISTRATION_TIMEOUT", Value: sc.CustomerConfig.RegistrationTimeout},
			{Name: "REGISTRATION_MAX_ATTEMPTS", Value: sc.CustomerConfig.RegistrationMaxAttempts},

			// Profile settings
			{Name: "PROFILE_ALLOW_UPDATES", Value: sc.CustomerConfig.ProfileAllowUpdates},
			{Name: "PROFILE_REQUIRE_VALIDATION", Value: sc.CustomerConfig.ProfileRequireValidation},
			{Name: "PROFILE_UPDATE_TIMEOUT", Value: sc.CustomerConfig.ProfileUpdateTimeout},
			{Name: "PROFILE_MAX_UPDATE_ATTEMPTS", Value: sc.CustomerConfig.ProfileMaxUpdateAttempts},

			// Address settings
			{Name: "ADDRESS_MAX_PER_CUSTOMER", Value: sc.CustomerConfig.AddressMaxPerCustomer},
			{Name: "ADDRESS_REQUIRE_VALIDATION", Value: sc.CustomerConfig.AddressRequireValidation},
			{Name: "ADDRESS_VALIDATION_TIMEOUT", Value: sc.CustomerConfig.AddressValidationTimeout},
			{Name: "ADDRESS_GEOCODING_ENABLED", Value: sc.CustomerConfig.AddressGeocodingEnabled},
			{Name: "ADDRESS_GEOCODING_SERVICE_URL", Value: sc.CustomerConfig.AddressGeocodingServiceURL},

			// Preferences settings
			{Name: "PREFERENCES_ALLOW_UPDATES", Value: sc.CustomerConfig.PreferencesAllowUpdates},
			{Name: "PREFERENCES_UPDATE_TIMEOUT", Value: sc.CustomerConfig.PreferencesUpdateTimeout},
			{Name: "PREFERENCES_MAX_UPDATE_ATTEMPTS", Value: sc.CustomerConfig.PreferencesMaxUpdateAttempts},

			// Analytics settings
			{Name: "ANALYTICS_ENABLE_ACTIVITY_TRACKING", Value: sc.CustomerConfig.AnalyticsEnableActivityTracking},
			{Name: "ANALYTICS_ACTIVITY_RETENTION_PERIOD", Value: sc.CustomerConfig.AnalyticsActivityRetentionPeriod},
			{Name: "ANALYTICS_METRICS_UPDATE_INTERVAL", Value: sc.CustomerConfig.AnalyticsMetricsUpdateInterval},
			{Name: "ANALYTICS_SEGMENT_UPDATE_INTERVAL", Value: sc.CustomerConfig.AnalyticsSegmentUpdateInterval},
			{Name: "ANALYTICS_ENABLE_ML_SEGMENTATION", Value: sc.CustomerConfig.AnalyticsEnableMLSegmentation},
			{Name: "ANALYTICS_ML_SERVICE_URL", Value: sc.CustomerConfig.AnalyticsMLServiceURL},

			// Privacy settings
			{Name: "PRIVACY_ENABLE_DATA_EXPORT", Value: sc.CustomerConfig.PrivacyEnableDataExport},
			{Name: "PRIVACY_ENABLE_DATA_DELETION", Value: sc.CustomerConfig.PrivacyEnableDataDeletion},
			{Name: "PRIVACY_DATA_EXPORT_TIMEOUT", Value: sc.CustomerConfig.PrivacyDataExportTimeout},
			{Name: "PRIVACY_DATA_DELETION_TIMEOUT", Value: sc.CustomerConfig.PrivacyDataDeletionTimeout},
			{Name: "PRIVACY_REQUIRE_CONSENT_TRACKING", Value: sc.CustomerConfig.PrivacyRequireConsentTracking},
			{Name: "PRIVACY_CONSENT_RETENTION_PERIOD", Value: sc.CustomerConfig.PrivacyConsentRetentionPeriod},
			{Name: "PRIVACY_ENABLE_OPT_OUT_REQUESTS", Value: sc.CustomerConfig.PrivacyEnableOptOutRequests},
			{Name: "PRIVACY_OPT_OUT_PROCESSING_TIMEOUT", Value: sc.CustomerConfig.PrivacyOptOutProcessingTimeout},
		}
		envVars = append(envVars, customerEnvVars...)
	}

	if sc.InventoryConfig != nil {
		inventoryEnvVars := []corev1.EnvVar{
			// Low Stock Alert settings
			{Name: "LOW_STOCK_ALERT_ENABLED", Value: sc.InventoryConfig.LowStockAlertEnabled},
			{Name: "LOW_STOCK_ALERT_CHECK_INTERVAL", Value: sc.InventoryConfig.LowStockAlertCheckInterval},
			{Name: "LOW_STOCK_ALERT_DEFAULT_THRESHOLD", Value: sc.InventoryConfig.LowStockAlertDefaultThreshold},
			{Name: "LOW_STOCK_ALERT_BATCH_SIZE", Value: sc.InventoryConfig.LowStockAlertBatchSize},

			// Reservation settings
			{Name: "RESERVATION_DEFAULT_CART_TTL", Value: sc.InventoryConfig.ReservationDefaultCartTTL},
			{Name: "RESERVATION_DEFAULT_ORDER_TTL", Value: sc.InventoryConfig.ReservationDefaultOrderTTL},
			{Name: "RESERVATION_MAX_RETRIES", Value: sc.InventoryConfig.ReservationMaxRetries},

			// Cleanup settings
			{Name: "CLEANUP_ENABLED", Value: sc.InventoryConfig.CleanupEnabled},
			{Name: "CLEANUP_INTERVAL", Value: sc.InventoryConfig.CleanupInterval},
			{Name: "CLEANUP_BATCH_SIZE", Value: sc.InventoryConfig.CleanupBatchSize},
		}
		envVars = append(envVars, inventoryEnvVars...)
	}

	if sc.NotificationConfig != nil {
		notificationEnvVars := []corev1.EnvVar{
			// Email provider settings
			{Name: "EMAIL_PROVIDER", Value: sc.NotificationConfig.EmailProvider},
			{Name: "SMTP_HOST", Value: sc.NotificationConfig.SMTPHost},
			{Name: "SMTP_PORT", Value: sc.NotificationConfig.SMTPPort},
			// SMS provider settings
			{Name: "SMS_PROVIDER", Value: sc.NotificationConfig.SMSProvider},
			// Notification type settings
			{Name: "PUSH_NOTIFICATION_ENABLED", Value: sc.NotificationConfig.PushNotificationEnabled},
			{Name: "EMAIL_ENABLED", Value: sc.NotificationConfig.EmailEnabled},
			{Name: "SMS_ENABLED", Value: sc.NotificationConfig.SMSEnabled},
			// Template settings
			{Name: "TEMPLATE_PATH", Value: sc.NotificationConfig.TemplatePath},
		}
		envVars = append(envVars, notificationEnvVars...)

		// Add SMTP credentials from secret - Fixed naming to match service expectations
		envVars = append(envVars, corev1.EnvVar{
			Name: "SMTP_USERNAME",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: "notification-secret"},
					Key:                  "smtp-username",
				},
			},
		})

		envVars = append(envVars, corev1.EnvVar{
			Name: "SMTP_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: "notification-secret"},
					Key:                  "smtp-password",
				},
			},
		})

		// Add Twilio credentials from secret - Fixed naming to match service expectations
		envVars = append(envVars, corev1.EnvVar{
			Name: "TWILIO_ACCOUNT_SID",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: "notification-secret"},
					Key:                  "twilio-account-sid",
				},
			},
		})

		envVars = append(envVars, corev1.EnvVar{
			Name: "TWILIO_AUTH_TOKEN",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: "notification-secret"},
					Key:                  "twilio-auth-token",
				},
			},
		})
	}

	if sc.DiscountConfig != nil {
		discountEnvVars := []corev1.EnvVar{
			// Discount limits and rules
			{Name: "MAX_DISCOUNT_PERCENTAGE", Value: sc.DiscountConfig.MaxDiscountPercentage},
			{Name: "MIN_ORDER_AMOUNT", Value: sc.DiscountConfig.MinOrderAmount},
			{Name: "MAX_USAGE_PER_CUSTOMER", Value: sc.DiscountConfig.MaxUsagePerCustomer},
			{Name: "MAX_USAGE_TOTAL", Value: sc.DiscountConfig.MaxUsageTotal},
			// Discount management settings
			{Name: "EXPIRATION_CHECK_INTERVAL", Value: sc.DiscountConfig.ExpirationCheckInterval},
			{Name: "AUTO_EXPIRE_ENABLED", Value: sc.DiscountConfig.AutoExpireEnabled},
			{Name: "STACKABLE_DISCOUNTS", Value: sc.DiscountConfig.StackableDiscounts},
			{Name: "EXCLUSIVE_DISCOUNTS", Value: sc.DiscountConfig.ExclusiveDiscounts},
		}
		envVars = append(envVars, discountEnvVars...)
	}

	if sc.CheckoutConfig != nil {
		checkoutEnvVars := []corev1.EnvVar{
			// Checkout timeout settings
			{Name: "CHECKOUT_TIMEOUT", Value: sc.CheckoutConfig.CheckoutTimeout},
			{Name: "PAYMENT_TIMEOUT", Value: sc.CheckoutConfig.PaymentTimeout},
			{Name: "INVENTORY_TIMEOUT", Value: sc.CheckoutConfig.InventoryTimeout},
			{Name: "SHIPPING_TIMEOUT", Value: sc.CheckoutConfig.ShippingTimeout},
			// Checkout feature settings
			{Name: "TAX_CALCULATION_ENABLED", Value: sc.CheckoutConfig.TaxCalculationEnabled},
			{Name: "SHIPPING_CALCULATION_ENABLED", Value: sc.CheckoutConfig.ShippingCalculationEnabled},
			{Name: "GUEST_CHECKOUT_ENABLED", Value: sc.CheckoutConfig.GuestCheckoutEnabled},
			{Name: "AUTO_APPLY_DISCOUNTS", Value: sc.CheckoutConfig.AutoApplyDiscounts},
			{Name: "REQUIRE_SHIPPING_ADDRESS", Value: sc.CheckoutConfig.RequireShippingAddress},
		}
		envVars = append(envVars, checkoutEnvVars...)
	}

	if sc.AuthConfig != nil {
		authEnvVars := []corev1.EnvVar{
			// JWT and token settings
			{Name: "JWT_EXPIRATION", Value: sc.AuthConfig.JWTExpiration},
			{Name: "REFRESH_TOKEN_EXPIRATION", Value: sc.AuthConfig.RefreshTokenExpiration},
			{Name: "PASSWORD_MIN_LENGTH", Value: sc.AuthConfig.PasswordMinLength},
			{Name: "REQUIRE_EMAIL_VERIFICATION", Value: sc.AuthConfig.RequireEmailVerification},
			{Name: "MAX_LOGIN_ATTEMPTS", Value: sc.AuthConfig.MaxLoginAttempts},
			{Name: "LOCKOUT_DURATION", Value: sc.AuthConfig.LockoutDuration},
			{Name: "OAUTH_PROVIDERS", Value: sc.AuthConfig.OAuthProviders},
			{Name: "SESSION_TIMEOUT", Value: sc.AuthConfig.SessionTimeout},
			{Name: "TOKEN_REFRESH_ENABLED", Value: sc.AuthConfig.TokenRefreshEnabled},
		}
		envVars = append(envVars, authEnvVars...)

		// Add JWT secret from secret - Fixed naming to match service expectations
		envVars = append(envVars, corev1.EnvVar{
			Name: "JWT_SECRET",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: "auth-secret"},
					Key:                  "jwt-secret",
				},
			},
		})
	}

	if sc.GraphQLConfig != nil {
		graphqlEnvVars := []corev1.EnvVar{
			{Name: "GRAPHQL_PLAYGROUND_ENABLED", Value: sc.GraphQLConfig.PlaygroundEnabled},
			{Name: "GRAPHQL_INTROSPECTION_ENABLED", Value: sc.GraphQLConfig.IntrospectionEnabled},
			{Name: "GRAPHQL_COMPLEXITY_LIMIT", Value: sc.GraphQLConfig.ComplexityLimit},
			{Name: "GRAPHQL_QUERY_TIMEOUT", Value: sc.GraphQLConfig.QueryTimeout},
			{Name: "GRAPHQL_MAX_QUERY_DEPTH", Value: sc.GraphQLConfig.MaxQueryDepth},
			{Name: "GRAPHQL_MAX_QUERY_COMPLEXITY", Value: sc.GraphQLConfig.MaxQueryComplexity},
		}
		envVars = append(envVars, graphqlEnvVars...)
	}

	if sc.InvoiceConfig != nil {
		invoiceEnvVars := []corev1.EnvVar{
			{Name: "INVOICE_DEFAULT_CURRENCY", Value: sc.InvoiceConfig.DefaultCurrency},
			{Name: "INVOICE_PREFIX", Value: sc.InvoiceConfig.InvoicePrefix},
			{Name: "INVOICE_DUE_DATE_DAYS", Value: sc.InvoiceConfig.DueDateDays},
			{Name: "INVOICE_LATE_FEE_PERCENTAGE", Value: sc.InvoiceConfig.LateFeePercentage},
			{Name: "INVOICE_TAX_RATE", Value: sc.InvoiceConfig.TaxRate},
			{Name: "INVOICE_AUTO_GENERATE_ENABLED", Value: sc.InvoiceConfig.AutoGenerateEnabled},
			{Name: "INVOICE_PDF_TEMPLATE_PATH", Value: sc.InvoiceConfig.PDFTemplatePath},
			{Name: "INVOICE_EMAIL_ENABLED", Value: sc.InvoiceConfig.EmailEnabled},
		}
		envVars = append(envVars, invoiceEnvVars...)
	}

	if sc.MonolithConfig != nil {
		monolithEnvVars := []corev1.EnvVar{
			{Name: "EMBEDDED_NATS_PORT", Value: sc.MonolithConfig.EmbeddedNATSPort},
			{Name: "EMBEDDED_NATS_JETSTREAM", Value: sc.MonolithConfig.EmbeddedNATSJetStream},
			{Name: "EMBEDDED_NATS_DATA_DIR", Value: sc.MonolithConfig.EmbeddedNATSDataDir},
			{Name: "EMBEDDED_NATS_LOG_LEVEL", Value: sc.MonolithConfig.EmbeddedNATSLogLevel},
			{Name: "EMBEDDED_NATS_DEBUG", Value: sc.MonolithConfig.EmbeddedNATSDebug},
			{Name: "EMBEDDED_NATS_TRACE", Value: sc.MonolithConfig.EmbeddedNATSTrace},
		}
		envVars = append(envVars, monolithEnvVars...)
	}

	return envVars
}

// getDatabaseName returns the database name for a service and PR number
func getDatabaseName(serviceName, prNumber string) string {
	// Remove "-service" suffix if present
	dbName := serviceName
	if len(serviceName) > 8 && serviceName[len(serviceName)-8:] == "-service" {
		dbName = serviceName[:len(serviceName)-8]
	}
	return "pishop_" + dbName + "_pr_" + prNumber
}

// GetServicePorts returns the standardized ports for a service
func GetServicePorts() []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{ContainerPort: 8080, Name: "http"},
	}
}
