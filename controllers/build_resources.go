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

	packagev1alpha1 "github.com/ArangoGutierrez/spack-operator/api/v1alpha1"
	build "github.com/openshift/api/build/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *BuildReconciler) createBuild(ctx context.Context, spkg *packagev1alpha1.Build) (ctrl.Result, error) {

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

	bc := &build.BuildConfig{}
	if err := r.Client.Create(ctx, bc); err != nil {
		r.Log.Error(err, "Failed to create the BuildConfig")
		return ctrl.Result{}, err
	}

	return ctrl.Result{Requeue: true, RequeueAfter: 3 * time.Second}, nil
}

func (r *BuildReconciler) validateBuild(ctx context.Context, spkg *packagev1alpha1.Build) (ctrl.Result, error) {

	r.Log.Info("Validating package", "package", spkg.Name)
	tmp := spkg.DeepCopy()
	opts := []client.UpdateOption{}
	tmp.Status.State = packagev1alpha1.ValidatedPackage
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
