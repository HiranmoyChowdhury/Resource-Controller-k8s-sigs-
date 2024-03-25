package controller

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	bastardv1 "resource.controller.sigs/resource-controller-k8s-sigs/api/v1"
	"resource.controller.sigs/resource-controller-k8s-sigs/utils"
)

func setDefaultFields(bastard *bastardv1.Bastard) {
	if bastard.Spec.DeletionPolicy == "" {
		var str string = utils.DefaultDeletionPolicy
		bastard.Spec.DeletionPolicy = bastardv1.DeletionPolicy(str)
	}
	if bastard.Spec.ServiceSpec.ServiceType == "" {
		var str string = utils.DefaultServiceType
		bastard.Spec.ServiceSpec.ServiceType = corev1.ServiceType(str)
	}
}
func setDeployOwner(deployment *appsv1.Deployment, bastard *bastardv1.Bastard) {
	if bastard.Spec.DeletionPolicy == "WipeOut" {

		deployment.OwnerReferences = []metav1.OwnerReference{
			*metav1.NewControllerRef(bastard, bastardv1.GroupVersion.WithKind(bastard.Kind)),
		}
	}
}
func setSvcOwner(service *corev1.Service, bastard *bastardv1.Bastard) {
	if bastard.Spec.DeletionPolicy == "WipeOut" {
		service.OwnerReferences = []metav1.OwnerReference{
			*metav1.NewControllerRef(bastard, bastardv1.GroupVersion.WithKind(bastard.Kind)),
		}
	}

}
