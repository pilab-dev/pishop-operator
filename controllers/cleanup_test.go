package controllers

import (
	"context"
	"testing"

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

func TestCleanup(t *testing.T) {
	// Test suite is now managed by suite_test.go
}

var _ = Describe("Cleanup", func() {
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
			Client:         fakeClient,
			Scheme:         scheme,
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

	Context("cleanupServices", func() {
		It("should delete namespace successfully", func() {
			prStack := &pishopv1alpha1.PRStack{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pr",
					Namespace: "default",
				},
				Spec: pishopv1alpha1.PRStackSpec{
					PRNumber: "123",
				},
			}
			
			// Create namespace first
			namespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pr-123-shop-pilab-hu",
				},
			}
			Expect(fakeClient.Create(ctx, namespace)).To(Succeed())
			
			err := reconciler.cleanupServices(ctx, prStack)
			Expect(err).ToNot(HaveOccurred())
			
			// Check if namespace was deleted
			var deletedNamespace corev1.Namespace
			err = fakeClient.Get(ctx, client.ObjectKey{Name: "pr-123-shop-pilab-hu"}, &deletedNamespace)
			Expect(err).To(HaveOccurred()) // Should not exist
		})

		It("should not fail if namespace doesn't exist", func() {
			prStack := &pishopv1alpha1.PRStack{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pr",
					Namespace: "default",
				},
				Spec: pishopv1alpha1.PRStackSpec{
					PRNumber: "123",
				},
			}
			
			err := reconciler.cleanupServices(ctx, prStack)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("cleanupAllResources", func() {
		It("should complete cleanup successfully", func() {
			prStack := &pishopv1alpha1.PRStack{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pr",
					Namespace: "default",
				},
				Spec: pishopv1alpha1.PRStackSpec{
					PRNumber: "123",
				},
			}
			
			err := reconciler.cleanupAllResources(ctx, prStack)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("cleanupBackupPVC", func() {
		It("should delete backup PVC successfully", func() {
			prStack := &pishopv1alpha1.PRStack{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pr",
					Namespace: "default",
				},
				Spec: pishopv1alpha1.PRStackSpec{
					PRNumber: "123",
				},
			}
			
			// Create PVC first
			pvc := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mongodb-backup-pvc",
					Namespace: "pr-123-shop-pilab-hu",
				},
			}
			Expect(fakeClient.Create(ctx, pvc)).To(Succeed())
			
			err := reconciler.cleanupBackupPVC(ctx, prStack, "pr-123-shop-pilab-hu")
			Expect(err).ToNot(HaveOccurred())
			
			// Check if PVC was deleted
			var deletedPVC corev1.PersistentVolumeClaim
			err = fakeClient.Get(ctx, client.ObjectKey{
				Name:      "mongodb-backup-pvc",
				Namespace: "pr-123-shop-pilab-hu",
			}, &deletedPVC)
			Expect(err).To(HaveOccurred()) // Should not exist
		})

		It("should not fail if PVC doesn't exist", func() {
			prStack := &pishopv1alpha1.PRStack{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pr",
					Namespace: "default",
				},
				Spec: pishopv1alpha1.PRStackSpec{
					PRNumber: "123",
				},
			}
			
			err := reconciler.cleanupBackupPVC(ctx, prStack, "pr-123-shop-pilab-hu")
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("ensureBackupStorage", func() {
		It("should create backup PVC when backup is enabled", func() {
			prStack := &pishopv1alpha1.PRStack{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pr",
					Namespace: "default",
				},
				Spec: pishopv1alpha1.PRStackSpec{
					PRNumber: "123",
					BackupConfig: &pishopv1alpha1.BackupConfig{
						Enabled:      true,
						StorageSize:  "10Gi",
						StorageClass: "standard",
					},
				},
			}
			
			err := reconciler.ensureBackupStorage(ctx, prStack, "pr-123-shop-pilab-hu")
			Expect(err).ToNot(HaveOccurred())
			
			// Check if PVC was created
			var pvc corev1.PersistentVolumeClaim
			Expect(fakeClient.Get(ctx, client.ObjectKey{
				Name:      "mongodb-backup-pvc",
				Namespace: "pr-123-shop-pilab-hu",
			}, &pvc)).To(Succeed())
		})

		It("should skip PVC creation when backup is disabled", func() {
			prStack := &pishopv1alpha1.PRStack{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pr",
					Namespace: "default",
				},
				Spec: pishopv1alpha1.PRStackSpec{
					PRNumber: "123",
					BackupConfig: &pishopv1alpha1.BackupConfig{
						Enabled: false,
					},
				},
			}
			
			err := reconciler.ensureBackupStorage(ctx, prStack, "pr-123-shop-pilab-hu")
			Expect(err).ToNot(HaveOccurred())
			
			// Check if PVC was not created
			var pvc corev1.PersistentVolumeClaim
			err = fakeClient.Get(ctx, client.ObjectKey{
				Name:      "mongodb-backup-pvc",
				Namespace: "pr-123-shop-pilab-hu",
			}, &pvc)
			Expect(err).To(HaveOccurred()) // Should not exist
		})
	})
})
