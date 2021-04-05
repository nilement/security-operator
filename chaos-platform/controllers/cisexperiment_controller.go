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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	experimentsv1alpha1 "github.com/nilement/security-operator/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

// CisExperimentReconciler reconciles a CisExperiment object
type CisExperimentReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=experiments.chaosplatform.com,resources=cisexperiments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=experiments.chaosplatform.com,resources=cisexperiments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=experiments.chaosplatform.com,resources=cisexperiments/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CisExperiment object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *CisExperimentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("cisexperiment", req.NamespacedName)

	cisExperiment := &experimentsv1alpha1.CisExperiment{}
	err := r.Get(ctx, req.NamespacedName, cisExperiment)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("CIS Experiment resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get CIS Experiment")
		return ctrl.Result{}, err
	}

	var jobs batchv1.JobList

	err = r.List(ctx, &jobs, client.InNamespace(req.Namespace), client.MatchingFields{jobOwnerCisKey: req.Name})
	if err != nil {
		log.Error(err, "Unable to list jobs")
		return ctrl.Result{}, err
	}

	var activeJobs []*batchv1.Job
	var successfulJobs []*batchv1.Job
	var failedJobs []*batchv1.Job

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

	completions := len(activeJobs) + len(successfulJobs)

	if int64(completions) < 1 {
		job := r.constructJobForCISMisconfig(cisExperiment)
		log.Info("Applying CIS 4.1.1 misconfiguration", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
		err = r.Create(ctx, job)
		if err != nil {
			log.Error(err, "Failed to create new Job", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
			return ctrl.Result{RequeueAfter: 5}, err
		}
	}

	fmt.Println("CIS Experiment controller is running")

	return ctrl.Result{}, nil
}

func (r *CisExperimentReconciler) constructJobForCISMisconfig(e *experimentsv1alpha1.CisExperiment) *batchv1.Job {
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Labels:       make(map[string]string),
			Annotations:  make(map[string]string),
			GenerateName: "cis-4.1.1-",
			Namespace:    e.Namespace,
		},
		Spec: *e.Spec.JobTemplate.Spec.DeepCopy(),
	}
	for k, v := range e.Spec.JobTemplate.ObjectMeta.Annotations {
		job.Annotations[k] = v
	}
	for k, v := range e.Spec.JobTemplate.ObjectMeta.Labels {
		job.Labels[k] = v
	}
	job.Labels["SecurityChaos"] = "experiment"

	ctrl.SetControllerReference(e, job, r.Scheme)
	return job
}

func (r *CisExperimentReconciler) isJobFinished(job *batchv1.Job) (bool, batchv1.JobConditionType) {
	for _, c := range job.Status.Conditions {
		if (c.Type == batchv1.JobComplete || c.Type == batchv1.JobFailed) && c.Status == corev1.ConditionTrue {
			return true, c.Type
		}
	}

	return false, ""
}

var (
	jobOwnerCisKey = ".metadata.cis.controller"
	apiGVStr       = "experiments.chaosplatform.com/v1alpha1"
)

// SetupWithManager sets up the controller with the Manager.
func (r *CisExperimentReconciler) SetupWithManager(mgr ctrl.Manager) error {

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &batchv1.Job{}, jobOwnerCisKey, func(rawObj client.Object) []string {
		// grab the job object, extract the owner...
		job := rawObj.(*batchv1.Job)
		owner := metav1.GetControllerOf(job)
		if owner == nil {
			return nil
		}
		// ...make sure it's a CisExperiment...
		if owner.APIVersion != apiGVStr || owner.Kind != "CisExperiment" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&experimentsv1alpha1.CisExperiment{}).
		Owns(&batchv1.Job{}).
		Complete(r)
}
