package controllers

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"k8s.io/client-go/tools/record"

	pishopv1alpha1 "go.pilab.hu/shop/pishop-provisioner/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	appsv1 "k8s.io/api/apps/v1"
)

func TestPRStackReconciler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PRStack Controller Suite")
}

var _ = Describe("PRStack Controller", func() {
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
		Expect(appsv1.AddToScheme(scheme)).To(Succeed())
		
		fakeClient = fake.NewClientBuilder().
			WithScheme(scheme).
			WithStatusSubresource(&pishopv1alpha1.PRStack{}).
			Build()
		
		recorder := record.NewFakeRecorder(100)
		reconciler = &PRStackReconciler{
			Client:         fakeClient,
			Scheme:         scheme,
			Recorder:       recorder,
			MongoURI:       "mongodb://localhost:27017",
			MongoUsername:  "admin",
			MongoPassword:  "password",
			BaseDomain:     "shop.pilab.hu",
		}
		
		ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	})

	AfterEach(func() {
		cancel()
	})

	Context("Reconcile", func() {
		It("should handle PRStack not found", func() {
			req := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-pr",
					Namespace: "default",
				},
			}
			
			result, err := reconciler.Reconcile(ctx, req)
			Expect(err).ToNot(HaveOccurred())
			Expect(result.Requeue).To(BeFalse())
		})

		It("should add finalizer on first reconcile", func() {
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
			
			Expect(fakeClient.Create(ctx, prStack)).To(Succeed())
			
			req := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-pr",
					Namespace: "default",
				},
			}
			
			result, err := reconciler.Reconcile(ctx, req)
			Expect(err).ToNot(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(RequeueIntervalShort))
			
			// Check if finalizer was added
			var updatedPRStack pishopv1alpha1.PRStack
			Expect(fakeClient.Get(ctx, types.NamespacedName{Name: "test-pr", Namespace: "default"}, &updatedPRStack)).To(Succeed())
			Expect(updatedPRStack.Finalizers).To(ContainElement(FinalizerName))
		})

		It("should set creation time on first reconcile", func() {
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
			
			Expect(fakeClient.Create(ctx, prStack)).To(Succeed())
			
			req := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-pr",
					Namespace: "default",
				},
			}
			
			_, err := reconciler.Reconcile(ctx, req)
			Expect(err).ToNot(HaveOccurred())
			
			// Check if creation time was set
			var updatedPRStack pishopv1alpha1.PRStack
			Expect(fakeClient.Get(ctx, types.NamespacedName{Name: "test-pr", Namespace: "default"}, &updatedPRStack)).To(Succeed())
			Expect(updatedPRStack.Status.CreatedAt).ToNot(BeNil())
			Expect(updatedPRStack.Status.LastActiveAt).ToNot(BeNil())
		})

		It("should handle inactive stack", func() {
			prStack := &pishopv1alpha1.PRStack{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pr",
					Namespace: "default",
				},
				Spec: pishopv1alpha1.PRStackSpec{
					PRNumber: "123",
					Active:   false,
				},
				Status: pishopv1alpha1.PRStackStatus{
					CreatedAt: &metav1.Time{Time: time.Now()},
				},
			}
			
			Expect(fakeClient.Create(ctx, prStack)).To(Succeed())
			
			req := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-pr",
					Namespace: "default",
				},
			}
			
			result, err := reconciler.Reconcile(ctx, req)
			Expect(err).ToNot(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(RequeueIntervalLong))
			
			// Check if phase was set to Inactive
			var updatedPRStack pishopv1alpha1.PRStack
			Expect(fakeClient.Get(ctx, types.NamespacedName{Name: "test-pr", Namespace: "default"}, &updatedPRStack)).To(Succeed())
			Expect(updatedPRStack.Status.Phase).To(Equal(PhaseInactive))
		})

		It("should handle stack expiration", func() {
			oldTime := time.Now().Add(-2 * time.Hour) // 2 hours ago
			prStack := &pishopv1alpha1.PRStack{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pr",
					Namespace: "default",
				},
				Spec: pishopv1alpha1.PRStackSpec{
					PRNumber: "123",
					Active:   true,
				},
				Status: pishopv1alpha1.PRStackStatus{
					CreatedAt:     &metav1.Time{Time: oldTime},
					LastActiveAt:  &metav1.Time{Time: oldTime},
				},
			}
			
			Expect(fakeClient.Create(ctx, prStack)).To(Succeed())
			
			req := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-pr",
					Namespace: "default",
				},
			}
			
			result, err := reconciler.Reconcile(ctx, req)
			Expect(err).ToNot(HaveOccurred())
			// The stack gets reactivated instead of being set to inactive
			Expect(result.RequeueAfter).To(Equal(RequeueIntervalShort))
			
			// Check if LastActiveAt was updated (reactivation)
			var updatedPRStack pishopv1alpha1.PRStack
			Expect(fakeClient.Get(ctx, types.NamespacedName{Name: "test-pr", Namespace: "default"}, &updatedPRStack)).To(Succeed())
			Expect(updatedPRStack.Status.LastActiveAt).ToNot(BeNil())
			Expect(updatedPRStack.Status.LastActiveAt.Time.After(oldTime)).To(BeTrue())
		})
	})

	Context("Helper Functions", func() {
		It("should generate correct namespace name", func() {
			namespace := reconciler.getNamespaceName("123")
			Expect(namespace).To(Equal("pr-123-shop-pilab-hu"))
		})

		It("should return custom domain when specified", func() {
			prStack := &pishopv1alpha1.PRStack{
				Spec: pishopv1alpha1.PRStackSpec{
					CustomDomain: "custom.example.com",
				},
			}
			
			domain := reconciler.getDomain(prStack)
			Expect(domain).To(Equal("custom.example.com"))
		})

		It("should return default domain when custom domain not specified", func() {
			prStack := &pishopv1alpha1.PRStack{
				Spec: pishopv1alpha1.PRStackSpec{
					PRNumber: "123",
				},
			}
			
			domain := reconciler.getDomain(prStack)
			Expect(domain).To(Equal("pr-123.shop.pilab.hu"))
		})

		It("should detect stack expiration correctly", func() {
			oldTime := time.Now().Add(-2 * time.Hour)
			prStack := &pishopv1alpha1.PRStack{
				Status: pishopv1alpha1.PRStackStatus{
					LastActiveAt: &metav1.Time{Time: oldTime},
				},
			}
			
			Expect(reconciler.isStackExpired(prStack)).To(BeTrue())
		})

		It("should not detect stack as expired when recently active", func() {
			recentTime := time.Now().Add(-30 * time.Minute)
			prStack := &pishopv1alpha1.PRStack{
				Status: pishopv1alpha1.PRStackStatus{
					LastActiveAt: &metav1.Time{Time: recentTime},
				},
			}
			
			Expect(reconciler.isStackExpired(prStack)).To(BeFalse())
		})

		It("should detect rollout needed when deployedAt changes", func() {
			now := metav1.Now()
			prStack := &pishopv1alpha1.PRStack{
				Spec: pishopv1alpha1.PRStackSpec{
					DeployedAt: &now,
				},
				Status: pishopv1alpha1.PRStackStatus{
					LastDeployedAt: &metav1.Time{Time: now.Add(-time.Hour)},
				},
			}
			
			Expect(reconciler.shouldRolloutDeployments(prStack)).To(BeTrue())
		})

		It("should not detect rollout needed when deployedAt unchanged", func() {
			now := metav1.Now()
			prStack := &pishopv1alpha1.PRStack{
				Spec: pishopv1alpha1.PRStackSpec{
					DeployedAt: &now,
				},
				Status: pishopv1alpha1.PRStackStatus{
					LastDeployedAt: &now,
				},
			}
			
			Expect(reconciler.shouldRolloutDeployments(prStack)).To(BeFalse())
		})
	})

	Context("String Helper Functions", func() {
		It("should contain string correctly", func() {
			slice := []string{"a", "b", "c"}
			Expect(containsString(slice, "b")).To(BeTrue())
			Expect(containsString(slice, "d")).To(BeFalse())
		})

		It("should remove string correctly", func() {
			slice := []string{"a", "b", "c"}
			result := removeString(slice, "b")
			Expect(result).To(Equal([]string{"a", "c"}))
			// Original slice should be unchanged
			Expect(slice).To(Equal([]string{"a", "b", "c"}))
		})

		It("should not modify slice if string not found", func() {
			slice := []string{"a", "b", "c"}
			result := removeString(slice, "d")
			Expect(result).To(Equal([]string{"a", "b", "c"}))
		})
	})
})
