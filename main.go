/*
Copyright 2022.

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
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"

	appsv1 "github.com/miha3009/planner/api/v1"
	"github.com/miha3009/planner/controllers"
	types "github.com/miha3009/planner/controllers/types"
	"github.com/prometheus/common/log"
	//+kubebuilder:scaffold:imports
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(appsv1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	log.Info("Creating the Manager.")
	config := ctrl.GetConfigOrDie()
	mgr, err := ctrl.NewManager(config, ctrl.Options{
		Scheme: scheme,
	})
	if err != nil {
		log.Error(err, "Unable to start manager")
		os.Exit(1)
	}

	clientset, err := metricsv.NewForConfig(config)
	if err != nil {
		log.Error(err, "Unable to get metrics client")
		os.Exit(1)
	}

	log.Info("Starting the Controller.")
	events := make(chan types.Event, 10)
	events <- types.Start
	reconciler := &controllers.PlannerReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Planner"),
		Scheme: mgr.GetScheme(),
		Metrics: clientset,
		Events: events,
		Cache: types.PlannerCache{},
		CancelFunc: nil,
		LastStart: time.Time{},
	}
	if err = reconciler.SetupWithManager(mgr); err != nil {
		log.Error(err, "unable to create controller", "controller", "Planner")
		os.Exit(1)
	}
	
	go controllers.RunServer(reconciler)

	log.Info("Starting the Manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Error(err, "problem running manager")
		os.Exit(1)
	}
}

