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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

var _ = Describe("Service Deployment Tests", func() {
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
		Expect(networkingv1.AddToScheme(scheme)).To(Succeed())
		
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
			GitHubUsername: "testuser",
			GitHubToken:    "testtoken",
			GitHubEmail:    "test@example.com",
		}
		
		ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	})

	AfterEach(func() {
		cancel()
	})

	Context("Edge Cases - createServiceDeployment", func() {
		var prStack *pishopv1alpha1.PRStack
		var namespace string

		BeforeEach(func() {
			namespace = "pr-123-shop-pilab-hu"
			
			// Create namespace
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespace,
				},
			}
			Expect(fakeClient.Create(ctx, ns)).To(Succeed())

			prStack = &pishopv1alpha1.PRStack{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pr-123",
					Namespace: "default",
				},
				Spec: pishopv1alpha1.PRStackSpec{
					PRNumber: "123",
					Active:   true,
				},
				Status: pishopv1alpha1.PRStackStatus{
					MongoDB: &pishopv1alpha1.MongoDBCredentials{
						User:             "test-user",
						Password:         "test-password",
						ConnectionString: "mongodb://test:27017",
						Databases:        []string{"test_db"},
					},
					NATS: &pishopv1alpha1.NATSConfig{
						ConnectionString: "nats://nats:4222",
						SubjectPrefix:    "pishop.pr.123",
					},
					Redis: &pishopv1alpha1.RedisConfig{
						ConnectionString: "redis://redis:6379",
						KeyPrefix:        "pishop:pr:123:",
					},
				},
			}
		})

		It("should create deployment with active replicas", func() {
			err := reconciler.createServiceDeployment(ctx, prStack, namespace, "product-service")
			Expect(err).ToNot(HaveOccurred())

			var deployment appsv1.Deployment
			Expect(fakeClient.Get(ctx, client.ObjectKey{
				Name:      "product-service",
				Namespace: namespace,
			}, &deployment)).To(Succeed())

			Expect(*deployment.Spec.Replicas).To(Equal(int32(1)))
		})

		It("should create deployment with zero replicas when inactive", func() {
			prStack.Spec.Active = false

			err := reconciler.createServiceDeployment(ctx, prStack, namespace, "product-service")
			Expect(err).ToNot(HaveOccurred())

			var deployment appsv1.Deployment
			Expect(fakeClient.Get(ctx, client.ObjectKey{
				Name:      "product-service",
				Namespace: namespace,
			}, &deployment)).To(Succeed())

			Expect(*deployment.Spec.Replicas).To(Equal(int32(0)))
		})

		It("should create service with correct labels", func() {
			err := reconciler.createServiceDeployment(ctx, prStack, namespace, "product-service")
			Expect(err).ToNot(HaveOccurred())

			var service corev1.Service
			Expect(fakeClient.Get(ctx, client.ObjectKey{
				Name:      "product-service",
				Namespace: namespace,
			}, &service)).To(Succeed())

			Expect(service.Labels["app"]).To(Equal("product-service"))
			Expect(service.Labels["pr-number"]).To(Equal("123"))
		})

		It("should create ingress only for graphql-service", func() {
			err := reconciler.createServiceDeployment(ctx, prStack, namespace, "graphql-service")
			Expect(err).ToNot(HaveOccurred())

			var ingress networkingv1.Ingress
			err = fakeClient.Get(ctx, client.ObjectKey{
				Name:      "graphql-service",
				Namespace: namespace,
			}, &ingress)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should not create ingress for non-graphql services", func() {
			err := reconciler.createServiceDeployment(ctx, prStack, namespace, "product-service")
			Expect(err).ToNot(HaveOccurred())

			var ingress networkingv1.Ingress
			err = fakeClient.Get(ctx, client.ObjectKey{
				Name:      "product-service",
				Namespace: namespace,
			}, &ingress)
			Expect(err).To(HaveOccurred())
		})

		It("should handle deployment with custom resource limits", func() {
			prStack.Spec.ResourceLimits = &pishopv1alpha1.ResourceLimits{
				CPULimit:    "2",
				MemoryLimit: "2Gi",
			}

			err := reconciler.createServiceDeployment(ctx, prStack, namespace, "product-service")
			Expect(err).ToNot(HaveOccurred())

			var deployment appsv1.Deployment
			Expect(fakeClient.Get(ctx, client.ObjectKey{
				Name:      "product-service",
				Namespace: namespace,
			}, &deployment)).To(Succeed())

			resources := deployment.Spec.Template.Spec.Containers[0].Resources
			Expect(resources.Limits.Cpu().String()).To(Equal("2"))
			Expect(resources.Limits.Memory().String()).To(Equal("2Gi"))
		})

		It("should handle service with empty name", func() {
			err := reconciler.createServiceDeployment(ctx, prStack, namespace, "")
			// Should not panic, may fail gracefully
			// Just verify it doesn't crash
			_ = err
		})

		It("should update existing deployment", func() {
			// Create first time
			err := reconciler.createServiceDeployment(ctx, prStack, namespace, "product-service")
			Expect(err).ToNot(HaveOccurred())

			// Update
			prStack.Spec.ImageTag = "v2.0.0"
			err = reconciler.createServiceDeployment(ctx, prStack, namespace, "product-service")
			Expect(err).ToNot(HaveOccurred())

			var deployment appsv1.Deployment
			Expect(fakeClient.Get(ctx, client.ObjectKey{
				Name:      "product-service",
				Namespace: namespace,
			}, &deployment)).To(Succeed())

			Expect(deployment.Spec.Template.Spec.Containers[0].Image).To(ContainSubstring("v2.0.0"))
		})
	})

	Context("Edge Cases - createIngress", func() {
		var prStack *pishopv1alpha1.PRStack
		var namespace string

		BeforeEach(func() {
			namespace = "pr-456-shop-pilab-hu"
			prStack = &pishopv1alpha1.PRStack{
				Spec: pishopv1alpha1.PRStackSpec{
					PRNumber: "456",
				},
			}
		})

		It("should create ingress with default domain", func() {
			ingress := reconciler.createIngress(prStack, namespace, "graphql-service", "graphql")

			Expect(ingress.Spec.Rules).To(HaveLen(1))
			Expect(ingress.Spec.Rules[0].Host).To(Equal("pr-456.shop.pilab.hu"))
		})

		It("should create ingress with custom domain", func() {
			prStack.Spec.CustomDomain = "custom.example.com"
			ingress := reconciler.createIngress(prStack, namespace, "graphql-service", "graphql")

			Expect(ingress.Spec.Rules[0].Host).To(Equal("custom.example.com"))
		})

		It("should include TLS configuration when secret specified", func() {
			prStack.Spec.IngressTlsSecretName = "tls-secret"
			ingress := reconciler.createIngress(prStack, namespace, "graphql-service", "graphql")

			Expect(ingress.Spec.TLS).To(HaveLen(1))
			Expect(ingress.Spec.TLS[0].SecretName).To(Equal("tls-secret"))
		})

		It("should not include TLS when secret not specified", func() {
			ingress := reconciler.createIngress(prStack, namespace, "graphql-service", "graphql")

			Expect(ingress.Spec.TLS).To(BeNil())
		})

		It("should handle empty path prefix", func() {
			ingress := reconciler.createIngress(prStack, namespace, "graphql-service", "")

			Expect(ingress.Spec.Rules[0].HTTP.Paths[0].Path).To(Equal("/"))
		})

		It("should set correct backend service", func() {
			ingress := reconciler.createIngress(prStack, namespace, "graphql-service", "graphql")

			backend := ingress.Spec.Rules[0].HTTP.Paths[0].Backend
			Expect(backend.Service.Name).To(Equal("graphql-service"))
			Expect(backend.Service.Port.Number).To(Equal(int32(8080)))
		})
	})

	Context("Edge Cases - createRegistrySecret", func() {
		var namespace string

		BeforeEach(func() {
			namespace = "pr-789-shop-pilab-hu"
			
			// Create namespace
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespace,
				},
			}
			Expect(fakeClient.Create(ctx, ns)).To(Succeed())
		})

		It("should create registry secret with GitHub credentials", func() {
			err := reconciler.createRegistrySecret(ctx, namespace)
			Expect(err).ToNot(HaveOccurred())

			var secret corev1.Secret
			Expect(fakeClient.Get(ctx, client.ObjectKey{
				Name:      "ghcr-secret",
				Namespace: namespace,
			}, &secret)).To(Succeed())

			Expect(secret.Type).To(Equal(corev1.SecretTypeDockerConfigJson))
			Expect(secret.Data).To(HaveKey(corev1.DockerConfigJsonKey))
		})

		It("should skip creation when GitHub credentials not configured", func() {
			reconciler.GitHubUsername = ""
			reconciler.GitHubToken = ""

			err := reconciler.createRegistrySecret(ctx, namespace)
			Expect(err).ToNot(HaveOccurred())

			var secret corev1.Secret
			err = fakeClient.Get(ctx, client.ObjectKey{
				Name:      "ghcr-secret",
				Namespace: namespace,
			}, &secret)
			Expect(err).To(HaveOccurred())
		})

		It("should handle empty GitHub username", func() {
			reconciler.GitHubUsername = ""

			err := reconciler.createRegistrySecret(ctx, namespace)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle empty GitHub token", func() {
			reconciler.GitHubToken = ""

			err := reconciler.createRegistrySecret(ctx, namespace)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle special characters in credentials", func() {
			reconciler.GitHubUsername = "user@domain"
			reconciler.GitHubToken = "token!@#$%"

			err := reconciler.createRegistrySecret(ctx, namespace)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("Edge Cases - createMongoDBResources", func() {
		var prStack *pishopv1alpha1.PRStack
		var namespace string

		BeforeEach(func() {
			namespace = "pr-999-shop-pilab-hu"
			
			// Create namespace
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespace,
				},
			}
			Expect(fakeClient.Create(ctx, ns)).To(Succeed())

			prStack = &pishopv1alpha1.PRStack{
				Spec: pishopv1alpha1.PRStackSpec{
					PRNumber: "999",
				},
				Status: pishopv1alpha1.PRStackStatus{
					MongoDB: &pishopv1alpha1.MongoDBCredentials{
						User:             "test-user",
						Password:         "test-password",
						ConnectionString: "mongodb://test:27017",
						Databases:        []string{"test_db"},
					},
					NATS: &pishopv1alpha1.NATSConfig{
						ConnectionString: "nats://nats:4222",
					},
					Redis: &pishopv1alpha1.RedisConfig{
						ConnectionString: "redis://redis:6379",
					},
				},
			}
		})

		It("should create MongoDB ConfigMap", func() {
			err := reconciler.createMongoDBResources(ctx, prStack, namespace)
			Expect(err).ToNot(HaveOccurred())

			var configMap corev1.ConfigMap
			Expect(fakeClient.Get(ctx, client.ObjectKey{
				Name:      "mongodb-config",
				Namespace: namespace,
			}, &configMap)).To(Succeed())

			Expect(configMap.Data).To(HaveKey("uri"))
		})

		It("should create MongoDB Secret", func() {
			err := reconciler.createMongoDBResources(ctx, prStack, namespace)
			Expect(err).ToNot(HaveOccurred())

			var secret corev1.Secret
			Expect(fakeClient.Get(ctx, client.ObjectKey{
				Name:      "mongodb-secret",
				Namespace: namespace,
			}, &secret)).To(Succeed())

			Expect(secret.StringData).To(HaveKey("username"))
			Expect(secret.StringData).To(HaveKey("password"))
		})

		It("should create NATS ConfigMap", func() {
			err := reconciler.createMongoDBResources(ctx, prStack, namespace)
			Expect(err).ToNot(HaveOccurred())

			var configMap corev1.ConfigMap
			Expect(fakeClient.Get(ctx, client.ObjectKey{
				Name:      "nats-config",
				Namespace: namespace,
			}, &configMap)).To(Succeed())

			Expect(configMap.Data).To(HaveKey("url"))
		})

		It("should create Redis ConfigMap", func() {
			err := reconciler.createMongoDBResources(ctx, prStack, namespace)
			Expect(err).ToNot(HaveOccurred())

			var configMap corev1.ConfigMap
			Expect(fakeClient.Get(ctx, client.ObjectKey{
				Name:      "redis-config",
				Namespace: namespace,
			}, &configMap)).To(Succeed())

			Expect(configMap.Data).To(HaveKey("url"))
		})

		It("should handle empty MongoDB credentials", func() {
			prStack.Status.MongoDB.User = ""
			prStack.Status.MongoDB.Password = ""

			err := reconciler.createMongoDBResources(ctx, prStack, namespace)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should update existing resources", func() {
			// Create first time
			err := reconciler.createMongoDBResources(ctx, prStack, namespace)
			Expect(err).ToNot(HaveOccurred())

			// Update
			prStack.Status.MongoDB.ConnectionString = "mongodb://updated:27017"
			err = reconciler.createMongoDBResources(ctx, prStack, namespace)
			Expect(err).ToNot(HaveOccurred())

			var configMap corev1.ConfigMap
			Expect(fakeClient.Get(ctx, client.ObjectKey{
				Name:      "mongodb-config",
				Namespace: namespace,
			}, &configMap)).To(Succeed())

			Expect(configMap.Data["uri"]).To(ContainSubstring("updated"))
		})
	})
})

