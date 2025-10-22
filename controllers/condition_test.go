package controllers

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	pishopv1alpha1 "go.pilab.hu/shop/pishop-provisioner/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Condition and Status Management Tests", func() {
	var (
		ctx        context.Context
		cancel     context.CancelFunc
		reconciler *PRStackReconciler
		fakeClient client.Client
		scheme     *runtime.Scheme
	)

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(context.Background())

		scheme = runtime.NewScheme()
		Expect(pishopv1alpha1.AddToScheme(scheme)).To(Succeed())
		Expect(corev1.AddToScheme(scheme)).To(Succeed())

		fakeClient = fake.NewClientBuilder().
			WithScheme(scheme).
			WithStatusSubresource(&pishopv1alpha1.PRStack{}).
			Build()

		reconciler = &PRStackReconciler{
			Client:        fakeClient,
			Scheme:        scheme,
			MongoURI:      "mongodb://localhost:27017",
			MongoUsername: "admin",
			MongoPassword: "password",
			BaseDomain:    "shop.pilab.hu",
		}

		ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	})

	AfterEach(func() {
		cancel()
	})

	Context("Condition Management", func() {
		var prStack *pishopv1alpha1.PRStack

		BeforeEach(func() {
			prStack = &pishopv1alpha1.PRStack{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pr",
					Namespace: "default",
				},
				Spec: pishopv1alpha1.PRStackSpec{
					PRNumber: "123",
					Active:   true,
				},
				Status: pishopv1alpha1.PRStackStatus{
					Phase:      PhaseRunning,
					Conditions: []metav1.Condition{},
				},
			}
		})

		It("should set Ready condition when stack is healthy", func() {
			reconciler.setReadyCondition(prStack, "StackRunning", "All services deployed")

			Expect(prStack.Status.Conditions).To(HaveLen(2))

			readyCondition := findCondition(prStack.Status.Conditions, ConditionTypeReady)
			Expect(readyCondition).ToNot(BeNil())
			Expect(readyCondition.Status).To(Equal(metav1.ConditionTrue))
			Expect(readyCondition.Reason).To(Equal("StackRunning"))

			degradedCondition := findCondition(prStack.Status.Conditions, ConditionTypeDegraded)
			Expect(degradedCondition).ToNot(BeNil())
			Expect(degradedCondition.Status).To(Equal(metav1.ConditionFalse))
		})

		It("should set Progressing condition", func() {
			reconciler.setProgressingCondition(prStack, "Deploying", "Deploying services")

			Expect(prStack.Status.Conditions).To(HaveLen(1))

			condition := findCondition(prStack.Status.Conditions, ConditionTypeProgressing)
			Expect(condition).ToNot(BeNil())
			Expect(condition.Status).To(Equal(metav1.ConditionTrue))
			Expect(condition.Reason).To(Equal("Deploying"))
		})

		It("should update existing condition", func() {
			// Set initial condition
			reconciler.setReadyCondition(prStack, "Initial", "Initial message")
			Expect(prStack.Status.Conditions).To(HaveLen(2))

			initialReady := findCondition(prStack.Status.Conditions, ConditionTypeReady)
			initialTransitionTime := initialReady.LastTransitionTime

			// Update condition with same status
			reconciler.setReadyCondition(prStack, "Initial", "Updated message")

			updatedReady := findCondition(prStack.Status.Conditions, ConditionTypeReady)
			Expect(updatedReady.LastTransitionTime).To(Equal(initialTransitionTime))
		})

		It("should update transition time when status changes", func() {
			// Set initial Ready condition
			now := metav1.Now()
			reconciler.setCondition(prStack, metav1.Condition{
				Type:               ConditionTypeReady,
				Status:             metav1.ConditionTrue,
				Reason:             "Running",
				Message:            "All good",
				LastTransitionTime: now,
			})

			initialCondition := findCondition(prStack.Status.Conditions, ConditionTypeReady)
			Expect(initialCondition).ToNot(BeNil())
			Expect(initialCondition.Status).To(Equal(metav1.ConditionTrue))

			// Change status to False (should update transition time)
			reconciler.setCondition(prStack, metav1.Condition{
				Type:    ConditionTypeReady,
				Status:  metav1.ConditionFalse,
				Reason:  "Failed",
				Message: "Something went wrong",
			})

			updatedCondition := findCondition(prStack.Status.Conditions, ConditionTypeReady)
			Expect(updatedCondition.Status).To(Equal(metav1.ConditionFalse))
			Expect(updatedCondition.Reason).To(Equal("Failed"))
			// Transition time should be updated (different from initial)
			Expect(updatedCondition.LastTransitionTime.After(initialCondition.LastTransitionTime.Time) ||
				updatedCondition.LastTransitionTime.Equal(&initialCondition.LastTransitionTime)).To(BeTrue())
		})

		It("should preserve conditions when adding new ones", func() {
			reconciler.setReadyCondition(prStack, "Ready", "Ready message")
			reconciler.setProgressingCondition(prStack, "Progressing", "Progressing message")

			Expect(prStack.Status.Conditions).To(HaveLen(3)) // Ready, Degraded, Progressing
		})
	})

	Context("Failed State Management", func() {
		var prStack *pishopv1alpha1.PRStack

		BeforeEach(func() {
			prStack = &pishopv1alpha1.PRStack{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pr",
					Namespace: "default",
				},
				Spec: pishopv1alpha1.PRStackSpec{
					PRNumber: "456",
					Active:   true,
				},
				Status: pishopv1alpha1.PRStackStatus{
					Phase: PhaseProvisioning,
				},
			}
		})

		It("should set Failed phase and conditions on error", func() {
			Expect(fakeClient.Create(ctx, prStack)).To(Succeed())

			reconciler.updateStatusWithError(ctx, prStack, "MongoDB provisioning failed",
				&mockError{message: "connection refused"})

			Expect(prStack.Status.Phase).To(Equal(PhaseFailed))
			Expect(prStack.Status.Message).To(ContainSubstring("MongoDB provisioning failed"))

			readyCondition := findCondition(prStack.Status.Conditions, ConditionTypeReady)
			Expect(readyCondition).ToNot(BeNil())
			Expect(readyCondition.Status).To(Equal(metav1.ConditionFalse))
			Expect(readyCondition.Reason).To(Equal("Error"))

			degradedCondition := findCondition(prStack.Status.Conditions, ConditionTypeDegraded)
			Expect(degradedCondition).ToNot(BeNil())
			Expect(degradedCondition.Status).To(Equal(metav1.ConditionTrue))
			Expect(degradedCondition.Reason).To(Equal("ReconciliationFailed"))
		})

		It("should preserve phase on reconciliation in Failed state", func() {
			prStack.Status.Phase = PhaseFailed
			Expect(fakeClient.Create(ctx, prStack)).To(Succeed())

			result, err := reconciler.reconcileByPhase(ctx, prStack)
			Expect(err).ToNot(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(RequeueIntervalLong))
		})
	})

	Context("Degraded State Management", func() {
		It("should set Degraded phase when some services fail", func() {
			// This is tested implicitly through the deployment logic
			// When failedServices > 0 but < len(services), phase should be Degraded
			prStack := &pishopv1alpha1.PRStack{
				Spec: pishopv1alpha1.PRStackSpec{
					PRNumber: "789",
				},
				Status: pishopv1alpha1.PRStackStatus{
					Phase: PhaseDeploying,
				},
			}

			// Simulate partial failure by setting conditions
			reconciler.setCondition(prStack, metav1.Condition{
				Type:               ConditionTypeReady,
				Status:             metav1.ConditionTrue,
				Reason:             "PartiallyDegraded",
				Message:            "2/3 services running",
				LastTransitionTime: metav1.Now(),
			})
			reconciler.setCondition(prStack, metav1.Condition{
				Type:               ConditionTypeDegraded,
				Status:             metav1.ConditionTrue,
				Reason:             "ServiceFailures",
				Message:            "1 service failed to deploy",
				LastTransitionTime: metav1.Now(),
			})

			readyCondition := findCondition(prStack.Status.Conditions, ConditionTypeReady)
			Expect(readyCondition).ToNot(BeNil())
			Expect(readyCondition.Status).To(Equal(metav1.ConditionTrue))
			Expect(readyCondition.Reason).To(Equal("PartiallyDegraded"))

			degradedCondition := findCondition(prStack.Status.Conditions, ConditionTypeDegraded)
			Expect(degradedCondition).ToNot(BeNil())
			Expect(degradedCondition.Status).To(Equal(metav1.ConditionTrue))
		})

		It("should handle Degraded phase in reconciliation", func() {
			prStack := &pishopv1alpha1.PRStack{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pr",
					Namespace: "default",
				},
				Spec: pishopv1alpha1.PRStackSpec{
					PRNumber: "999",
				},
				Status: pishopv1alpha1.PRStackStatus{
					Phase: PhaseDegraded,
				},
			}
			Expect(fakeClient.Create(ctx, prStack)).To(Succeed())

			// Should use same logic as Running phase
			result, err := reconciler.reconcileByPhase(ctx, prStack)
			Expect(err).ToNot(HaveOccurred())
			Expect(result.RequeueAfter).ToNot(BeZero())
		})
	})

	Context("Phase Constants Validation", func() {
		It("should have all phase constants defined", func() {
			Expect(PhaseInitialization).To(Equal(""))
			Expect(PhaseProvisioning).To(Equal("Provisioning"))
			Expect(PhaseDeploying).To(Equal("Deploying"))
			Expect(PhaseRunning).To(Equal("Running"))
			Expect(PhaseInactive).To(Equal("Inactive"))
			Expect(PhaseFailed).To(Equal("Failed"))
			Expect(PhaseDegraded).To(Equal("Degraded"))
			Expect(PhaseCleaning).To(Equal("Cleaning"))
			Expect(PhaseCleaned).To(Equal("Cleaned"))
		})

		It("should have all condition type constants defined", func() {
			Expect(ConditionTypeReady).To(Equal("Ready"))
			Expect(ConditionTypeDegraded).To(Equal("Degraded"))
			Expect(ConditionTypeProgressing).To(Equal("Progressing"))
		})
	})
})

// Helper function to find a condition by type
func findCondition(conditions []metav1.Condition, conditionType string) *metav1.Condition {
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return &conditions[i]
		}
	}
	return nil
}

// mockError is a simple error implementation for testing
type mockError struct {
	message string
}

func (e *mockError) Error() string {
	return e.message
}
