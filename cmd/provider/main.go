/*
Copyright 2025 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"os"
	"path/filepath"
	"runtime"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"

	xpcontroller "github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/feature"
	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	"github.com/crossplane/crossplane-runtime/v2/pkg/ratelimiter"

	"github.com/rossigee/provider-mailgun/apis"
	"github.com/rossigee/provider-mailgun/internal/controller"
	"github.com/rossigee/provider-mailgun/internal/features"
	"github.com/rossigee/provider-mailgun/internal/health"
	"github.com/rossigee/provider-mailgun/internal/version"
)

func main() {
	var (
		app                     = kingpin.New(filepath.Base(os.Args[0]), "Mailgun Crossplane provider").DefaultEnvars()
		debug                   = app.Flag("debug", "Run with debug logging.").Short('d').Bool()
		syncInterval            = app.Flag("sync", "Sync interval controls how often all resources will be double checked for drift.").Short('s').Default("1h").Duration()
		pollInterval            = app.Flag("poll", "Poll interval controls how often an individual resource should be checked for drift.").Default("1m").Duration()
		leaderElection          = app.Flag("leader-election", "Use leader election for the controller manager.").Short('l').Default("false").OverrideDefaultFromEnvar("LEADER_ELECTION").Bool()
		maxReconcileRate        = app.Flag("max-reconcile-rate", "The global maximum rate per second at which resources may checked for drift from the desired state.").Default("100").Int()
		enableManagementPolicies = app.Flag("enable-management-policies", "Enable support for Management Policies.").Default("true").Bool()
	)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	zl := zap.New(zap.UseDevMode(*debug))
	log := logging.NewLogrLogger(zl.WithName("provider-mailgun"))

	// Always set a logger for controller-runtime, but adjust verbosity
	if *debug {
		// In debug mode, use full verbosity
		ctrl.SetLogger(zl)
	} else {
		// In production mode, use a less verbose logger to avoid noise
		ctrl.SetLogger(zl.WithValues("source", "controller-runtime").V(1))
	}

	// Log startup information with build and configuration details
	log.Info("Provider starting up",
		"provider", "provider-mailgun",
		"version", version.Version,
		"go-version", runtime.Version(),
		"platform", runtime.GOOS+"/"+runtime.GOARCH,
		"sync-interval", syncInterval.String(),
		"poll-interval", pollInterval.String(),
		"max-reconcile-rate", *maxReconcileRate,
		"leader-election", *leaderElection,
		"management-policies", *enableManagementPolicies,
		"debug-mode", *debug)

	log.Debug("Detailed startup configuration",
		"sync-interval", syncInterval.String(),
		"poll-interval", pollInterval.String(),
		"max-reconcile-rate", *maxReconcileRate)

	cfg, err := ctrl.GetConfig()
	kingpin.FatalIfError(err, "Cannot get API server rest config")

	// Get the namespace to watch for resources
	namespace, err := getWatchNamespace()
	kingpin.FatalIfError(err, "Cannot get watch namespace")

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		LeaderElection:                *leaderElection,
		LeaderElectionID:              "crossplane-leader-election-provider-mailgun",
		LeaderElectionResourceLock:    resourcelock.LeasesResourceLock,
		Cache:                         cache.Options{DefaultNamespaces: map[string]cache.Config{namespace: {}}},
		LeaderElectionReleaseOnCancel: true,
		Metrics: server.Options{
			BindAddress: ":8080", // Single HTTP server for both metrics and health checks
		},
	})
	kingpin.FatalIfError(err, "Cannot create controller manager")

	// Add our APIs to the manager
	kingpin.FatalIfError(apis.AddToScheme(mgr.GetScheme()), "Cannot add Mailgun APIs to scheme")

	// Setup feature flags
	featureFlags := &feature.Flags{}
	if *enableManagementPolicies {
		featureFlags.Enable(features.EnableAlphaManagementPolicies)
	}

	// Setup rate limiter
	rateLimiter := ratelimiter.NewGlobal(*maxReconcileRate)

	// Setup controller options
	o := xpcontroller.Options{
		Logger:                  log,
		MaxConcurrentReconciles: *maxReconcileRate,
		PollInterval:            *pollInterval,
		GlobalRateLimiter:       rateLimiter,
		Features:                featureFlags,
	}

	// Setup all controllers
	kingpin.FatalIfError(controller.Setup(mgr, o), "Cannot setup Mailgun controllers")

	// Add health checks to the manager's built-in endpoints
	healthChecker := health.NewHealthChecker(mgr.GetClient(), nil)
	kingpin.FatalIfError(mgr.AddHealthzCheck("mailgun-provider", healthChecker.HealthzCheck), "Cannot add healthz check")
	kingpin.FatalIfError(mgr.AddReadyzCheck("mailgun-provider", healthChecker.ReadyzCheck), "Cannot add readyz check")

	log.Info("Starting manager")
	kingpin.FatalIfError(mgr.Start(ctrl.SetupSignalHandler()), "Cannot start controller manager")
}

// getWatchNamespace returns the namespace the operator should be watching for changes
func getWatchNamespace() (string, error) {
	ns, found := os.LookupEnv("WATCH_NAMESPACE")
	if !found {
		return "", nil
	}
	return ns, nil
}
