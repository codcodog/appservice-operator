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

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appv1 "github.com/codcodog/appservice-operator/api/v1"
	"github.com/codcodog/appservice-operator/resources"
)

// AppServiceReconciler reconciles a AppService object
type AppServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
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
	xlog := log.FromContext(ctx).WithValues("AppService", req.NamespacedName)
	xlog.Info("appservice-operator reconcile")

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
		xlog.Error(err, "Deployment not ready")
		return ctrl.Result{}, err
	}
	if err := r.ensureService(ctx, req, &instance); err != nil {
		xlog.Error(err, "Service not ready")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// 若不存在 deployment 则创建
// 若存在，则去更新
func (r *AppServiceReconciler) ensureDeployment(ctx context.Context, req ctrl.Request, instance *appv1.AppService) error {
	var deployment v1.Deployment
	err := r.Client.Get(ctx, req.NamespacedName, &deployment)
	if err != nil {
		if errors.IsNotFound(err) {
			if err := r.Client.Create(ctx, resources.NewDeployment(instance)); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

// 若不存在 service 则创建
// 若存在，则去更新
func (r *AppServiceReconciler) ensureService(ctx context.Context, req ctrl.Request, instance *appv1.AppService) error {
	var service corev1.Service
	err := r.Client.Get(ctx, req.NamespacedName, &service)
	if err != nil {
		if errors.IsNotFound(err) {
			if err := r.Client.Create(ctx, resources.NewService(instance)); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AppServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1.AppService{}).
		Complete(r)
}
