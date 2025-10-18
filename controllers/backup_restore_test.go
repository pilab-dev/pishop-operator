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
	batchv1 "k8s.io/api/batch/v1"
)

func TestBackupRestore(t *testing.T) {
	// Test suite is now managed by suite_test.go
}

var _ = Describe("Backup Restore Manager", func() {
	var (
		ctx           context.Context
		cancel        context.CancelFunc
		backupManager *BackupRestoreManager
		fakeClient    client.Client
		scheme        *runtime.Scheme
	)

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(context.Background())
		
		scheme = runtime.NewScheme()
		Expect(pishopv1alpha1.AddToScheme(scheme)).To(Succeed())
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(batchv1.AddToScheme(scheme)).To(Succeed())
		
		fakeClient = fake.NewClientBuilder().
			WithScheme(scheme).
			Build()
		
		backupManager = &BackupRestoreManager{
			Client:        fakeClient,
			MongoURI:      "mongodb://localhost:27017",
			MongoUsername: "admin",
			MongoPassword: "password",
			BackupPath:    "/backups",
		}
		
		ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	})

	AfterEach(func() {
		cancel()
	})

	Context("CreateBackup", func() {
		It("should create backup job successfully", func() {
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
			
			err := backupManager.CreateBackup(ctx, prStack)
			Expect(err).ToNot(HaveOccurred())
			
			// Check if backup job was created
			var jobs batchv1.JobList
			Expect(fakeClient.List(ctx, &jobs)).To(Succeed())
			Expect(jobs.Items).To(HaveLen(1))
			Expect(jobs.Items[0].Name).To(ContainSubstring("backup-pr-123-"))
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
			
			err := backupManager.CreateBackup(ctx, prStack)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("RestoreBackup", func() {
		It("should create restore job successfully", func() {
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
			
			backupName := "pr-123-20240101-120000"
			err := backupManager.RestoreBackup(ctx, prStack, backupName)
			Expect(err).ToNot(HaveOccurred())
			
			// Check if restore job was created
			var jobs batchv1.JobList
			Expect(fakeClient.List(ctx, &jobs)).To(Succeed())
			Expect(jobs.Items).To(HaveLen(1))
			Expect(jobs.Items[0].Name).To(Equal("restore-" + backupName))
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
			
			err := backupManager.RestoreBackup(ctx, prStack, "test-backup")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("ListBackups", func() {
		It("should return placeholder backups", func() {
			backups, err := backupManager.ListBackups(ctx, "123")
			Expect(err).ToNot(HaveOccurred())
			Expect(backups).To(HaveLen(2))
			Expect(backups[0]).To(Equal("pr-123-20240101-120000"))
			Expect(backups[1]).To(Equal("pr-123-20240102-120000"))
		})
	})

	Context("CleanupOldBackups", func() {
		It("should complete without error", func() {
			err := backupManager.CleanupOldBackups(ctx, "123", 30)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("Helper Functions", func() {
		It("should generate correct backup script", func() {
			spec := &BackupSpec{
				PRNumber:    "123",
				Databases:   []string{"test_db1", "test_db2"},
				BackupName:  "test-backup",
				Compression: true,
			}
			
			script := backupManager.generateBackupScript(spec)
			Expect(script).To(ContainSubstring("PR_NUMBER"))
			Expect(script).To(ContainSubstring("BACKUP_NAME"))
			Expect(script).To(ContainSubstring("test_db1"))
			Expect(script).To(ContainSubstring("test_db2"))
			Expect(script).To(ContainSubstring("mongodump"))
			Expect(script).To(ContainSubstring("--gzip"))
		})

		It("should generate correct restore script", func() {
			spec := &RestoreSpec{
				PRNumber:   "123",
				BackupName: "test-backup",
				Databases:  []string{"test_db1", "test_db2"},
			}
			
			script := backupManager.generateRestoreScript(spec)
			Expect(script).To(ContainSubstring("PR_NUMBER"))
			Expect(script).To(ContainSubstring("BACKUP_NAME"))
			Expect(script).To(ContainSubstring("test_db1"))
			Expect(script).To(ContainSubstring("test_db2"))
			Expect(script).To(ContainSubstring("mongorestore"))
			Expect(script).To(ContainSubstring("--gzip"))
		})
	})
})
