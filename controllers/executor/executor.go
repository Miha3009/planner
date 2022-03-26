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

package executor

import (
	"context"
	"time"
	types "github.com/miha3009/planner/controllers/types"
	appsv1 "github.com/miha3009/planner/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
//	driver "github.com/miha3009/planner/controllers/executor/driver"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/prometheus/common/log"
)

type NodeDriver interface {
	AddNode() bool
	DeleteNode(node *corev1.Node) bool
}

func ExecutePlan(ctx context.Context, events chan types.Event, cache *types.PlannerCache, clt client.Client, cltset *clientset.Clientset, planner appsv1.PlannerSpec) {
	/*dp, _ := cltset.AppsV1().Deployments("default").List(ctx, metav1.ListOptions{})
	for i := range dp.Items {
		log.Info(dp.Items[i].Name)
		err := cltset.AppsV1().Deployments("default").Delete(ctx, dp.Items[i].Name, metav1.DeleteOptions{})
		if err != nil {
			log.Info(err)
		}
	}*/

	cache.Plan.Movements = []types.Movement{genRandomMove(cache)}
	for _, move := range cache.Plan.Movements {
		movePod(ctx, cltset, move)
	}
	
	log.Info("Plan executed")
	events <- types.ExecutingEnded
}

func movePod (ctx context.Context, cltset *clientset.Clientset, move types.Movement) {
	newPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: move.Pod.Namespace,
			Name:      NewNameForPod(move.Pod),
			//Labels:    move.Pod.Labels,
		},
		Spec: move.Pod.Spec,
	}

	err := createPod(ctx, cltset, newPod, move.NewNode)
	if err != nil {
		log.Info(err)
		return
	}
	
	for !isPodRunning(ctx, cltset, newPod) {
		time.Sleep(time.Second)
	}
	
	/*dp, _ := cltset.AppsV1().Deployments("default").List(ctx, metav1.ListOptions{})
	for i := range dp.Items {
		dp.Items[i] = 
		_, err := cltset.AppsV1().Deployments("default").Update(ctx, &dp.Items[i], metav1.UpdateOptions{})
		if err != nil {
			log.Info(err)
		}
	}*/
			
	deletePod(ctx, cltset, move.Pod)
	
	newPod.Labels = move.Pod.Labels
	_, err = cltset.CoreV1().Pods("default").Update(ctx, newPod, metav1.UpdateOptions{})
	if err != nil {
		log.Info(err)
	}
}

func isPodRunning(ctx context.Context, cltset *clientset.Clientset, pod *corev1.Pod) bool {
	pod, err := cltset.CoreV1().Pods(pod.Namespace).Get(ctx, pod.Name, metav1.GetOptions{})
	if err != nil {
		log.Info(err)
	}
	return pod.Status.Phase == corev1.PodRunning
}

func deletePod(ctx context.Context, cltset *clientset.Clientset, pod *corev1.Pod) {
	err := cltset.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{})
	if err != nil {
		log.Info(err)
	}
}

func createPod(ctx context.Context, cltset *clientset.Clientset, pod *corev1.Pod, node *corev1.Node) error {
	pod.Spec.NodeName = node.Name
	_, err := cltset.CoreV1().Pods(pod.Namespace).Create(ctx, pod, metav1.CreateOptions{})
	return err
}

func genRandomMove(cache *types.PlannerCache) types.Movement {
	pod := &corev1.Pod{}
	oldNode := &corev1.Node{}
	newNode := &corev1.Node{}
	for i := range cache.Pods {
		for j := range cache.Pods[i] {
			pod = &cache.Pods[i][j]
			oldNode = &cache.Nodes[i]
		}
	}
	
	if oldNode.Name == cache.Nodes[0].Name {
		newNode = &cache.Nodes[1]
	} else {
		newNode = &cache.Nodes[0]
	}

	move := types.Movement{Pod: pod, OldNode: oldNode, NewNode: newNode,}

	log.Info("OldNode: ", oldNode.Name, ", NewNode: ", newNode.Name)
	
	return move
}

