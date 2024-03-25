package controller

import (
	"bytes"
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	namespcedname "k8s.io/apimachinery/pkg/types"
	bastardv1 "resource.controller.sigs/resource-controller-k8s-sigs/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"resource.controller.sigs/resource-controller-k8s-sigs/utils"
)

func (r *BastardReconciler) newDeployment(bastard *bastardv1.Bastard, name string, deployment *appsv1.Deployment) {

	labels := make(map[string]string)
	for k, v := range bastard.Spec.Labels {
		labels[k] = v
	}
	labels["dracarys"] = "im-now-the-servant-of-the-white-walkers"
	labels["uid"] = string(bastard.UID)

	deployment.Name = name

	deployment.Spec.Replicas = bastard.Spec.DeploymentSpec.Replicas

	deployment.OwnerReferences = []metav1.OwnerReference{
		*metav1.NewControllerRef(bastard, bastardv1.GroupVersion.WithKind(bastard.Kind)),
	}
	playWithOwner(&deployment.OwnerReferences, bastard)

	if bastard.ObjectMeta.Namespace != "" {
		deployment.Namespace = bastard.ObjectMeta.Namespace
	}
	deployment.Labels = labels

	deploymentImage := bastard.Spec.DeploymentSpec.Image

	containerPorts := []corev1.ContainerPort{}

	if bastard.Spec.ServiceSpec.TargetPort != nil {
		containerPorts[0].ContainerPort = *bastard.Spec.ServiceSpec.TargetPort
	}
	deployment.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: labels,
	}
	deployment.Spec.Template = corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    utils.ContainerName,
					Image:   deploymentImage,
					Command: bastard.Spec.DeploymentSpec.Commands,
					Ports:   containerPorts,
				},
			},
		},
	}
}

func (r *BastardReconciler) newService(bastard *bastardv1.Bastard, name string, service *corev1.Service) *corev1.Service {
	labels := make(map[string]string)
	for k, v := range bastard.Spec.Labels {
		labels[k] = v
	}
	labels["dracarys"] = "im-now-the-servant-of-the-white-walkers"
	labels["uid"] = string(bastard.UID)

	service.Name = name
	serviceType := bastard.Spec.ServiceSpec.ServiceType

	if serviceType == "Headless" {
		serviceType = ""
	}
	service.Spec.Type = corev1.ServiceType(serviceType)

	servicePort := bastard.Spec.ServiceSpec.Port
	ports := []corev1.ServicePort{
		{
			Port: *servicePort,
		},
	}

	service.OwnerReferences = []metav1.OwnerReference{
		*metav1.NewControllerRef(bastard, bastardv1.GroupVersion.WithKind(bastard.Kind)),
	}
	playWithOwner(&service.OwnerReferences, bastard)

	if bastard.ObjectMeta.Namespace != "" {
		service.Namespace = bastard.ObjectMeta.Namespace
	}
	service.Labels = labels

	if bastard.Spec.ServiceSpec.NodePort != nil {
		ports[0].NodePort = *bastard.Spec.ServiceSpec.NodePort
	}
	if bastard.Spec.ServiceSpec.TargetPort != nil {
		ports[0].TargetPort.IntVal = *bastard.Spec.ServiceSpec.TargetPort
	}

	service.Spec.Ports = ports
	service.Spec.Selector = labels

	return service
}
func (r *BastardReconciler) GetDeploymentName(bastard *bastardv1.Bastard) string {
	UID := bastard.UID
	deploymentList := appsv1.DeploymentList{}
	err := r.List(context.TODO(), &deploymentList, client.InNamespace(bastard.Namespace), client.MatchingLabels{"dracarys": "im-now-the-servant-of-the-white-walkers"})

	if err == nil {
		for _, deployment := range deploymentList.Items {
			if deployment.Labels["uid"] == string(UID) {
				return deployment.Name
			}
		}
	}
	var deploymentName bytes.Buffer
	deploymentName.WriteString(ToLowerCase(bastard.Name))
	if bastard.Spec.DeploymentSpec.Name != "" {
		deploymentName.WriteString("-")
		deploymentName.WriteString(ToLowerCase(bastard.Spec.DeploymentSpec.Name))
	}

	for i := 0; i != -1; i++ {
		name, err := r.deploymentNameIsExist(bastard, deploymentName.String(), int32(i))
		if err == nil {
			return name
		}
	}

	return deploymentName.String()
}

func (r *BastardReconciler) deploymentNameIsExist(bastard *bastardv1.Bastard, name string, cnt int32) (string, error) {
	_name := fmt.Sprintf("%s%s%s", name, "-", String(cnt))

	err := r.Get(context.TODO(), namespcedname.NamespacedName{bastard.Namespace, _name}, &appsv1.Deployment{})
	if err != nil {
		return _name, nil
	}
	return "", fmt.Errorf("deployment Name has already occupied")

}
func (r *BastardReconciler) GetServiceName(bastard *bastardv1.Bastard) string {
	UID := bastard.UID
	serviceList := &corev1.ServiceList{}
	err := r.List(context.TODO(), serviceList, client.InNamespace(bastard.Namespace), client.MatchingLabels{"dracarys": "im-now-the-servant-of-the-white-walkers"})
	if err == nil {
		for _, service := range serviceList.Items {
			if service.Labels["uid"] == string(UID) {
				return service.Name
			}
		}
	}

	svcName := ToLowerCase(bastard.Name)
	if bastard.Spec.ServiceSpec.Name != "" {
		svcName = fmt.Sprintf("%s%s%s", svcName, "-", ToLowerCase(bastard.Spec.ServiceSpec.Name))
	}

	for i := 0; i != -1; i++ {
		name, err := r.ServiceNameExist(bastard, svcName, int32(i))
		if err == nil {
			return name
		}
	}

	return svcName
}

func (r *BastardReconciler) ServiceNameExist(bastard *bastardv1.Bastard, name string, cnt int32) (string, error) {
	_name := fmt.Sprintf("%s%s%s", name, "-", String(cnt))
	err := r.Get(context.TODO(), namespcedname.NamespacedName{bastard.Namespace, _name}, &corev1.Service{})

	if err != nil {
		return _name, nil
	}
	return "", fmt.Errorf("service Name already occupied")

}

func deploymentSpecGotUpdate(bastard *bastardv1.Bastard, deployment *appsv1.Deployment) bool {
	if (bastard.Spec.DeploymentSpec.Replicas != nil && *bastard.Spec.DeploymentSpec.Replicas != *deployment.Spec.Replicas) == true {
		return true
	}
	if (bastard.Spec.DeploymentSpec.Image != "" && bastard.Spec.DeploymentSpec.Image != deployment.Spec.Template.Spec.Containers[0].Image) ||
		(bastard.Spec.DeletionPolicy == "WipeOut" && *deployment.OwnerReferences[0].BlockOwnerDeletion == false) ||
		(bastard.Spec.DeletionPolicy == "Delete" && *deployment.OwnerReferences[0].BlockOwnerDeletion == true) {
		return true
	}
	return false

}
func serviceSpecGotUpdate(bastard *bastardv1.Bastard, service *corev1.Service) bool {
	if (bastard.Spec.ServiceSpec.Port != nil && *bastard.Spec.ServiceSpec.Port != service.Spec.Ports[0].Port) ||
		(bastard.Spec.ServiceSpec.NodePort != nil && *bastard.Spec.ServiceSpec.NodePort != service.Spec.Ports[0].NodePort) ||
		(bastard.Spec.ServiceSpec.TargetPort != nil && *bastard.Spec.ServiceSpec.TargetPort != service.Spec.Ports[0].TargetPort.IntVal) ||
		(bastard.Spec.DeletionPolicy == "WipeOut" && *service.OwnerReferences[0].BlockOwnerDeletion == false) ||
		(bastard.Spec.DeletionPolicy == "Delete" && *service.OwnerReferences[0].BlockOwnerDeletion == true) {
		return true
	}
	return false
}
func String(n int32) string {
	buf := [11]byte{}
	pos := len(buf)
	i := int64(n)
	signed := i < 0
	if signed {
		i = -i
	}
	for {
		pos--
		buf[pos], i = '0'+byte(i%10), i/10
		if i == 0 {
			if signed {
				pos--
				buf[pos] = '-'
			}
			return string(buf[pos:])
		}
	}
}
func ToLowerCase(s string) string {
	var result string = ""
	for _, char := range s {
		if char >= 'A' && char <= 'Z' {
			result += string(char + 32)
		} else if char <= 'a' && char >= 'z' {
			continue
		} else {
			result += string(char)
		}
	}
	return result + "-"
}
