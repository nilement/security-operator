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

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	experimentsv1alpha1 "github.com/nilement/security-operator/api/v1alpha1"
)

// ValidationReconciler reconciles a Validation object
type ValidationReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=experiments.chaosplatform.com,resources=validations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=experiments.chaosplatform.com,resources=validations/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=experiments.chaosplatform.com,resources=validations/finalizers,verbs=update
// +kubebuilder:rbac:groups=experiments.chaosplatform.com,resources=cisexperiments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=experiments.chaosplatform.com,resources=cisexperiments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=experiments.chaosplatform.com,resources=cisexperiments/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Validation object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile

func ReconcileValidator() {

}

func (r *ValidationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// var node corev1.Node

	validators := &experimentsv1alpha1.ValidationList{}
	err := r.List(ctx, validators, client.InNamespace(req.Namespace), client.MatchingLabels{"securitychaos": "validation"})
	log := r.Log.WithValues("validation", req.NamespacedName)

	if len(validators.Items) == 0 {
		log.Info("No valdiators registered.")
		return ctrl.Result{}, nil
	}
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Validation resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Validation")
		return ctrl.Result{}, err
	}

	var activeJobs []*batchv1.Job
	var successfulJobs []*batchv1.Job
	var failedJobs []*batchv1.Job

	// check jobs
	var jobs batchv1.JobList

	err = r.List(ctx, &jobs, client.InNamespace(req.Namespace), client.MatchingLabels{"SecurityChaos": "experiment"})
	if err != nil {
		log.Error(err, "Unable to list jobs")
		return ctrl.Result{}, err
	}

	for i, job := range jobs.Items {
		_, finishedType := r.isJobFinished(&job)
		switch finishedType {
		case "": // ongoing
			activeJobs = append(activeJobs, &jobs.Items[i])
		case batchv1.JobFailed:
			failedJobs = append(failedJobs, &jobs.Items[i])
		case batchv1.JobComplete:
			successfulJobs = append(successfulJobs, &jobs.Items[i])
		}
	}

	// check pods
	var pods corev1.PodList
	err = r.List(ctx, &pods, client.InNamespace(req.Namespace), client.MatchingLabels{"SecurityChaos": "experiment"})
	if err != nil {
		log.Error(err, "Unable to list pods")
		return ctrl.Result{}, err
	}

	newPods := 0
	newJobs := 0
	var latest metav1.Time
	var latestJob metav1.Time
	nodes := make(map[string]bool)
	for _, validator := range validators.Items {
		for _, job := range successfulJobs {
			if job.Status.CompletionTime.After(validator.Status.LastJob.Time) {
				newJobs++
				if job.Status.CompletionTime.After(latestJob.Time) {
					latestJob = *job.Status.CompletionTime
				}
			}
		}
		if newJobs > 0 {
			validator.Status.LastJob = latestJob
		}
		for _, pod := range pods.Items {
			if pod.CreationTimestamp.After(validator.Status.LastPod.Time) {
				nodes[pod.Spec.NodeName] = true
				newPods++
				if pod.CreationTimestamp.After(latest.Time) {
					latest = pod.CreationTimestamp
				}
			}
		}
		if newPods > 0 {
			validator.Status.LastPod = latest
		}
		completions := newPods + newJobs

		if completions >= validator.Spec.ExperimentsToTrigger {
			for nd := range nodes {
				job := r.jobForKubeBench(&validator, nd)
				log.Info("Creating a new validation Job on Node:", "Job.Namespace", job.Namespace, "Job.Name", job.Name, "Node.Name", nd)
				err = r.Create(ctx, job)
				if err != nil {
					log.Error(err, "Failed to create new Job", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
					return ctrl.Result{RequeueAfter: 5}, err
				}
			}
		}

		err = r.Status().Update(ctx, &validator)
		if err != nil {
			log.Error(err, "Failed to update Validation status")
			return ctrl.Result{}, err
		}
	}

	fmt.Println("Validation controller is running")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ValidationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	chaosApiFilter := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return e.Object.GetLabels()["SecurityChaos"] == "experiment"
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return e.ObjectNew.GetLabels()["SecurityChaos"] == "experiment"
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&experimentsv1alpha1.Validation{}).
		Owns(&batchv1.Job{}).
		Owns(&corev1.Pod{}).
		Watches(&source.Kind{
			Type: &corev1.Pod{},
		},
			&handler.EnqueueRequestForObject{},
			builder.WithPredicates(chaosApiFilter),
		).
		Complete(r)
}

func (r *ValidationReconciler) jobForKubeBench(e *experimentsv1alpha1.Validation, node string) *batchv1.Job {
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Labels:       make(map[string]string),
			Annotations:  make(map[string]string),
			GenerateName: "kube-bench-",
			Namespace:    e.Namespace,
		},
		Spec: *e.Spec.JobTemplate.Spec.DeepCopy(),
	}
	for k, v := range e.Spec.JobTemplate.Annotations {
		job.Annotations[k] = v
	}
	for k, v := range e.Spec.JobTemplate.Labels {
		job.Labels[k] = v
	}
	job.Spec.Template.Spec.NodeName = node
	ctrl.SetControllerReference(e, job, r.Scheme)
	return job
}

func (r *ValidationReconciler) jobForKubeHunter(e *experimentsv1alpha1.Validation) *batchv1.Job {
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Labels:       make(map[string]string),
			Annotations:  make(map[string]string),
			GenerateName: "kube-hunter-",
			Namespace:    e.Namespace,
		},
		Spec: *e.Spec.JobTemplate.Spec.DeepCopy(),
	}
	for k, v := range e.Spec.JobTemplate.Annotations {
		job.Annotations[k] = v
	}
	for k, v := range e.Spec.JobTemplate.Labels {
		job.Labels[k] = v
	}
	ctrl.SetControllerReference(e, job, r.Scheme)
	return job
}

func (r *ValidationReconciler) isJobFinished(job *batchv1.Job) (bool, batchv1.JobConditionType) {
	for _, c := range job.Status.Conditions {
		if (c.Type == batchv1.JobComplete || c.Type == batchv1.JobFailed) && c.Status == corev1.ConditionTrue {
			return true, c.Type
		}
	}

	return false, ""
}
