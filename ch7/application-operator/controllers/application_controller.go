/*
Copyright 2024 Geray.

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
	"reflect"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	dappsv1 "github.com/geray/application-operator/api/v1"
)

const GenericRequeueDuration = 1 * time.Minute

var CounterReconcileApplication int64

// ApplicationReconciler reconciles a Application object
type ApplicationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:resources:path=applications,singular=application,scope=Namespaced,shortName=app

//+kubebuilder:rbac:groups=apps.geray.cn,resources=applications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.geray.cn,resources=applications/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.geray.cn,resources=applications/finalizers,verbs=update

//+kubebuilder:rbac:groups=apps.geray.cn,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.geray.cn,resources=deployments/status,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.geray.cn,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.geray.cn,resources=services/status,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Application object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.2/pkg/reconcile

func (r *ApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	<-time.NewTicker(100 * time.Millisecond).C
	log := log.FromContext(ctx)

	// TODO(user): your logic here
	CounterReconcileApplication += 1
	log.Info("Starting a reconcile", "number", CounterReconcileApplication)

	app := &dappsv1.Application{}
	if err := r.Get(ctx, req.NamespacedName, app); err != nil {
		if errors.IsNotFound(err) {
			log.Info("Application not found.")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get the Applicaion, will requeue after a short time.")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}

	// reconcile sub-resources
	var result ctrl.Result
	var err error

	result, err = r.reconcileDeployment(ctx, app)
	if err != nil {
		log.Error(err, "Failed to reconcile Deployment.")
		return result, err
	}

	result, err = r.reconcileService(ctx, app)
	if err != nil {
		log.Error(err, "Failed to reconcile Service.")
		return result, err
	}

	log.Info("All resources have been reconciled.")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	setupLog := ctrl.Log.WithName("setup")

	return ctrl.NewControllerManagedBy(mgr).
		For(&dappsv1.Application{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				return false
			},
			DeleteFunc: func(event event.DeleteEvent) bool {
				setupLog.Info("The Deployment has been deleted.", "name", event.Object.GetName())
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				if event.ObjectNew.GetResourceVersion() == event.ObjectOld.GetResourceVersion() {
					return false

				}
				if reflect.DeepEqual(event.ObjectNew.(*appsv1.Deployment).Spec, event.ObjectOld.(*appsv1.Deployment).Spec) {
					return false
				}
				return true
			},
		})).
		// Deployment
		Owns(&appsv1.Deployment{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(ce event.CreateEvent) bool {
				return false
			},
			DeleteFunc: func(event event.DeleteEvent) bool {
				setupLog.Info("The Deployment has been deleted.", "name", event.Object.GetName)
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				if event.ObjectNew.GetResourceVersion() == event.ObjectOld.GetResourceVersion() {
					return false
				}
				if reflect.DeepEqual(event.ObjectNew.(*appsv1.Deployment).Spec, event.ObjectOld.(*appsv1.Deployment).Spec) {
					return false
				}
				return true
			},
			GenericFunc: nil,
		})).
		// Service
		Owns(&corev1.Service{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(ce event.CreateEvent) bool {
				return false
			},
			DeleteFunc: func(event event.DeleteEvent) bool {
				setupLog.Info("The Service has been deleted.", "name", event.Object.GetName())
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				if event.ObjectNew.GetResourceVersion() == event.ObjectOld.GetResourceVersion() {
					return false
				}
				if reflect.DeepEqual(event.ObjectNew.(*dappsv1.Application).Spec, event.ObjectOld.(*dappsv1.Application).Spec) {
					return false
				}
				return true
			},
		})).
		Complete(r)
}
