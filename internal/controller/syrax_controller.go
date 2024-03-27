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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	namespcedname "k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/record"
	syraxv1 "resource.controller.sigs/resource-controller-k8s-sigs/api/v1"
	targaryenv1 "resource.controller.sigs/resource-controller-k8s-sigs/api/v1"
	"resource.controller.sigs/resource-controller-k8s-sigs/utils"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

// SyraxReconciler reconciles a Syrax object
type SyraxReconciler struct {
	client.Client
	SubRCClient client.SubResourceClient
	Cache       cache.Cache
	Scheme      *runtime.Scheme
	Recorder    record.EventRecorder
}

//+kubebuilder:rbac:groups=targaryen.resource.controller.sigs,resources=syraxs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=targaryen.resource.controller.sigs,resources=syraxs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=targaryen.resource.controller.sigs,resources=syraxs/finalizers,verbs=update
//+kubebuilder:rbac:groups=targaryen.resource.controller.sigs,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=targaryen.resource.controller.sigs,resources=deployments/status,verbs=get
//+kubebuilder:rbac:groups=targaryen.resource.controller.sigs,resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Syrax object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *SyraxReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	fmt.Println("Reconcile started")

	syrax := &syraxv1.Syrax{}
	err := r.Get(context.TODO(), req.NamespacedName, syrax)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("syrax '%s' in work queue no longer exists", req.NamespacedName))
			return ctrl.Result{}, nil

		}
		return ctrl.Result{}, nil
	}

	if ctrlutil.ContainsFinalizer(syrax, utils.DefaultFinalizer) == false && syrax.Spec.DeletionPolicy == "Delete" {
		ctrlutil.AddFinalizer(syrax, utils.DefaultFinalizer)
		err = r.Update(context.TODO(), syrax)
	}
	if err != nil {
		return ctrl.Result{}, err
	}

	setDefaultFields(syrax)

	deploymentName := r.getDeploymentName(syrax)
	serviceName := r.getServiceName(syrax)

	deployment := &appsv1.Deployment{}
	if err = r.Get(context.TODO(), namespcedname.NamespacedName{req.Namespace, deploymentName}, deployment); err != nil {
		r.newDeployment(syrax, deploymentName, deployment)
		err = r.Create(context.TODO(), deployment)
	}
	if err != nil {
		r.Recorder.Event(syrax, "Warning", err.Error(), fmt.Sprintf("the deployment for syrax kind with name %s is not present and can't be created", syrax.Name))
		return ctrl.Result{}, nil
	}

	service := &corev1.Service{}
	if err = r.Get(context.TODO(), namespcedname.NamespacedName{req.Namespace, serviceName}, service); err != nil {
		service = r.newService(syrax, serviceName, service)
		err = r.Create(context.TODO(), service)
	}

	if err != nil {
		r.Recorder.Event(syrax, "Warning", err.Error(), fmt.Sprintf("the service for syrax kind with name %s is not present and can't be created", syrax.Name))
		return ctrl.Result{}, nil
	}

	if ifDeployUpdated(syrax, deployment) == true {
		log.Log.Info("Update deployment resource")
		r.newDeployment(syrax, deploymentName, deployment)
		err = r.Update(context.TODO(), deployment)
	}
	if err != nil {
		r.Recorder.Event(syrax, "Normal", "", fmt.Sprintf("the deployment for syrax kind with name %s failed in updatation, requeue", syrax.Name))
		return ctrl.Result{}, err
	} else {
		r.Recorder.Event(syrax, "Normal", "", fmt.Sprintf("the deployment for syrax kind with name %s is successfully updated", syrax.Name))
	}

	if ifSvcUpdated(syrax, service) {
		log.Log.Info("Update service resource")
		r.newService(syrax, serviceName, service)
		err = r.Update(context.TODO(), service)
	}
	if err != nil {
		r.Recorder.Event(syrax, "Normal", "", fmt.Sprintf("the service for syrax kind with name %s failed in updatation, requeue", syrax.Name))
		return ctrl.Result{}, err
	} else {
		r.Recorder.Event(syrax, "Normal", "", fmt.Sprintf("the service for syrax kind with name %s is successfully updated", syrax.Name))
	}
	if syrax.DeletionTimestamp != nil {
		ctrlutil.RemoveFinalizer(syrax, utils.DefaultFinalizer)
		time.Sleep(20 * time.Millisecond)
		r.Update(context.TODO(), syrax)
	}

	err = r.updateSyraxStatus(syrax, deployment, service)
	if err != nil {
		r.Recorder.Event(syrax, "Normal", "", fmt.Sprintf("the status subresource for syrax kind with name %s failed in update", syrax.Name))
		return ctrl.Result{}, err
	}

	r.Recorder.Event(syrax, "Normal", "", fmt.Sprintf("for syrax kind with name %s everything is fine", syrax.Name))

	return ctrl.Result{}, nil
}
func (r *SyraxReconciler) updateSyraxStatus(syrax *syraxv1.Syrax, deployment *appsv1.Deployment, service *corev1.Service) error {

	syrax.Status.AvailableReplicas = &deployment.Status.AvailableReplicas

	err := r.Status().Update(context.TODO(), syrax)
	return err

}

var (
	jobOwnerKey = ".metadata.controller"
)

func (r *SyraxReconciler) SetupWithManager(mgr ctrl.Manager) error {

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &appsv1.Deployment{}, jobOwnerKey, func(rawObj client.Object) []string {
		depu := rawObj.(*appsv1.Deployment)

		owner := metav1.GetControllerOf(depu)
		if owner == nil {
			return nil
		}
		// ...make sure it's a CronJob...
		if owner.Kind != utils.Kind {
			return nil
		}
		return []string{owner.Name}
	}); err != nil {
		return err
	}
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &corev1.Service{}, jobOwnerKey, func(rawObj client.Object) []string {
		// grab the job object, extract the owner...
		svc := rawObj.(*corev1.Service)
		owner := metav1.GetControllerOf(svc)
		if owner == nil {
			return nil
		}
		// ...make sure it's a CronJob...
		if owner.Kind != utils.Kind {
			return nil
		}
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&targaryenv1.Syrax{}).
		Owns(&appsv1.Deployment{}, builder.MatchEveryOwner).
		Owns(&corev1.Service{}, builder.MatchEveryOwner).
		Complete(r)
}
