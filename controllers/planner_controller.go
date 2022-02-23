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

	appsv1 "github.com/miha3009/planner/api/v1"
	"github.com/go-logr/logr"
	"github.com/prometheus/common/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	
	log.Info("LOG PODS")
	podList := &corev1.PodList{}
    	opts := []client.ListOption{}
    	
	if err := r.Client.List(ctx, podList, opts...); err != nil {
		log.Error(err, "Failed to get pods")
		return ctrl.Result{}, err
	}
	for _, pod := range podList.Items {
		for _, container := range pod.Spec.Containers {
			log.Info("Pod name: ", pod.Name, ", Requested cpu: ", container.Resources.Requests.Cpu(), ", Requested memory: ", container.Resources.Requests.Memory())
		}
	}
	
	log.Info("LOG NODES")
	nodeList := &corev1.NodeList{}
    	opts = []client.ListOption{}
    	
	if err := r.Client.List(ctx, nodeList, opts...); err != nil {
		log.Error(err, "Failed to get nodes")
		return ctrl.Result{}, err
	}
	for _, node := range nodeList.Items {
		log.Info("Node name: ", node.Name, ", Capacity cpu: ", node.Status.Capacity.Cpu(), ", Capacity memory: ", node.Status.Capacity.Memory())
	}
	
	return ctrl.Result{RequeueAfter: time.Second*5}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PlannerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Planner{}).
		Complete(r)
}

