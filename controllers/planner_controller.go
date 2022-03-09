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

package controllers

import (
	"context"
	"time"
	"math/rand"

	appsv1 "github.com/miha3009/planner/api/v1"
	"github.com/go-logr/logr"
	"github.com/prometheus/common/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
//	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	
	info "github.com/miha3009/planner/controllers/informer"
	rescheduler "github.com/miha3009/planner/controllers/rescheduler"
	workload "github.com/miha3009/planner/controllers/workload"
	executor "github.com/miha3009/planner/controllers/executor"
)


// PlannerReconciler reconciles a Planner object
type PlannerReconciler struct {
	Client client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps.hse.ru,resources=planners,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.hse.ru,resources=planners/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;

func (r *PlannerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("planner", req.NamespacedName)
	rand.Seed(time.Now().UnixNano())
	restart := ctrl.Result{RequeueAfter: time.Second*5}

	planner, err := r.GetPlanner(ctx, req)
	if err != nil {
		log.Error(err, ". Failed to get Planner")
		return restart, err
	}
	
	if planner == nil || !planner.Spec.Active {
		return restart, nil
	}
	
	nodes, err := info.GetNodes(r.Client, ctx)
	if err != nil {
		log.Error(err, ". Failed to get nodes")
		return restart, err
	}
	log.Info("Found ", len(nodes), " nodes")

	pods, err := info.GetPods(planner, nodes, r.Client, ctx)
	if err != nil {
		log.Error(err, ". Failed to get pods")
		return restart, err
	}
	log.Info("Found ", len(pods), " pods")

	workload.UpdatePodResources(pods)

	plan := rescheduler.GenPlan(planner.Spec.Constraints, planner.Spec.Preferences, nodes, pods)
	log.Info("Plan length: ", len(plan.Movements))

	err = executor.ExecutePlan(plan)
	if err != nil {
		log.Error(err, "Failed to execute plan");
		return restart, err
	}

	return ctrl.Result{RequeueAfter: time.Second*time.Duration(planner.Spec.Delay)}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PlannerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &corev1.Pod{}, "spec.nodeName", func(rawObj client.Object) []string {
		pod := rawObj.(*corev1.Pod)
		return []string{pod.Spec.NodeName}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Planner{}).
		Complete(r)
}

func (r *PlannerReconciler) GetPlanner(ctx context.Context, req ctrl.Request) (*appsv1.Planner, error) {
	planner := &appsv1.Planner{}
	if err := r.Client.Get(ctx, req.NamespacedName, planner); err != nil {
		if errors.IsNotFound(err) {
			log.Info("Planner resource not found. Ignoring since object must be deleted")
			return nil, nil
		}
		return nil, err
	}
	return planner, nil
}

