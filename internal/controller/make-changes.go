package controller

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	bastardv1 "resource.controller.sigs/resource-controller-k8s-sigs/api/v1"
	"resource.controller.sigs/resource-controller-k8s-sigs/utils"
)

func fillTheGapFirst(bastard *bastardv1.Bastard) {
	if bastard.Spec.DeletionPolicy == "" {
		var str string = utils.DefaultDeletionPolicy
		bastard.Spec.DeletionPolicy = bastardv1.DeletionPolicy(str)
	}
	if bastard.Spec.ServiceSpec.ServiceType == "" {
		var str string = utils.DefaultServiceType
		bastard.Spec.ServiceSpec.ServiceType = corev1.ServiceType(str)
	}
}
func playWithOwner(reference *[]metav1.OwnerReference, bastard *bastardv1.Bastard) {
	if bastard.Spec.DeletionPolicy == "Delete" {
		reference = nil
	}
}
