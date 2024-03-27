package controller

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	syraxv1 "resource.controller.sigs/resource-controller-k8s-sigs/api/v1"
	"resource.controller.sigs/resource-controller-k8s-sigs/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
func setOwner(object client.Object, syrax *syraxv1.Syrax) {
	if syrax.DeletionTimestamp != nil {
		object.SetOwnerReferences(nil)
		return
	}
	object.SetOwnerReferences([]metav1.OwnerReference{
		*metav1.NewControllerRef(syrax, syraxv1.GroupVersion.WithKind(syrax.Kind)),
	})
}
