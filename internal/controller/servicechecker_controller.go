/*
Copyright 2025 mamrezb.

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

package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	maintenancev1alpha1 "github.com/mamrezb/maintenance-window-manager/api/v1alpha1"
)

// ServiceCheckerReconciler reconciles a ServiceChecker object
type ServiceCheckerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=maintenance.mamrezb.com,resources=servicecheckers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=maintenance.mamrezb.com,resources=servicecheckers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=maintenance.mamrezb.com,resources=servicecheckers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ServiceChecker object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.4/pkg/reconcile
func (r *ServiceCheckerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// 1. Fetch the ServiceChecker instance
	sc := &maintenancev1alpha1.ServiceChecker{}
	if err := r.Get(ctx, req.NamespacedName, sc); err != nil {
		if k8serrors.IsNotFound(err) {
			// The object might be deleted
			logger.Info("ServiceChecker resource not found. Ignoring since object must be deleted.")
			return reconcile.Result{}, nil
		}
		logger.Error(err, "Failed to get ServiceChecker")
		return reconcile.Result{}, err
	}

	logger.Info("Reconciling ServiceChecker", "name", sc.Name, "namespace", sc.Namespace)

	// 2. Build a new status based on the actual Endpoints we find
	newStatuses := make([]maintenancev1alpha1.ServiceStatus, 0, len(sc.Spec.Services))
	for _, svcRef := range sc.Spec.Services {
		svcReady := false

		// We'll check the Endpoints for that service
		endpointsObj := &corev1.Endpoints{}
		err := r.Get(ctx, types.NamespacedName{
			Name:      svcRef.Name,
			Namespace: svcRef.Namespace,
		}, endpointsObj)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				logger.Info("Endpoints not found", "svc", svcRef.Name, "ns", svcRef.Namespace)
			} else {
				logger.Error(err, "Error getting Endpoints", "svc", svcRef.Name, "ns", svcRef.Namespace)
			}
		} else {
			// If any Subset has at least one address, we consider it "Ready"
			for _, subset := range endpointsObj.Subsets {
				if len(subset.Addresses) > 0 {
					svcReady = true
					break
				}
			}
		}

		newStatuses = append(newStatuses, maintenancev1alpha1.ServiceStatus{
			Name:      svcRef.Name,
			Namespace: svcRef.Namespace,
			Ready:     svcReady,
		})
	}

	// 3. Update the ServiceChecker status
	sc.Status.ServiceStatuses = newStatuses
	if err := r.Status().Update(ctx, sc); err != nil {
		logger.Error(err, "Failed to update ServiceChecker status")
		return ctrl.Result{}, err
	}

	logger.Info("Successfully updated status", "ServiceStatuses", newStatuses)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceCheckerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// We watch:
	//  - ServiceChecker objects
	//  - Endpoints objects (so we see real-time changes)
	return ctrl.NewControllerManagedBy(mgr).
		For(&maintenancev1alpha1.ServiceChecker{}).
		Watches(
			&corev1.Endpoints{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
				endpointsObj, ok := obj.(*corev1.Endpoints)
				if !ok {
					return nil
				}
				return r.findRelatedServiceCheckers(ctx, endpointsObj)
			}),
		).
		Complete(r)
}

// findRelatedServiceCheckers scans all ServiceChecker CRs to see if they reference the Endpoints
func (r *ServiceCheckerReconciler) findRelatedServiceCheckers(ctx context.Context, endpointsObj *corev1.Endpoints) []reconcile.Request {
	logger := log.FromContext(ctx)

	// e.g., myv1alpha1.ServiceCheckerList if your CR is named "ServiceChecker"
	var scList maintenancev1alpha1.ServiceCheckerList
	if err := r.List(ctx, &scList); err != nil {
		logger.Error(err, "Failed to list ServiceCheckers")
		return nil
	}

	requests := make([]reconcile.Request, 0, len(scList.Items))
	for _, sc := range scList.Items {
		for _, svcRef := range sc.Spec.Services {
			if svcRef.Name == endpointsObj.Name && svcRef.Namespace == endpointsObj.Namespace {
				req := reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      sc.Name,
						Namespace: sc.Namespace,
					},
				}
				requests = append(requests, req)
				// We can break if no need to check multiple matches
				break
			}
		}
	}

	if len(requests) > 0 {
		logger.Info("Endpoints changed -> Reconcile ServiceChecker(s)",
			"endpoints", endpointsObj.Name, "namespace", endpointsObj.Namespace,
			"count", len(requests))
	}
	return requests
}
