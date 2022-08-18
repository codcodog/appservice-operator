package resources

import (
	appv1 "github.com/codcodog/appservice-operator/api/v1"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewDeployment(appService *appv1.AppService) *v1.Deployment {
	labels := map[string]string{
		"app": appService.ObjectMeta.Name,
	}
	selector := &metav1.LabelSelector{MatchLabels: labels}

	deployment := v1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      appService.Name,
			Namespace: appService.Namespace,
			// 标注主从关系
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(appService.GetObjectMeta(), appService.GroupVersionKind()),
			},
		},
		Spec: v1.DeploymentSpec{
			Replicas: appService.Spec.Replicas,
			Selector: selector,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: newContainers(appService),
				},
			},
		},
	}

	return &deployment
}

func newContainers(appService *appv1.AppService) []corev1.Container {
	var containerPorts []corev1.ContainerPort
	for _, item := range appService.Spec.Ports {
		containerPort := corev1.ContainerPort{
			ContainerPort: item.TargetPort.IntVal,
		}
		containerPorts = append(containerPorts, containerPort)
	}

	return []corev1.Container{
		corev1.Container{
			Name:            appService.Name,
			Image:           appService.Spec.Image,
			Ports:           containerPorts,
			ImagePullPolicy: corev1.PullIfNotPresent,
		},
	}
}
