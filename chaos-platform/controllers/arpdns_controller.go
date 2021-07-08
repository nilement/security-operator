/*
Copyright 2021.

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

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	experimentsv1alpha1 "github.com/nilement/security-operator/api/v1alpha1"
)

// ArpDnsReconciler reconciles a ArpDns object
type ArpDnsReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=experiments.chaosplatform.com,resources=arpdns,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=experiments.chaosplatform.com,resources=arpdns/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=experiments.chaosplatform.com,resources=arpdns/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ArpDns object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *ArpDnsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// r.Client.Get(ctx, )
	log := r.Log.WithValues("ArpDns", req.NamespacedName)

	arpDns := &experimentsv1alpha1.ArpDns{}
	err := r.Get(ctx, req.NamespacedName, arpDns)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("ArpDns resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get ArpDns.")
		return ctrl.Result{}, err
	}

	found := &corev1.Pod{}
	err = r.Get(ctx, types.NamespacedName{Name: arpDns.Name, Namespace: arpDns.Namespace}, found)

	if err != nil && errors.IsNotFound(err) {
		pod := r.constructPodForArpDns(arpDns)
		log.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
		err = r.Create(ctx, pod)
		if err != nil {
			log.Error(err, "Failed to create new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
			return ctrl.Result{}, err
		}
		// Pod created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get the Pod")
		return ctrl.Result{}, err
	}

	if !found.DeletionTimestamp.IsZero() {
		log.Info("Pod is being deleted", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)
		return ctrl.Result{RequeueAfter: time.Second * 5}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ArpDnsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&experimentsv1alpha1.ArpDns{}).
		Complete(r)
}

func (r *ArpDnsReconciler) constructPodForArpDns(e *experimentsv1alpha1.ArpDns) *corev1.Pod {
	spec := *e.Spec.PodTemplate.DeepCopy()
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
			Name:        e.Name,
			Namespace:   e.Namespace,
		},
		Spec: spec,
	}
	pod.ObjectMeta.Labels["SecurityChaos"] = "experiment"
	ctrl.SetControllerReference(e, pod, r.Scheme)
	return pod
}
