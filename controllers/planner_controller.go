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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
	types "github.com/miha3009/planner/controllers/types"
	
	informer "github.com/miha3009/planner/controllers/informer"
	rescheduler "github.com/miha3009/planner/controllers/rescheduler"
	resourceupdater "github.com/miha3009/planner/controllers/resourceupdater"
	executor "github.com/miha3009/planner/controllers/executor"
)

// PlannerReconciler reconciles a Planner object
type PlannerReconciler struct {
	Client client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Metrics *metricsv.Clientset
	Events chan types.Event
	Cache types.PlannerCache
	Context context.Context
	CancelFunc *context.CancelFunc
	LastStart time.Time
}

//+kubebuilder:rbac:groups=apps.hse.ru,resources=planners,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.hse.ru,resources=planners/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;

func (r *PlannerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("planner", req.NamespacedName)
	rand.Seed(time.Now().UnixNano())
	restart := ctrl.Result{RequeueAfter: time.Second}

	planner, err := r.GetPlanner(ctx, req)
	if err != nil || planner == nil {
		log.Error(err, ". Failed to get Planner")
		return restart, err
	}

	select {
		case e := <- r.Events:
			if r.ProcessEvent(ctx, planner, e) {
				r.UpdatePlanner(ctx, planner)
			}
		default:
			break
	}

	log.Info("Active: ", planner.Status.Active)
	if !planner.Status.Active {
		return restart, nil
	}
	
	if r.CancelFunc == nil {
		context, cancelFunc := context.WithCancel(ctx)
		r.Context = context
		r.CancelFunc = &cancelFunc		
	}
	
	if planner.Status.Phase == appsv1.Ready {
		nextStart := r.LastStart.Add(time.Second*time.Duration(planner.Spec.PlanningInterval))
		if nextStart.Before(time.Now()) {
			go informer.GetInfo(r.Context, r.Events, &r.Cache, r.Client, r.Metrics, planner.Spec)
			r.LastStart = time.Now()
			planner.Status.Phase = appsv1.Informing
			r.UpdatePlanner(ctx, planner)
		}
	}

	return restart, nil
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

func (r *PlannerReconciler) ProcessEvent(ctx context.Context, planner *appsv1.Planner, e types.Event) bool {
	switch e {
		case types.Start:
			if !planner.Status.Active || r.CancelFunc == nil {
				planner.Status.Active = true
				planner.Status.Phase = appsv1.Ready
				context, cancelFunc := context.WithCancel(ctx)
				r.Context = context
				r.CancelFunc = &cancelFunc
				log.Info("Planner started")
				return true
			}
		case types.Stop:
			if planner.Status.Active {
				planner.Status.Active = false
				planner.Status.Phase = appsv1.Ready
				if r.CancelFunc != nil {
					(*r.CancelFunc)()
					r.CancelFunc = nil
				}
				r.ClearCache()
				log.Info("Planner stopped")
				return true
			}
		case types.InformingEnded:
			go resourceupdater.UpdatePodResources(r.Context, r.Events, &r.Cache, planner.Spec)
			planner.Status.Phase = appsv1.ResourcesUpdating
			return true
		case types.ResourceUpdatingEnded:
			go rescheduler.GenPlan(r.Context, r.Events, &r.Cache, planner.Spec)
			planner.Status.Phase = appsv1.Planning
			return true
		case types.PlanningEnded:
			go executor.ExecutePlan(r.Context, r.Events, &r.Cache, planner.Spec)
			planner.Status.Phase = appsv1.Executing
			return true
		case types.ExecutingEnded:
			planner.Status.Phase = appsv1.Ready
			r.ClearCache()
			return true
		case types.PhaseEndedWithError:
			nextStart := r.LastStart.Add(time.Second*time.Duration(planner.Spec.PlanningInterval))
			log.Info("Error. Planner will restart after ", nextStart)
			planner.Status.Phase = appsv1.Ready
			r.ClearCache()
			return true
	}
	return false
}

func (r *PlannerReconciler) UpdatePlanner(ctx context.Context, planner *appsv1.Planner) {
	if err := r.Client.Status().Update(ctx, planner); err != nil {
		log.Error(err)
	}
}

func (r *PlannerReconciler) ClearCache() {
	r.Cache = types.PlannerCache{}
}

