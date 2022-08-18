package resources

import (
	appv1 "github.com/codcodog/appservice-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewService(appService *appv1.AppService) *corev1.Service {
	labels := map[string]string{
		"app": appService.ObjectMeta.Name,
	}

	service := corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      appService.ObjectMeta.Name,
			Namespace: appService.Namespace,
			// 标注主从关系
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(appService.GetObjectMeta(), appService.GroupVersionKind()),
			},
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeNodePort,
			Ports:    appService.Spec.Ports,
			Selector: labels,
		},
	}

	return &service
}
