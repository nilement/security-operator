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
	"strings"
	"time"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	experimentsv1alpha1 "github.com/nilement/security-operator/api/v1alpha1"
)

// MisconfigurationReconciler reconciles a Misconfiguration object
type MisconfigurationReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=experiments.chaosplatform.com,resources=misconfigurations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=experiments.chaosplatform.com,resources=misconfigurations/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=experiments.chaosplatform.com,resources=misconfigurations/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Misconfiguration object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *MisconfigurationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("Misconfiguration", req.NamespacedName)

	misconfiguration := &experimentsv1alpha1.Misconfiguration{}
	err := r.Get(ctx, req.NamespacedName, misconfiguration)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Misconfiguration resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Misconfiguration")
		return ctrl.Result{}, err
	}

	found := &corev1.Pod{}
	err = r.Get(ctx, types.NamespacedName{Name: misconfiguration.Name, Namespace: misconfiguration.Namespace}, found)

	misconfigurations := misconfiguration.Spec.KubeletMisconfigurations

	if err != nil && errors.IsNotFound(err) {
		if len(misconfigurations) > 0 {
			pod := r.constructPodForKubeletMisconfig(misconfiguration)
			log.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
			err = r.Create(ctx, pod)
			if err != nil {
				log.Error(err, "Failed to create new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
				return ctrl.Result{}, err
			}
			// Pod created successfully - return and requeue
			return ctrl.Result{Requeue: true}, nil
		}
	} else if err != nil {
		log.Error(err, "Failed to get the Pod")
		return ctrl.Result{}, err
	}

	if !found.DeletionTimestamp.IsZero() {
		log.Info("Pod is being deleted", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)
		return ctrl.Result{RequeueAfter: time.Second * 5}, nil
	}

	if !r.compareKubeletMisconfigurations(misconfigurations, found.Annotations) {
		log.Info("Delete old Pod due to new Kubelet misconfigs", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)
		err = r.Delete(ctx, found)
		if err != nil {
			log.Error(err, "Failed to delete the old Pod", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)
			return ctrl.Result{}, err
		}
		// Pod created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}

func (r *MisconfigurationReconciler) compareKubeletMisconfigurations(state []string, spec map[string]string) bool {
	annotations := spec["kubelet"]
	if len(state) == 0 && annotations == "" {
		return true
	}
	experiments := strings.Split(annotations, ";")
	if len(state) != len(experiments) {
		return false
	}
	for idx := range experiments {
		if experiments[idx] != state[idx] {
			return false
		}
	}
	return true
}

var jobOwnerMisconfigKey = ".metadata.misconfig.controller"

// SetupWithManager sets up the controller with the Manager.
func (r *MisconfigurationReconciler) SetupWithManager(mgr ctrl.Manager) error {

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &batchv1.Job{}, jobOwnerMisconfigKey, func(rawObj client.Object) []string {
		// grab the job object, extract the owner...
		job := rawObj.(*batchv1.Job)
		owner := metav1.GetControllerOf(job)
		if owner == nil {
			return nil
		}
		// ...make sure it's a Misconfiguration...
		if owner.APIVersion != apiGVStr || owner.Kind != "Misconfiguration" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&experimentsv1alpha1.Misconfiguration{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}

func (r *MisconfigurationReconciler) constructPodForKubeletMisconfig(e *experimentsv1alpha1.Misconfiguration) *corev1.Pod {
	configs := e.Spec.KubeletMisconfigurations
	configs = append(configs, e.Spec.KubeletMisconfigurations...)
	spec := *e.Spec.PodTemplate.DeepCopy()
	spec.Containers[0].Args = configs

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
			Name:        e.Name,
			Namespace:   e.Namespace,
		},
		Spec: spec,
	}
	pod.ObjectMeta.Annotations["kubelet"] = strings.Join(e.Spec.KubeletMisconfigurations, ";")
	pod.ObjectMeta.Labels["SecurityChaos"] = "experiment"
	ctrl.SetControllerReference(e, pod, r.Scheme)
	return pod
}
