package controllers

// DefaultServices defines the list of services that are provisioned by default
var DefaultServices = []string{
	"product-service",
	"cart-service",
	"order-service",
	"payment-service",
	"customer-service",
	"inventory-service",
	"notification-service",
	"discount-service",
	"checkout-service",
	"analytics-service",
	"auth-service",
	"graphql-service",
}

// DefaultServicesString is the comma-separated string of default services
const DefaultServicesString = "product-service,cart-service,order-service,payment-service,customer-service,inventory-service,notification-service,discount-service,checkout-service,analytics-service,auth-service,graphql-service"
