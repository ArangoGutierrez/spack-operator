/*
Copyright 2021.

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
	"time"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	packagev1alpha1 "github.com/ArangoGutierrez/spack-operator/api/v1alpha1"
)

// SpackReconciler reconciles a Spack object
type SpackReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=package.spack.io,resources=spacks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=package.spack.io,resources=spacks/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=package.spack.io,resources=spacks/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods/log,verbs=get
// +kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=config.openshift.io,resources=clusterversions,verbs=get
// +kubebuilder:rbac:groups=config.openshift.io,resources=proxies,verbs=get;list
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=security.openshift.io,resources=securitycontextconstraints,verbs=use;get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=image.openshift.io,resources=imagestreams,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=image.openshift.io,resources=imagestreams/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=image.openshift.io,resources=imagestreams/layers,verbs=get
// +kubebuilder:rbac:groups=core,resources=imagestreams/layers,verbs=get
// +kubebuilder:rbac:groups=build.openshift.io,resources=buildconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=build.openshift.io,resources=builds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=events,verbs=list;watch;create;update;patch
// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;update;
// +kubebuilder:rbac:groups=core,resources=persistentvolumes,verbs=get;list;watch;create;delete
// +kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=storage.k8s.io,resources=csinodes,verbs=get;list;watch
// +kubebuilder:rbac:groups=storage.k8s.io,resources=storageclasses,verbs=watch
// +kubebuilder:rbac:groups=storage.k8s.io,resources=csidrivers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=endpoints,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=prometheusrules,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *SpackReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("spack", req.NamespacedName)

	spkg := &packagev1alpha1.Spack{}
	err := r.Get(ctx, req.NamespacedName, spkg)
	// Error reading the object - requeue the request.
	if err != nil {
		// handle deletion of resource
		if errors.IsNotFound(err) {
			// User deleted the cluster resource so delete the pipeline resources
			r.Log.Info("resource has been deleted", "req", req.Name, "got", spkg.Name)
			return ctrl.Result{}, nil
		}

		r.Log.Error(err, "requeueing event since there was an error reading object")
		return ctrl.Result{Requeue: true}, err
	}

	r.Log.Info("reconciling at status: " + string(spkg.InstallStatus()))
	switch spkg.InstallStatus() {
	case packagev1alpha1.EmptyStatus:
		return r.createPackage(ctx, spkg)
	case packagev1alpha1.AppliedStatus:
		return r.validatePackage(ctx, spkg)
	case packagev1alpha1.ValidadtedPackage:
		r.Log.Info("Spack Package Validated", "package", spkg.Name)
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SpackReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&packagev1alpha1.Spack{}).
		Owns(&v1.Pod{}).
		Owns(&appsv1.DaemonSet{}).
		Complete(r)
}

func (r *SpackReconciler) createPackage(ctx context.Context, spkg *packagev1alpha1.Spack) (ctrl.Result, error) {

	r.Log.Info("Creating package build", "package", spkg.Name)
	tmp := spkg.DeepCopy()
	opts := []client.UpdateOption{}
	tmp.Status.State = packagev1alpha1.AppliedStatus
	if err := r.Client.Status().Update(ctx, tmp, opts...); err != nil {
		r.Log.Error(err, "status update failed")
		return ctrl.Result{}, err
	}

	objKey := types.NamespacedName{
		Namespace: tmp.Namespace,
		Name:      tmp.Name,
	}
	if err := r.Client.Get(ctx, objKey, tmp); err != nil {
		r.Log.Error(err, "status update failed to refresh object")
		return ctrl.Result{}, err
	}

	return ctrl.Result{Requeue: true, RequeueAfter: 3 * time.Second}, nil
}

func (r *SpackReconciler) validatePackage(ctx context.Context, spkg *packagev1alpha1.Spack) (ctrl.Result, error) {

	r.Log.Info("Validating package", "package", spkg.Name)
	tmp := spkg.DeepCopy()
	opts := []client.UpdateOption{}
	tmp.Status.State = packagev1alpha1.ValidadtedPackage
	if err := r.Client.Status().Update(ctx, tmp, opts...); err != nil {
		r.Log.Error(err, "status update failed")
		return ctrl.Result{}, err
	}

	objKey := types.NamespacedName{
		Namespace: tmp.Namespace,
		Name:      tmp.Name,
	}
	if err := r.Client.Get(ctx, objKey, tmp); err != nil {
		r.Log.Error(err, "status update failed to refresh object")
		return ctrl.Result{}, err
	}

	return ctrl.Result{Requeue: true, RequeueAfter: 3 * time.Second}, nil
}
