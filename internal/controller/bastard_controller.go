/*
Copyright 2024.

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

package controller

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	namespcedname "k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/record"
	bastardv1 "resource.controller.sigs/resource-controller-k8s-sigs/api/v1"
	targaryenv1 "resource.controller.sigs/resource-controller-k8s-sigs/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// BastardReconciler reconciles a Bastard object
type BastardReconciler struct {
	client.Client
	SubRCClient client.SubResourceClient
	Scheme      *runtime.Scheme
	Recorder    record.EventRecorder
}

//+kubebuilder:rbac:groups=targaryen.resource.controller.sigs,resources=bastards,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=targaryen.resource.controller.sigs,resources=bastards/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=targaryen.resource.controller.sigs,resources=bastards/finalizers,verbs=update
//+kubebuilder:rbac:groups=targaryen.resource.controller.sigs,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=targaryen.resource.controller.sigs,resources=deployments/status,verbs=get
//+kubebuilder:rbac:groups=targaryen.resource.controller.sigs,resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Bastard object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *BastardReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	fmt.Println("Reconcile started")

	bastard := &bastardv1.Bastard{}
	err := r.Get(context.TODO(), req.NamespacedName, bastard)

	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("bastard '%s' in work queue no longer exists", req.NamespacedName))
			return ctrl.Result{}, nil

		}

		return ctrl.Result{}, nil

	}
	fillTheGapFirst(bastard)

	deploymentName := r.GetDeploymentName(bastard)
	serviceName := r.GetServiceName(bastard)

	deployment := &appsv1.Deployment{}
	if err = r.Get(context.TODO(), namespcedname.NamespacedName{req.Namespace, deploymentName}, deployment); err != nil {
		r.newDeployment(bastard, deploymentName, deployment)
		err = r.Create(context.TODO(), deployment)
	}
	if err != nil {
		r.Recorder.Event(bastard, "Warning", err.Error(), fmt.Sprintf("the deployment for bastard kind with name %s is not present and can't be created", bastard.Name))
		return ctrl.Result{}, nil
	}

	service := &corev1.Service{}
	if err = r.Get(context.TODO(), namespcedname.NamespacedName{req.Namespace, serviceName}, service); err != nil {
		service = r.newService(bastard, serviceName, service)
		err = r.Create(context.TODO(), service)
	}

	if err != nil {
		r.Recorder.Event(bastard, "Warning", err.Error(), fmt.Sprintf("the service for bastard kind with name %s is not present and can't be created", bastard.Name))
		return ctrl.Result{}, nil
	}

	if deploymentSpecGotUpdate(bastard, deployment) == true {
		log.Log.Info("Update deployment resource")
		r.newDeployment(bastard, deploymentName, deployment)
		err = r.Update(context.TODO(), deployment)
	}
	if err != nil {
		r.Recorder.Event(bastard, "Normal", "", fmt.Sprintf("the deployment for bastard kind with name %s failed in updatation, requeue", bastard.Name))
		return ctrl.Result{}, err
	} else {
		r.Recorder.Event(bastard, "Normal", "", fmt.Sprintf("the deployment for bastard kind with name %s is successfully updated", bastard.Name))
	}

	if serviceSpecGotUpdate(bastard, service) {
		log.Log.Info("Update service resource")
		r.newService(bastard, serviceName, service)
		err = r.Update(context.TODO(), service)
	}
	if err != nil {
		r.Recorder.Event(bastard, "Normal", "", fmt.Sprintf("the service for bastard kind with name %s failed in updatation, requeue", bastard.Name))
		return ctrl.Result{}, err
	} else {
		r.Recorder.Event(bastard, "Normal", "", fmt.Sprintf("the service for bastard kind with name %s is successfully updated", bastard.Name))
	}

	err = r.updateBastardStatus(bastard, deployment, service)
	if err != nil {
		r.Recorder.Event(bastard, "Normal", "", fmt.Sprintf("the status subresource for bastard kind with name %s failed in update", bastard.Name))
		return ctrl.Result{}, err
	}

	r.Recorder.Event(bastard, "Normal", "", fmt.Sprintf("for bastard kind with name %s everything is fine", bastard.Name))
	return ctrl.Result{}, nil
}
func (r *BastardReconciler) updateBastardStatus(bastard *bastardv1.Bastard, deployment *appsv1.Deployment, service *corev1.Service) error {

	bastard.Status.AvailableReplicas = &deployment.Status.AvailableReplicas

	err := r.Status().Update(context.TODO(), bastard)

	return err
}
func (r *BastardReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&targaryenv1.Bastard{}).
		Complete(r)
}
