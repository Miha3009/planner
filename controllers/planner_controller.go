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
    "math/rand"
    "time"

    "github.com/go-logr/logr"
    appsv1 "github.com/miha3009/planner/api/v1"
    "github.com/prometheus/common/log"
    corev1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/runtime"
    clientset "k8s.io/client-go/kubernetes"
    metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"

    executor "github.com/miha3009/planner/controllers/executor"
    informer "github.com/miha3009/planner/controllers/informer"
    rescheduler "github.com/miha3009/planner/controllers/rescheduler"
    resourceupdater "github.com/miha3009/planner/controllers/resourceupdater"
    types "github.com/miha3009/planner/controllers/types"
)

type Process struct {
    Context    context.Context
    CancelFunc context.CancelFunc
}

// PlannerReconciler reconciles a Planner object
type PlannerReconciler struct {
    Client         client.Client
    Clientset      *clientset.Clientset
    Log            logr.Logger
    Scheme         *runtime.Scheme
    MetricsClient  *metricsv.Clientset
    Events         chan types.Event
    Cache          *types.PlannerCache
    MainProcess    *Process
    MetricsProcess *Process
    LastStart      time.Time
    Informer       informer.Informer // for testing purpose
    Executor       executor.Executor // for testing purpose
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

    planner, err := r.Informer.GetPlanner(ctx, r.Client, req)
    if err != nil || planner == nil {
        log.Error(err, ". Failed to get Planner")
        return restart, err
    }

    select {
    case e := <-r.Events:
        if r.ProcessEvent(ctx, planner, e) {
            r.Informer.UpdatePlanner(ctx, r.Client, planner)
        }
    default:
        break
    }

    if !planner.Status.Active {
        return restart, nil
    }

    if r.MainProcess == nil {
        r.Events <- types.Start
        return restart, nil
    }

    if r.MetricsProcess == nil {
        r.StartMetricsProcess(ctx, planner)
    }

    if r.Cache != nil {
        r.Cache.Metrics.SetMaxAge(time.Second * time.Duration(planner.Spec.MetrcisMaxAge))
    }

    if planner.Status.Phase == appsv1.Waiting {
        nextStart := r.LastStart.Add(time.Second * time.Duration(planner.Spec.PlanningInterval))
        if nextStart.Before(time.Now()) {
            r.Cache.Clear()
            go r.Informer.GetInfo(r.MainProcess.Context, r.Events, r.Cache, r.Client, planner.Spec)
            r.LastStart = time.Now()
            r.UpdatePhase(planner, appsv1.Waiting)
            r.Informer.UpdatePlanner(ctx, r.Client, planner)
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

func (r *PlannerReconciler) ProcessEvent(ctx context.Context, planner *appsv1.Planner, e types.Event) bool {
    switch e {
    case types.Start:
        if !planner.Status.Active || r.MainProcess == nil {
            planner.Status.Active = true
            r.UpdatePhase(planner, appsv1.Waiting)
            context, cancelFunc := context.WithCancel(ctx)
            r.MainProcess = &Process{
                Context:    context,
                CancelFunc: cancelFunc,
            }
            log.Info("Planner started")
            return true
        }
    case types.Stop:
        if planner.Status.Active {
            planner.Status.Active = false
            r.UpdatePhase(planner, appsv1.Waiting)
            if r.MainProcess != nil {
                r.MainProcess.CancelFunc()
                r.MainProcess = nil
            }
            if r.MetricsProcess != nil {
                r.MetricsProcess.CancelFunc()
                r.MetricsProcess = nil
            }
            log.Info("Planner stopped")
            return true
        }
    case types.InformingEnded:
        go resourceupdater.UpdatePodResources(r.MainProcess.Context, r.Events, r.Cache, planner.Spec)
        r.UpdatePhase(planner, appsv1.ResourcesUpdating)
        return true
    case types.ResourceUpdatingEnded:
        go rescheduler.GenPlan(r.MainProcess.Context, r.Events, r.Cache, planner.Spec)
        r.UpdatePhase(planner, appsv1.Planning)
        return true
    case types.PlanningEnded:
        go r.Executor.ExecutePlan(r.MainProcess.Context, r.Events, r.Cache, r.Client, r.Clientset, planner.Spec)
        r.UpdatePhase(planner, appsv1.Executing)
        return true
    case types.ExecutingEnded:
        r.UpdatePhase(planner, appsv1.Waiting)
        return true
    case types.PhaseEndedWithError:
        nextStart := r.LastStart.Add(time.Second * time.Duration(planner.Spec.PlanningInterval))
        log.Info("Error. Planner will restart after ", nextStart)
        r.UpdatePhase(planner, appsv1.Waiting)
        return true
    }
    return false
}

func (r *PlannerReconciler) StartMetricsProcess(ctx context.Context, planner *appsv1.Planner) {
    context, cancelFunc := context.WithCancel(ctx)
    r.MetricsProcess = &Process{
        Context:    context,
        CancelFunc: cancelFunc,
    }
    go r.Informer.RunMetircsListener(r.MetricsProcess.Context, r.Cache, r.MetricsClient, planner.Spec)
}

func (r *PlannerReconciler) UpdatePhase(planner *appsv1.Planner, phase appsv1.PlannerPhase) {
    planner.Status.Phase = phase
    switch phase {
        case appsv1.Waiting:
            r.Cache.Phase = "Waiting"
        case appsv1.Informing:
            r.Cache.Phase = "Collecting info"
        case appsv1.ResourcesUpdating:
            r.Cache.Phase = "Resource updating"
        case appsv1.Planning:
            r.Cache.Phase = "Plan generating"
        case appsv1.Executing:
            r.Cache.Phase = "Plan executing"
    }
}


