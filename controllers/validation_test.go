package controllers

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	pishopv1alpha1 "go.pilab.hu/shop/pishop-provisioner/api/v1alpha1"
)

func TestValidation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Validation Suite")
}

var _ = Describe("Validation", func() {
	Context("ValidatePRStack", func() {
		It("should validate valid PRStack", func() {
			prStack := &pishopv1alpha1.PRStack{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pr",
					Namespace: "default",
				},
				Spec: pishopv1alpha1.PRStackSpec{
					PRNumber: "123",
					Active:   true,
				},
			}
			
			err := ValidatePRStack(prStack)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should fail validation for empty PR number", func() {
			prStack := &pishopv1alpha1.PRStack{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pr",
					Namespace: "default",
				},
				Spec: pishopv1alpha1.PRStackSpec{
					PRNumber: "",
					Active:   true,
				},
			}
			
			err := ValidatePRStack(prStack)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("PR number is required"))
		})

		It("should fail validation for non-numeric PR number", func() {
			prStack := &pishopv1alpha1.PRStack{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pr",
					Namespace: "default",
				},
				Spec: pishopv1alpha1.PRStackSpec{
					PRNumber: "abc123",
					Active:   true,
				},
			}
			
			err := ValidatePRStack(prStack)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("PR number must be numeric"))
		})

		It("should fail validation for too long PR number", func() {
			prStack := &pishopv1alpha1.PRStack{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pr",
					Namespace: "default",
				},
				Spec: pishopv1alpha1.PRStackSpec{
					PRNumber: "12345678901", // 11 digits
					Active:   true,
				},
			}
			
			err := ValidatePRStack(prStack)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("PR number too long"))
		})

		It("should validate valid image tag", func() {
			prStack := &pishopv1alpha1.PRStack{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pr",
					Namespace: "default",
				},
				Spec: pishopv1alpha1.PRStackSpec{
					PRNumber: "123",
					ImageTag: "v1.2.3",
					Active:   true,
				},
			}
			
			err := ValidatePRStack(prStack)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should fail validation for invalid image tag", func() {
			prStack := &pishopv1alpha1.PRStack{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pr",
					Namespace: "default",
				},
				Spec: pishopv1alpha1.PRStackSpec{
					PRNumber: "123",
					ImageTag: "invalid@tag#",
					Active:   true,
				},
			}
			
			err := ValidatePRStack(prStack)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("image tag contains invalid characters"))
		})

		It("should validate valid custom domain", func() {
			prStack := &pishopv1alpha1.PRStack{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pr",
					Namespace: "default",
				},
				Spec: pishopv1alpha1.PRStackSpec{
					PRNumber:     "123",
					CustomDomain: "example.com",
					Active:       true,
				},
			}
			
			err := ValidatePRStack(prStack)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should fail validation for invalid custom domain", func() {
			prStack := &pishopv1alpha1.PRStack{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pr",
					Namespace: "default",
				},
				Spec: pishopv1alpha1.PRStackSpec{
					PRNumber:     "123",
					CustomDomain: "invalid..domain",
					Active:       true,
				},
			}
			
			err := ValidatePRStack(prStack)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid domain format"))
		})

		It("should validate valid resource limits", func() {
			prStack := &pishopv1alpha1.PRStack{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pr",
					Namespace: "default",
				},
				Spec: pishopv1alpha1.PRStackSpec{
					PRNumber: "123",
					Active:   true,
					ResourceLimits: &pishopv1alpha1.ResourceLimits{
						CPULimit:    "500m",
						MemoryLimit: "512Mi",
					},
				},
			}
			
			err := ValidatePRStack(prStack)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should fail validation for invalid resource limits", func() {
			prStack := &pishopv1alpha1.PRStack{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pr",
					Namespace: "default",
				},
				Spec: pishopv1alpha1.PRStackSpec{
					PRNumber: "123",
					Active:   true,
					ResourceLimits: &pishopv1alpha1.ResourceLimits{
						CPULimit:    "invalid-cpu",
						MemoryLimit: "512Mi",
					},
				},
			}
			
			err := ValidatePRStack(prStack)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid resource quantity format"))
		})

		It("should validate valid backup config", func() {
			prStack := &pishopv1alpha1.PRStack{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pr",
					Namespace: "default",
				},
				Spec: pishopv1alpha1.PRStackSpec{
					PRNumber: "123",
					Active:   true,
					BackupConfig: &pishopv1alpha1.BackupConfig{
						Enabled:       true,
						Schedule:      "0 0 * * *",
						RetentionDays: 30,
						StorageSize:   "10Gi",
					},
				},
			}
			
			err := ValidatePRStack(prStack)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should fail validation for invalid backup config", func() {
			prStack := &pishopv1alpha1.PRStack{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pr",
					Namespace: "default",
				},
				Spec: pishopv1alpha1.PRStackSpec{
					PRNumber: "123",
					Active:   true,
					BackupConfig: &pishopv1alpha1.BackupConfig{
						Enabled:       true,
						Schedule:      "invalid-cron",
						RetentionDays: -1,
						StorageSize:   "invalid-size",
					},
				},
			}
			
			err := ValidatePRStack(prStack)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid cron schedule format"))
		})
	})

	Context("validatePRNumber", func() {
		It("should validate numeric PR numbers", func() {
			Expect(validatePRNumber("123")).ToNot(HaveOccurred())
			Expect(validatePRNumber("1")).ToNot(HaveOccurred())
			Expect(validatePRNumber("9999999999")).ToNot(HaveOccurred()) // 10 digits
		})

		It("should reject non-numeric PR numbers", func() {
			Expect(validatePRNumber("abc")).To(HaveOccurred())
			Expect(validatePRNumber("123abc")).To(HaveOccurred())
			Expect(validatePRNumber("abc123")).To(HaveOccurred())
		})

		It("should reject empty PR numbers", func() {
			Expect(validatePRNumber("")).To(HaveOccurred())
		})

		It("should reject too long PR numbers", func() {
			Expect(validatePRNumber("12345678901")).To(HaveOccurred()) // 11 digits
		})
	})

	Context("validateImageTag", func() {
		It("should validate valid image tags", func() {
			Expect(validateImageTag("v1.2.3")).ToNot(HaveOccurred())
			Expect(validateImageTag("latest")).ToNot(HaveOccurred())
			Expect(validateImageTag("pr-123-abc123")).ToNot(HaveOccurred())
			Expect(validateImageTag("1.0.0-beta.1")).ToNot(HaveOccurred())
		})

		It("should reject invalid image tags", func() {
			Expect(validateImageTag("")).To(HaveOccurred())
			Expect(validateImageTag("invalid@tag")).To(HaveOccurred())
			Expect(validateImageTag("invalid#tag")).To(HaveOccurred())
			Expect(validateImageTag("invalid/tag")).To(HaveOccurred())
		})
	})

	Context("validateCustomDomain", func() {
		It("should validate valid domains", func() {
			Expect(validateCustomDomain("example.com")).ToNot(HaveOccurred())
			Expect(validateCustomDomain("sub.example.com")).ToNot(HaveOccurred())
			Expect(validateCustomDomain("test-domain.example.org")).ToNot(HaveOccurred())
		})

		It("should reject invalid domains", func() {
			Expect(validateCustomDomain("")).To(HaveOccurred())
			Expect(validateCustomDomain("invalid..domain")).To(HaveOccurred())
			Expect(validateCustomDomain("-invalid.com")).To(HaveOccurred())
			Expect(validateCustomDomain("invalid-.com")).To(HaveOccurred())
		})
	})
})
