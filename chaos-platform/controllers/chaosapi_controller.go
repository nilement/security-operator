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
	"fmt"
	"reflect"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	experimentsv1alpha1 "github.com/nilement/security-operator/api/v1alpha1"
)

// ChaosApiReconciler reconciles a ChaosApi object
type ChaosApiReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=experiments.chaosplatform.com,resources=chaosapis,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=experiments.chaosplatform.com,resources=chaosapis/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=experiments.chaosplatform.com,resources=chaosapis/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ChaosApi object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *ChaosApiReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("chaosapi", req.NamespacedName)

	chaosapi := &experimentsv1alpha1.ChaosApi{}
	err := r.Get(ctx, req.NamespacedName, chaosapi)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("ChaosAPI resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get ChaosAPI")
		return ctrl.Result{}, err
	}

	found := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: chaosapi.Name, Namespace: chaosapi.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		dep := r.deploymentForMaliciousPod(chaosapi)
		log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	size := chaosapi.Spec.Size
	if *found.Spec.Replicas != size {
		found.Spec.Replicas = &size
		err = r.Update(ctx, found)
		if err != nil {
			log.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return ctrl.Result{}, err
		}
		// Spec updated - return and requeue
		return ctrl.Result{Requeue: true}, nil
	}

	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(chaosapi.Namespace),
		client.MatchingLabels(labelsForMaliciouspod(chaosapi.Name)),
	}
	if err = r.List(ctx, podList, listOpts...); err != nil {
		log.Error(err, "Failed to list pods", "ChaosApi.Namespace", chaosapi.Namespace, "ChaosApi.Name", chaosapi.Name)
		return ctrl.Result{}, err
	}
	podNames := getPodNames(podList.Items)

	if !reflect.DeepEqual(podNames, chaosapi.Status.Nodes) {
		chaosapi.Status.Nodes = podNames
		err := r.Status().Update(ctx, chaosapi)
		if err != nil {
			log.Error(err, "Failed to update chaosapi status")
			return ctrl.Result{}, err
		}
	}

	fmt.Println("Chaos API is running")

	return ctrl.Result{}, nil
}

func (r *ChaosApiReconciler) deploymentForMaliciousPod(e *experimentsv1alpha1.ChaosApi) *appsv1.Deployment {
	ls := labelsForMaliciouspod(e.Name)
	replicas := e.Spec.Size

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      e.Name,
			Namespace: e.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: "192.168.99.118:5000/alpinev2",
						Name:  "chaosapi-sample",
					}},
				},
			},
		},
	}

	ctrl.SetControllerReference(e, dep, r.Scheme)
	return dep
}

// labelsForMemcached returns the labels for selecting the resources
// belonging to the given memcached CR name.
func labelsForMaliciouspod(name string) map[string]string {
	return map[string]string{"app": "chaosapi-sample", "chaosapi-sample_cr": name}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

// SetupWithManager sets up the controller with the Manager.
func (r *ChaosApiReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&experimentsv1alpha1.ChaosApi{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
