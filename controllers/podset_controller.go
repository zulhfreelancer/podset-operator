/*


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
	"reflect"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appv1alpha1 "github.com/redhat/podset-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodSetReconciler reconciles a PodSet object
type PodSetReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=app.example.com,resources=podsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=app.example.com,resources=podsets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=v1,resources=pods,verbs=get;list;watch;create;update;patch;delete

// GetPodSetInstance TODO
func (r *PodSetReconciler) GetPodSetInstance(req ctrl.Request) (*appv1alpha1.PodSet, error) {
	instance := &appv1alpha1.PodSet{}
	err := r.Get(context.Background(), req.NamespacedName, instance)
	return instance, err
}

// Reconcile is the core logic of controller
func (r *PodSetReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("podset", req.NamespacedName)

	// Fetch the PodSet instance (the parent of the pods)
	podSet, err := r.GetPodSetInstance(req)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request
		return reconcile.Result{}, err
	}

	// List all pods owned by this PodSet instance
	podList := &corev1.PodList{}
	labelz := map[string]string{
		"app":     podSet.Name, // the metadata.name field from user's CR PodSet YAML file
		"version": "v0.1",
	}
	labelSelector := labels.SelectorFromSet(labelz)
	listOpts := &client.ListOptions{Namespace: podSet.Namespace, LabelSelector: labelSelector}
	if err = r.List(context.Background(), podList, listOpts); err != nil {
		return reconcile.Result{}, err
	}

	// Count the pods that are pending or running and add them to available array
	var available []corev1.Pod
	for _, pod := range podList.Items {
		if pod.ObjectMeta.DeletionTimestamp != nil {
			continue
		}
		if pod.Status.Phase == corev1.PodRunning || pod.Status.Phase == corev1.PodPending {
			available = append(available, pod)
		}
	}
	numAvailable := int32(len(available))
	availableNames := []string{}
	for _, pod := range available {
		availableNames = append(availableNames, pod.ObjectMeta.Name)
	}

	// Update the status if necessary
	status := appv1alpha1.PodSetStatus{
		PodNames:          availableNames,
		AvailableReplicas: numAvailable,
	}
	if !reflect.DeepEqual(podSet.Status, status) {
		// We need to refresh the PodSet before we can update it
		// https://github.com/operator-framework/operator-sdk/issues/3968
		podSet, _ = r.GetPodSetInstance(req)
		podSet.Status = status
		err = r.Status().Update(context.Background(), podSet)
		if err != nil {
			r.Log.Error(err, "Failed to update PodSet status")
			return reconcile.Result{}, err
		}
	}

	// When the number of pods in the cluster is bigger that what we want, scale down
	if numAvailable > podSet.Spec.Replicas {
		r.Log.Info("Scaling down pods", "Currently available", numAvailable, "Required replicas", podSet.Spec.Replicas)
		diff := numAvailable - podSet.Spec.Replicas
		toDeletePods := available[:diff] // Syntax help: https://play.golang.org/p/SHAMCdd12sp
		for _, toDeletePod := range toDeletePods {
			err = r.Delete(context.Background(), &toDeletePod)
			if err != nil {
				r.Log.Error(err, "Failed to delete pod", "pod.name", toDeletePod.Name)
				return reconcile.Result{}, err
			}
		}
		return reconcile.Result{Requeue: true}, nil
	}

	// When the number of pods in the cluster is smaller that what we want, scale up
	if numAvailable < podSet.Spec.Replicas {
		r.Log.Info("Scaling up pods", "Currently available", numAvailable, "Required replicas", podSet.Spec.Replicas)
		// Define a new Pod object
		pod := newPodForCR(podSet)
		// Set PodSet instance as the owner of the Pod
		if err := controllerutil.SetControllerReference(podSet, pod, r.Scheme); err != nil {
			return reconcile.Result{}, err
		}
		err = r.Create(context.Background(), pod)
		if err != nil {
			r.Log.Error(err, "Failed to create pod", "pod.name", pod.Name)
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *appv1alpha1.PodSet) *corev1.Pod {
	labels := map[string]string{
		"app":     cr.Name, // the metadata.name field from user's CR PodSet YAML file
		"version": "v0.1",
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: cr.Name + "-pod-", // Pod name example: podset-sample-pod-jzxbw
			Namespace:    cr.Namespace,
			Labels:       labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}

// SetupWithManager defines how the controller will watch for resources
func (r *PodSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1alpha1.PodSet{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}
