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

func TestProvisioning(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Provisioning Suite")
}

var _ = Describe("Provisioning", func() {
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

	Context("createMongoDBSecret", func() {
		It("should create MongoDB secret successfully", func() {
			prStack := &pishopv1alpha1.PRStack{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pr",
					Namespace: "default",
				},
				Spec: pishopv1alpha1.PRStackSpec{
					PRNumber: "123",
				},
				Status: pishopv1alpha1.PRStackStatus{
					MongoDB: &pishopv1alpha1.MongoDBCredentials{
						User:             "test-user",
						Password:         "test-password",
						ConnectionString: "mongodb://test:27017",
						Databases:        []string{"test_db1", "test_db2"},
					},
				},
			}
			
			err := reconciler.createMongoDBSecret(ctx, prStack)
			Expect(err).ToNot(HaveOccurred())
			
			// Check if secret was created
			var secret corev1.Secret
			Expect(fakeClient.Get(ctx, client.ObjectKey{
				Name:      MongoDBSecretName,
				Namespace: "pr-123-shop-pilab-hu",
			}, &secret)).To(Succeed())
			
			Expect(secret.StringData["username"]).To(Equal("test-user"))
			Expect(secret.StringData["password"]).To(Equal("test-password"))
			Expect(secret.StringData["connectionString"]).To(Equal("mongodb://test:27017"))
			Expect(secret.StringData["databases"]).To(Equal("test_db1,test_db2"))
		})

		It("should fail when MongoDB status is nil", func() {
			prStack := &pishopv1alpha1.PRStack{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pr",
					Namespace: "default",
				},
				Spec: pishopv1alpha1.PRStackSpec{
					PRNumber: "123",
				},
				Status: pishopv1alpha1.PRStackStatus{
					MongoDB: nil,
				},
			}
			
			err := reconciler.createMongoDBSecret(ctx, prStack)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("MongoDB credentials not available"))
		})
	})

	Context("createNamespace", func() {
		It("should create namespace if it doesn't exist", func() {
			namespaceName := "pr-123-shop-pilab-hu"
			
			err := reconciler.createNamespace(ctx, namespaceName)
			Expect(err).ToNot(HaveOccurred())
			
			// Check if namespace was created
			var namespace corev1.Namespace
			Expect(fakeClient.Get(ctx, client.ObjectKey{Name: namespaceName}, &namespace)).To(Succeed())
			Expect(namespace.Name).To(Equal(namespaceName))
		})

		It("should not fail if namespace already exists", func() {
			namespaceName := "pr-123-shop-pilab-hu"
			
			// Create namespace first
			namespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespaceName,
				},
			}
			Expect(fakeClient.Create(ctx, namespace)).To(Succeed())
			
			// Try to create again
			err := reconciler.createNamespace(ctx, namespaceName)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("getResourceRequirements", func() {
		It("should return default resource requirements when no limits specified", func() {
			prStack := &pishopv1alpha1.PRStack{
				Spec: pishopv1alpha1.PRStackSpec{
					ResourceLimits: nil,
				},
			}
			
			requirements := reconciler.getResourceRequirements(prStack)
			
			Expect(requirements.Limits.Cpu().String()).To(Equal("500m"))
			Expect(requirements.Limits.Memory().String()).To(Equal("512Mi"))
			Expect(requirements.Requests.Cpu().String()).To(Equal("0"))
			Expect(requirements.Requests.Memory().String()).To(Equal("0"))
		})

		It("should return custom resource requirements when limits specified", func() {
			prStack := &pishopv1alpha1.PRStack{
				Spec: pishopv1alpha1.PRStackSpec{
					ResourceLimits: &pishopv1alpha1.ResourceLimits{
						CPULimit:    "1000m",
						MemoryLimit: "1Gi",
					},
				},
			}
			
			requirements := reconciler.getResourceRequirements(prStack)
			
			Expect(requirements.Limits.Cpu().String()).To(Equal("1"))
			Expect(requirements.Limits.Memory().String()).To(Equal("1Gi"))
			Expect(requirements.Requests.Cpu().String()).To(Equal("0"))
			Expect(requirements.Requests.Memory().String()).To(Equal("0"))
		})
	})

	Context("getImageTag", func() {
		It("should return custom image tag when specified", func() {
			prStack := &pishopv1alpha1.PRStack{
				Spec: pishopv1alpha1.PRStackSpec{
					PRNumber: "123",
					ImageTag: "v1.2.3",
				},
			}
			
			imageTag := getImageTag(prStack, "product-service")
			Expect(imageTag).To(Equal("ghcr.io/pilab-dev/shop-product-service:v1.2.3"))
		})

		It("should return default image tag when not specified", func() {
			prStack := &pishopv1alpha1.PRStack{
				Spec: pishopv1alpha1.PRStackSpec{
					PRNumber: "123",
				},
			}
			
			imageTag := getImageTag(prStack, "product-service")
			Expect(imageTag).To(Equal("ghcr.io/pilab-dev/shop-product-service:pr-123"))
		})
	})

	Context("CreateOrUpdate", func() {
		It("should create resource if it doesn't exist", func() {
			configMap := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-config",
					Namespace: "default",
				},
				Data: map[string]string{
					"key": "value",
				},
			}
			
			err := reconciler.CreateOrUpdate(ctx, configMap)
			Expect(err).ToNot(HaveOccurred())
			
			// Check if configmap was created
			var createdConfigMap corev1.ConfigMap
			Expect(fakeClient.Get(ctx, client.ObjectKey{
				Name:      "test-config",
				Namespace: "default",
			}, &createdConfigMap)).To(Succeed())
		})

		It("should update resource if it already exists", func() {
			configMap := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-config",
					Namespace: "default",
				},
				Data: map[string]string{
					"key": "value",
				},
			}
			
			// Create first
			Expect(fakeClient.Create(ctx, configMap)).To(Succeed())
			
			// Create a new configmap with updated data (simulating what CreateOrUpdate would do)
			updatedConfigMap := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-config",
					Namespace: "default",
				},
				Data: map[string]string{
					"key": "updated-value",
				},
			}
			
			err := reconciler.CreateOrUpdate(ctx, updatedConfigMap)
			Expect(err).ToNot(HaveOccurred())
			
			// Check if configmap was updated
			var finalConfigMap corev1.ConfigMap
			Expect(fakeClient.Get(ctx, client.ObjectKey{
				Name:      "test-config",
				Namespace: "default",
			}, &finalConfigMap)).To(Succeed())
			Expect(finalConfigMap.Data["key"]).To(Equal("updated-value"))
		})
	})
})
