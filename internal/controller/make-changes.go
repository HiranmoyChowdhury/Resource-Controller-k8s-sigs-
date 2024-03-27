package controller

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	syraxv1 "resource.controller.sigs/resource-controller-k8s-sigs/api/v1"
	"resource.controller.sigs/resource-controller-k8s-sigs/utils"
)

func setDefaultFields(syrax *syraxv1.Syrax) {
	if syrax.Spec.DeletionPolicy == "" {
		var str string = utils.DefaultDeletionPolicy
		syrax.Spec.DeletionPolicy = syraxv1.DeletionPolicy(str)
	}
	if syrax.Spec.ServiceSpec.ServiceType == "" {
		var str string = utils.DefaultServiceType
		syrax.Spec.ServiceSpec.ServiceType = corev1.ServiceType(str)
	}
}
func setDeployOwner(deployment *appsv1.Deployment, syrax *syraxv1.Syrax) {

	deployment.OwnerReferences = []metav1.OwnerReference{
		*metav1.NewControllerRef(syrax, syraxv1.GroupVersion.WithKind(syrax.Kind)),
	}
}
func setSvcOwner(service *corev1.Service, syrax *syraxv1.Syrax) {
	service.OwnerReferences = []metav1.OwnerReference{
		*metav1.NewControllerRef(syrax, syraxv1.GroupVersion.WithKind(syrax.Kind)),
	}

}
