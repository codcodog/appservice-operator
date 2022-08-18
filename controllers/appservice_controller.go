/*
Copyright 2022.

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
	"encoding/json"
	"reflect"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appv1 "github.com/codcodog/appservice-operator/api/v1"
	"github.com/codcodog/appservice-operator/resources"
	"github.com/go-logr/logr"
)

// AppServiceReconciler reconciles a AppService object
type AppServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	log    logr.Logger
}

//+kubebuilder:rbac:groups=app.codcodog.com,resources=appservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.codcodog.com,resources=appservices/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=app.codcodog.com,resources=appservices/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AppService object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *AppServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.log = log.FromContext(ctx).WithValues("AppService", req.NamespacedName)
	r.log.Info("appservice-operator reconcile")

	var instance appv1.AppService
	if err := r.Client.Get(ctx, req.NamespacedName, &instance); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	if instance.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	if err := r.ensureDeployment(ctx, req, &instance); err != nil {
		r.log.Error(err, "Deployment not ready")
		return ctrl.Result{}, err
	}
	if err := r.ensureService(ctx, req, &instance); err != nil {
		r.log.Error(err, "Service not ready")
		return ctrl.Result{}, err
	}
	if err := r.updateAnnotations(ctx, &instance); err != nil {
		r.log.Error(err, "Update annotations failed")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// 若不存在 deployment 则创建
// 若存在，则对比更新
func (r *AppServiceReconciler) ensureDeployment(ctx context.Context, req ctrl.Request, instance *appv1.AppService) error {
	var deployment v1.Deployment
	err := r.Client.Get(ctx, req.NamespacedName, &deployment)
	if err != nil {
		if errors.IsNotFound(err) {
			err = r.Client.Create(ctx, resources.NewDeployment(instance))
			if err != nil {
				r.log.Error(err, "create deployment error")
				return err
			}
		} else {
			r.log.Error(err, "get deployment error")
			return err
		}
	}

	isChanged, err := r.isChanged(instance)
	if err != nil {
		return err
	}

	if isChanged {
		newDeployment := resources.NewDeployment(instance)
		deployment.Spec = newDeployment.Spec
		if err := r.Client.Update(ctx, &deployment); err != nil {
			r.log.Error(err, "update deployment error")
			return err
		}
	}

	return nil
}

// 若不存在 service 则创建
// 若存在，则对比更新
func (r *AppServiceReconciler) ensureService(ctx context.Context, req ctrl.Request, instance *appv1.AppService) error {
	var service corev1.Service
	err := r.Client.Get(ctx, req.NamespacedName, &service)
	if err != nil {
		if errors.IsNotFound(err) {
			if err := r.Client.Create(ctx, resources.NewService(instance)); err != nil {
				r.log.Error(err, "create service error")
				return err
			}
		} else {
			r.log.Error(err, "get service error")
			return err
		}
	}

	isChanged, err := r.isChanged(instance)
	if err != nil {
		return err
	}

	if isChanged {
		clusterIP := service.Spec.ClusterIP
		newService := resources.NewService(instance)
		service.Spec = newService.Spec
		service.Spec.ClusterIP = clusterIP // # Spec.ClusterIP is imutable

		if err := r.Client.Update(ctx, &service); err != nil {
			r.log.Error(err, "update service error")
			return err
		}
	}

	return nil
}

// 更新 annotations
// 用来对比下次CRD是否有更新
func (r *AppServiceReconciler) updateAnnotations(ctx context.Context, instance *appv1.AppService) error {
	data, err := json.Marshal(instance.Spec)
	if err != nil {
		return err
	}

	if instance.Annotations != nil {
		instance.Annotations["spec"] = string(data)
	} else {
		instance.Annotations = map[string]string{"spec": string(data)}
	}

	if err := r.Client.Update(ctx, instance); err != nil {
		r.log.Error(err, "update annotations error")
		return err
	}

	return nil
}

// CRD是否发生变更
func (r *AppServiceReconciler) isChanged(instance *appv1.AppService) (bool, error) {
	// 若不存在，直接返回已变化
	if _, ok := instance.Annotations["spec"]; !ok {
		return true, nil
	}

	var oldSpec appv1.AppServiceSpec
	err := json.Unmarshal([]byte(instance.Annotations["spec"]), &oldSpec)
	if err != nil {
		r.log.Error(err, "json Unmarshal error")
		return false, err
	}

	if !reflect.DeepEqual(oldSpec, instance.Spec) {
		return true, nil
	}

	return false, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AppServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1.AppService{}).
		Owns(&v1.Deployment{}). // 监听从属资源
		Owns(&corev1.Service{}).
		Complete(r)
}
