package main

import (
	"flag"
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	pishopv1alpha1 "go.pilab.hu/shop/pishop-provisioner/api/v1alpha1"
	"go.pilab.hu/shop/pishop-provisioner/controllers"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")

	// Version information set during build
	Version   string
	Commit    string
	BuildDate string
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(pishopv1alpha1.AddToScheme(scheme))
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var mongoURI string
	var mongoUsername string
	var mongoPassword string

	var githubUsername string
	var githubToken string
	var githubEmail string

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&mongoURI, "mongo-uri", os.Getenv("MONGO_URI"), "MongoDB connection URI (required)")
	flag.StringVar(&mongoUsername, "mongo-username", getEnvOrDefault("MONGO_USERNAME", "admin"), "MongoDB admin username")
	flag.StringVar(&mongoPassword, "mongo-password", getEnvOrDefault("MONGO_PASSWORD", "password"), "MongoDB admin password")
	flag.StringVar(&githubUsername, "github-username", os.Getenv("GITHUB_USERNAME"), "GitHub username for container registry")
	flag.StringVar(&githubToken, "github-token", os.Getenv("GITHUB_TOKEN"), "GitHub token for container registry")
	flag.StringVar(&githubEmail, "github-email", os.Getenv("GITHUB_EMAIL"), "GitHub email for container registry")

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	if mongoURI == "" {
		setupLog.Error(fmt.Errorf("mongo-uri is required"), "unable to start manager")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsserver.Options{BindAddress: metricsAddr},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "pishop-provisioner.pilab.hu",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Initialize backup manager
	backupManager := &controllers.BackupRestoreManager{
		Client:        mgr.GetClient(),
		MongoURI:      mongoURI,
		MongoUsername: mongoUsername,
		MongoPassword: mongoPassword,
		BackupPath:    "/backups",
	}

	if err = (&controllers.PRStackReconciler{
		Client:         mgr.GetClient(),
		Scheme:         mgr.GetScheme(),
		Recorder:       mgr.GetEventRecorderFor("pishop-operator"),
		MongoURI:       mongoURI,
		MongoUsername:  mongoUsername,
		MongoPassword:  mongoPassword,
		BackupManager:  backupManager,
		GitHubUsername: githubUsername,
		GitHubToken:    githubToken,
		GitHubEmail:    githubEmail,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "PRStack")
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager",
		"version", Version,
		"commit", Commit,
		"buildDate", BuildDate)
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
