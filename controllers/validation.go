package controllers

import (
	"fmt"
	"regexp"
	"strings"

	pishopv1alpha1 "go.pilab.hu/shop/pishop-provisioner/api/v1alpha1"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}

// ValidatePRStack validates a PRStack resource
func ValidatePRStack(prStack *pishopv1alpha1.PRStack) error {
	var errors []error

	// Validate PR number
	if err := validatePRNumber(prStack.Spec.PRNumber); err != nil {
		errors = append(errors, err)
	}

	// Validate image tag if provided
	if prStack.Spec.ImageTag != "" {
		if err := validateImageTag(prStack.Spec.ImageTag); err != nil {
			errors = append(errors, err)
		}
	}

	// Validate custom domain if provided
	if prStack.Spec.CustomDomain != "" {
		if err := validateCustomDomain(prStack.Spec.CustomDomain); err != nil {
			errors = append(errors, err)
		}
	}

	// Validate resource limits if provided
	if prStack.Spec.ResourceLimits != nil {
		if err := validateResourceLimits(prStack.Spec.ResourceLimits); err != nil {
			errors = append(errors, err)
		}
	}

	// Validate backup config if provided
	if prStack.Spec.BackupConfig != nil {
		if err := validateBackupConfig(prStack.Spec.BackupConfig); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed: %v", errors)
	}

	return nil
}

// validatePRNumber validates the PR number format
func validatePRNumber(prNumber string) error {
	if prNumber == "" {
		return &ValidationError{Field: "prNumber", Message: "PR number is required"}
	}

	// PR number should be numeric
	matched, err := regexp.MatchString(`^\d+$`, prNumber)
	if err != nil {
		return &ValidationError{Field: "prNumber", Message: "invalid regex pattern"}
	}

	if !matched {
		return &ValidationError{Field: "prNumber", Message: "PR number must be numeric"}
	}

	// PR number should be reasonable length (1-10 digits)
	if len(prNumber) > 10 {
		return &ValidationError{Field: "prNumber", Message: "PR number too long (max 10 digits)"}
	}

	return nil
}

// validateImageTag validates the Docker image tag format
func validateImageTag(imageTag string) error {
	if imageTag == "" {
		return &ValidationError{Field: "imageTag", Message: "image tag cannot be empty"}
	}

	// Image tag should be valid Docker tag format
	matched, err := regexp.MatchString(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`, imageTag)
	if err != nil {
		return &ValidationError{Field: "imageTag", Message: "invalid regex pattern"}
	}

	if !matched {
		return &ValidationError{Field: "imageTag", Message: "image tag contains invalid characters"}
	}

	// Image tag should not be too long
	if len(imageTag) > 128 {
		return &ValidationError{Field: "imageTag", Message: "image tag too long (max 128 characters)"}
	}

	return nil
}

// validateCustomDomain validates the custom domain format
func validateCustomDomain(domain string) error {
	if domain == "" {
		return &ValidationError{Field: "customDomain", Message: "custom domain cannot be empty"}
	}

	// Basic domain validation
	matched, err := regexp.MatchString(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`, domain)
	if err != nil {
		return &ValidationError{Field: "customDomain", Message: "invalid regex pattern"}
	}

	if !matched {
		return &ValidationError{Field: "customDomain", Message: "invalid domain format"}
	}

	// Domain should not be too long
	if len(domain) > 253 {
		return &ValidationError{Field: "customDomain", Message: "domain too long (max 253 characters)"}
	}

	return nil
}

// validateResourceLimits validates resource limits
func validateResourceLimits(limits *pishopv1alpha1.ResourceLimits) error {
	if limits.CPULimit != "" {
		if err := validateResourceQuantity(limits.CPULimit, "cpuLimit"); err != nil {
			return err
		}
	}

	if limits.MemoryLimit != "" {
		if err := validateResourceQuantity(limits.MemoryLimit, "memoryLimit"); err != nil {
			return err
		}
	}

	if limits.StorageLimit != "" {
		if err := validateResourceQuantity(limits.StorageLimit, "storageLimit"); err != nil {
			return err
		}
	}

	return nil
}

// validateResourceQuantity validates Kubernetes resource quantity format
func validateResourceQuantity(quantity, fieldName string) error {
	if quantity == "" {
		return &ValidationError{Field: fieldName, Message: "resource quantity cannot be empty"}
	}

	// Basic validation for Kubernetes resource quantities
	matched, err := regexp.MatchString(`^[0-9]+(\.[0-9]+)?[a-zA-Z]*$`, quantity)
	if err != nil {
		return &ValidationError{Field: fieldName, Message: "invalid regex pattern"}
	}

	if !matched {
		return &ValidationError{Field: fieldName, Message: "invalid resource quantity format"}
	}

	return nil
}

// validateBackupConfig validates backup configuration
func validateBackupConfig(config *pishopv1alpha1.BackupConfig) error {
	if config.Enabled && config.Schedule != "" {
		// Basic cron validation (simplified)
		parts := strings.Fields(config.Schedule)
		if len(parts) != 5 {
			return &ValidationError{Field: "backupConfig.schedule", Message: "invalid cron schedule format (expected 5 fields)"}
		}
	}

	if config.RetentionDays < 0 {
		return &ValidationError{Field: "backupConfig.retentionDays", Message: "retention days cannot be negative"}
	}

	if config.RetentionDays > 3650 { // 10 years
		return &ValidationError{Field: "backupConfig.retentionDays", Message: "retention days too high (max 3650)"}
	}

	if config.StorageSize != "" {
		if err := validateResourceQuantity(config.StorageSize, "backupConfig.storageSize"); err != nil {
			return err
		}
	}

	return nil
}
