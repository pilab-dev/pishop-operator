package controllers

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

func TestServiceConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Service Config Suite")
}

var _ = Describe("Service Config", func() {
	Context("GetServiceConfig", func() {
		It("should return correct config for product-service", func() {
			config := GetServiceConfig("product-service", "123")
			
			Expect(config.ServiceName).To(Equal("product-service"))
			Expect(config.Environment).To(Equal("pr"))
			Expect(config.MongoDBDatabase).To(Equal("pishop_product_pr_123"))
			Expect(config.SecurityConfig).ToNot(BeNil())
			Expect(config.SecurityConfig.TLSEnabled).To(Equal("false"))
		})

		It("should return correct config for cart-service", func() {
			config := GetServiceConfig("cart-service", "456")
			
			Expect(config.ServiceName).To(Equal("cart-service"))
			Expect(config.MongoDBDatabase).To(Equal("pishop_cart_pr_456"))
			Expect(config.CartConfig).ToNot(BeNil())
			Expect(config.CartConfig.GuestExpiration).To(Equal("24h"))
			Expect(config.CartConfig.MaxItems).To(Equal("100"))
		})

		It("should return correct config for payment-service", func() {
			config := GetServiceConfig("payment-service", "789")
			
			Expect(config.ServiceName).To(Equal("payment-service"))
			Expect(config.MongoDBDatabase).To(Equal("pishop_payment_pr_789"))
			Expect(config.PaymentConfig).ToNot(BeNil())
			Expect(config.PaymentConfig.StripeEnabled).To(Equal("false"))
			Expect(config.PaymentConfig.Currency).To(Equal("USD"))
		})

		It("should return correct config for unknown service", func() {
			config := GetServiceConfig("unknown-service", "123")
			
			Expect(config.ServiceName).To(Equal("unknown-service"))
			Expect(config.MongoDBDatabase).To(Equal("pishop_unknown_pr_123"))
			Expect(config.SecurityConfig).To(BeNil())
			Expect(config.CartConfig).To(BeNil())
		})
	})

	Context("ToEnvVars", func() {
		It("should generate correct environment variables", func() {
			config := GetServiceConfig("product-service", "123")
			envVars := config.ToEnvVars()
			
			// Check basic service identification
			Expect(envVars).To(ContainElement(corev1.EnvVar{Name: "SERVICE_NAME", Value: "product-service"}))
			Expect(envVars).To(ContainElement(corev1.EnvVar{Name: "ENVIRONMENT", Value: "pr"}))
			Expect(envVars).To(ContainElement(corev1.EnvVar{Name: "LOG_LEVEL", Value: "info"}))
			
			// Check MongoDB configuration
			Expect(envVars).To(ContainElement(corev1.EnvVar{Name: "MONGODB_URI", Value: "mongodb://mongodb-service:27017"}))
			Expect(envVars).To(ContainElement(corev1.EnvVar{Name: "MONGODB_DATABASE", Value: "pishop_product_pr_123"}))
			
			// Check Redis configuration
			Expect(envVars).To(ContainElement(corev1.EnvVar{Name: "REDIS_URL", Value: "redis://redis-service:6379"}))
			
			// Check NATS configuration
			Expect(envVars).To(ContainElement(corev1.EnvVar{Name: "NATS_URL", Value: "nats://nats-service:4222"}))
		})

		It("should include MongoDB credentials from secret", func() {
			config := GetServiceConfig("product-service", "123")
			envVars := config.ToEnvVars()
			
			// Find MongoDB username env var
			var mongoUsernameVar corev1.EnvVar
			for _, envVar := range envVars {
				if envVar.Name == "MONGODB_USERNAME" {
					mongoUsernameVar = envVar
					break
				}
			}
			Expect(mongoUsernameVar.Name).To(Equal("MONGODB_USERNAME"))
			Expect(mongoUsernameVar.ValueFrom).ToNot(BeNil())
			Expect(mongoUsernameVar.ValueFrom.SecretKeyRef).ToNot(BeNil())
			Expect(mongoUsernameVar.ValueFrom.SecretKeyRef.Name).To(Equal("mongodb-secret"))
			Expect(mongoUsernameVar.ValueFrom.SecretKeyRef.Key).To(Equal("username"))
		})

		It("should include service-specific environment variables for cart-service", func() {
			config := GetServiceConfig("cart-service", "123")
			envVars := config.ToEnvVars()
			
			Expect(envVars).To(ContainElement(corev1.EnvVar{Name: "CART_GUEST_EXPIRATION", Value: "24h"}))
			Expect(envVars).To(ContainElement(corev1.EnvVar{Name: "CART_MAX_ITEMS", Value: "100"}))
			Expect(envVars).To(ContainElement(corev1.EnvVar{Name: "CART_CURRENCY", Value: "USD"}))
		})
	})

	Context("getDatabaseName", func() {
		It("should remove -service suffix", func() {
			dbName := getDatabaseName("product-service", "123")
			Expect(dbName).To(Equal("pishop_product_pr_123"))
		})

		It("should handle service name without -service suffix", func() {
			dbName := getDatabaseName("product", "123")
			Expect(dbName).To(Equal("pishop_product_pr_123"))
		})

		It("should handle empty service name", func() {
			dbName := getDatabaseName("", "123")
			Expect(dbName).To(Equal("pishop__pr_123"))
		})
	})

	Context("GetServicePorts", func() {
		It("should return correct ports", func() {
			ports := GetServicePorts()
			
			Expect(ports).To(HaveLen(1))
			Expect(ports[0].ContainerPort).To(Equal(int32(8080)))
			Expect(ports[0].Name).To(Equal("http"))
		})
	})
})
