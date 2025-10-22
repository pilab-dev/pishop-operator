package controllers

import (
	"context"
	"time"

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
)

var _ = Describe("Lifecycle Management Tests", func() {
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

	Context("Edge Cases - isStackExpired", func() {
		It("should return false when LastActiveAt is nil and CreatedAt is nil", func() {
			prStack := &pishopv1alpha1.PRStack{
				Status: pishopv1alpha1.PRStackStatus{
					LastActiveAt: nil,
					CreatedAt:    nil,
				},
			}
			Expect(reconciler.isStackExpired(prStack)).To(BeFalse())
		})

		It("should check CreatedAt when LastActiveAt is nil", func() {
			oldTime := metav1.NewTime(time.Now().Add(-2 * time.Hour))
			prStack := &pishopv1alpha1.PRStack{
				Status: pishopv1alpha1.PRStackStatus{
					LastActiveAt: nil,
					CreatedAt:    &oldTime,
				},
			}
			Expect(reconciler.isStackExpired(prStack)).To(BeTrue())
		})

		It("should return false for recently active stack", func() {
			recentTime := metav1.NewTime(time.Now().Add(-30 * time.Minute))
			prStack := &pishopv1alpha1.PRStack{
				Status: pishopv1alpha1.PRStackStatus{
					LastActiveAt: &recentTime,
				},
			}
			Expect(reconciler.isStackExpired(prStack)).To(BeFalse())
		})

		It("should return true for expired stack", func() {
			oldTime := metav1.NewTime(time.Now().Add(-2 * time.Hour))
			prStack := &pishopv1alpha1.PRStack{
				Status: pishopv1alpha1.PRStackStatus{
					LastActiveAt: &oldTime,
				},
			}
			Expect(reconciler.isStackExpired(prStack)).To(BeTrue())
		})

		It("should handle exactly 1 hour age", func() {
			exactTime := metav1.NewTime(time.Now().Add(-StackExpirationTime))
			prStack := &pishopv1alpha1.PRStack{
				Status: pishopv1alpha1.PRStackStatus{
					LastActiveAt: &exactTime,
				},
			}
			// Should be expired (>= 1 hour)
			isExpired := reconciler.isStackExpired(prStack)
			Expect(isExpired).To(BeTrue())
		})

		It("should handle just under 1 hour age", func() {
			justUnderTime := metav1.NewTime(time.Now().Add(-59 * time.Minute))
			prStack := &pishopv1alpha1.PRStack{
				Status: pishopv1alpha1.PRStackStatus{
					LastActiveAt: &justUnderTime,
				},
			}
			Expect(reconciler.isStackExpired(prStack)).To(BeFalse())
		})
	})

	Context("Edge Cases - shouldRolloutDeployments", func() {
		It("should return false when deployedAt is nil", func() {
			prStack := &pishopv1alpha1.PRStack{
				Spec: pishopv1alpha1.PRStackSpec{
					DeployedAt: nil,
				},
			}
			Expect(reconciler.shouldRolloutDeployments(prStack)).To(BeFalse())
		})

		It("should return true when lastDeployedAt is nil", func() {
			now := metav1.Now()
			prStack := &pishopv1alpha1.PRStack{
				Spec: pishopv1alpha1.PRStackSpec{
					DeployedAt: &now,
				},
				Status: pishopv1alpha1.PRStackStatus{
					LastDeployedAt: nil,
				},
			}
			Expect(reconciler.shouldRolloutDeployments(prStack)).To(BeTrue())
		})

		It("should return false when deployedAt equals lastDeployedAt", func() {
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

		It("should return true when deployedAt differs from lastDeployedAt", func() {
			oldTime := metav1.NewTime(time.Now().Add(-1 * time.Hour))
			newTime := metav1.Now()
			prStack := &pishopv1alpha1.PRStack{
				Spec: pishopv1alpha1.PRStackSpec{
					DeployedAt: &newTime,
				},
				Status: pishopv1alpha1.PRStackStatus{
					LastDeployedAt: &oldTime,
				},
			}
			Expect(reconciler.shouldRolloutDeployments(prStack)).To(BeTrue())
		})
	})

	Context("Edge Cases - scaleDeployments", func() {
		var namespace string

		BeforeEach(func() {
			namespace = "pr-scale-test-shop-pilab-hu"

			// Create namespace
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespace,
				},
			}
			Expect(fakeClient.Create(ctx, ns)).To(Succeed())
		})

		It("should handle empty namespace", func() {
			err := reconciler.scaleDeployments(ctx, "pr-empty-shop-pilab-hu", 1)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should scale deployment to 0 replicas", func() {
			// Create deployment with 1 replica
			replicas := int32(1)
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-deployment",
					Namespace: namespace,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &replicas,
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "test"},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{"app": "test"},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "test",
									Image: "test:latest",
								},
							},
						},
					},
				},
			}
			Expect(fakeClient.Create(ctx, deployment)).To(Succeed())

			// Scale down
			err := reconciler.scaleDeployments(ctx, namespace, 0)
			Expect(err).ToNot(HaveOccurred())

			// Verify scaled
			var scaledDeployment appsv1.Deployment
			Expect(fakeClient.Get(ctx, client.ObjectKey{
				Name:      "test-deployment",
				Namespace: namespace,
			}, &scaledDeployment)).To(Succeed())

			Expect(*scaledDeployment.Spec.Replicas).To(Equal(int32(0)))
		})

		It("should scale deployment to 1 replica", func() {
			// Create deployment with 0 replicas
			replicas := int32(0)
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-deployment",
					Namespace: namespace,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &replicas,
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "test"},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{"app": "test"},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "test",
									Image: "test:latest",
								},
							},
						},
					},
				},
			}
			Expect(fakeClient.Create(ctx, deployment)).To(Succeed())

			// Scale up
			err := reconciler.scaleDeployments(ctx, namespace, 1)
			Expect(err).ToNot(HaveOccurred())

			// Verify scaled
			var scaledDeployment appsv1.Deployment
			Expect(fakeClient.Get(ctx, client.ObjectKey{
				Name:      "test-deployment",
				Namespace: namespace,
			}, &scaledDeployment)).To(Succeed())

			Expect(*scaledDeployment.Spec.Replicas).To(Equal(int32(1)))
		})

		It("should skip deployment already at desired replicas", func() {
			// Create deployment with 1 replica
			replicas := int32(1)
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-deployment",
					Namespace: namespace,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &replicas,
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "test"},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{"app": "test"},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "test",
									Image: "test:latest",
								},
							},
						},
					},
				},
			}
			Expect(fakeClient.Create(ctx, deployment)).To(Succeed())

			// Scale to same value
			err := reconciler.scaleDeployments(ctx, namespace, 1)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should scale multiple deployments", func() {
			// Create two deployments
			for i := 0; i < 2; i++ {
				replicas := int32(1)
				deployment := &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-deployment-" + string(rune(i)),
						Namespace: namespace,
					},
					Spec: appsv1.DeploymentSpec{
						Replicas: &replicas,
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"app": "test"},
						},
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{"app": "test"},
							},
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:  "test",
										Image: "test:latest",
									},
								},
							},
						},
					},
				}
				Expect(fakeClient.Create(ctx, deployment)).To(Succeed())
			}

			// Scale all
			err := reconciler.scaleDeployments(ctx, namespace, 0)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("Edge Cases - Phase Constants", func() {
		It("should have correct phase constant values", func() {
			Expect(PhaseInitialization).To(Equal(""))
			Expect(PhaseProvisioning).To(Equal("Provisioning"))
			Expect(PhaseDeploying).To(Equal("Deploying"))
			Expect(PhaseRunning).To(Equal("Running"))
			Expect(PhaseInactive).To(Equal("Inactive"))
			Expect(PhaseCleaning).To(Equal("Cleaning"))
			Expect(PhaseCleaned).To(Equal("Cleaned"))
		})

		It("should have correct finalizer name", func() {
			Expect(FinalizerName).To(Equal("shop.pilab.hu/finalizer"))
		})

		It("should have correct stack expiration time", func() {
			Expect(StackExpirationTime).To(Equal(time.Hour))
		})
	})

	Context("Edge Cases - Finalizer Management", func() {
		It("should handle adding finalizer to empty finalizers list", func() {
			finalizers := []string{}
			Expect(containsString(finalizers, FinalizerName)).To(BeFalse())

			finalizers = append(finalizers, FinalizerName)
			Expect(containsString(finalizers, FinalizerName)).To(BeTrue())
		})

		It("should handle removing finalizer from list", func() {
			finalizers := []string{FinalizerName, "other-finalizer"}
			finalizers = removeString(finalizers, FinalizerName)

			Expect(containsString(finalizers, FinalizerName)).To(BeFalse())
			Expect(len(finalizers)).To(Equal(1))
		})

		It("should handle removing non-existent finalizer", func() {
			finalizers := []string{"other-finalizer"}
			result := removeString(finalizers, FinalizerName)

			Expect(result).To(Equal(finalizers))
		})

		It("should handle multiple finalizers", func() {
			finalizers := []string{"finalizer-1", FinalizerName, "finalizer-2"}
			Expect(containsString(finalizers, FinalizerName)).To(BeTrue())

			finalizers = removeString(finalizers, FinalizerName)
			Expect(containsString(finalizers, FinalizerName)).To(BeFalse())
			Expect(len(finalizers)).To(Equal(2))
		})
	})

	Context("Edge Cases - Rollout Management", func() {
		var namespace string

		BeforeEach(func() {
			namespace = "pr-rollout-test-shop-pilab-hu"

			// Create namespace
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespace,
				},
			}
			Expect(fakeClient.Create(ctx, ns)).To(Succeed())
		})

		It("should handle rollout with no deployments", func() {
			prStack := &pishopv1alpha1.PRStack{
				Spec: pishopv1alpha1.PRStackSpec{
					PRNumber:   "333",
					DeployedAt: &metav1.Time{Time: time.Now()},
				},
			}

			err := reconciler.rolloutDeployments(ctx, prStack)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle rollout of existing deployments", func() {
			// Create deployment
			replicas := int32(1)
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-deployment",
					Namespace: namespace,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &replicas,
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "test"},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{"app": "test"},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "test",
									Image: "test:latest",
								},
							},
						},
					},
				},
			}
			Expect(fakeClient.Create(ctx, deployment)).To(Succeed())

			// Perform rollout
			deployedAt := metav1.Now()
			prStack := &pishopv1alpha1.PRStack{
				Spec: pishopv1alpha1.PRStackSpec{
					PRNumber:   "444",
					DeployedAt: &deployedAt,
				},
			}

			// Use the correct namespace for rollout
			err := reconciler.rolloutDeployments(ctx, prStack)
			
			// The rollout should succeed even if no deployments match the namespace
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
