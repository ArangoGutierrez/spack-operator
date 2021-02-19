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

	s "strings"

	packagev1alpha1 "github.com/ArangoGutierrez/spack-operator/api/v1alpha1"
	buildv1 "github.com/openshift/api/build/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *BuildReconciler) createBuild(ctx context.Context, spkg *packagev1alpha1.Build) (ctrl.Result, error) {

	r.Log.Info("Creating package build", "package", spkg.Name)
	tmp := spkg.DeepCopy()
	// Create a configMap from the Spack environment on the CR

	baseBuildRecipe := new(string)
	*baseBuildRecipe = `
FROM spack-operator-base:spackv0.16.0 as builder

COPY ./spack.yaml /opt/spack-environment
COPY ./build.sh /usr/bin
RUN chmod a+x /usr/bin/build.sh
RUN mkdir -p /opt/view

RUN /usr/bin/build.sh
`

	// ensures that data stored in the ConfigMap cannot
	// be updated (only object metadata can be modified).
	immutable := new(bool)
	*immutable = true
	// configMap
	cm := corev1.ConfigMap{
		metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		metav1.ObjectMeta{
			Name:      s.Join([]string{tmp.Name, "env"}, "-"),
			Namespace: "spack-operator-system",
		},
		immutable,
		// TODO: create a config map for each environment in the BuildSpec CR
		map[string]string{*tmp.Spec.Environment[0].Name: *tmp.Spec.Environment[0].Data},
		map[string][]byte{},
	}
	if err := r.Client.Create(ctx, &cm); err != nil {
		r.Log.Error(err, "Failed to create the configMap")
		return ctrl.Result{Requeue: true, RequeueAfter: 1 * time.Second}, err
	}

	// configMapBuildSource
	// DestinationDir set to default (same context as the Dockerfile)
	cmbs := buildv1.ConfigMapBuildSource{
		ConfigMap: corev1.LocalObjectReference{
			Name: s.Join([]string{tmp.Name, "env"}, "-"),
		},
	}

	buildLogic := buildv1.ConfigMapBuildSource{
		ConfigMap: corev1.LocalObjectReference{
			Name: "spack-build-logic",
		},
	}

	// TODO: take this value from the CR and default to 3 if not given
	l := int32(3)
	// Create the buildConfig
	bc := &buildv1.BuildConfig{
		metav1.TypeMeta{},
		metav1.ObjectMeta{
			Name:      s.Join([]string{tmp.Name, "buildconfig"}, "-"),
			Namespace: tmp.Namespace,
		},
		buildv1.BuildConfigSpec{
			RunPolicy:                    buildv1.BuildRunPolicyParallel,
			SuccessfulBuildsHistoryLimit: &l,
			FailedBuildsHistoryLimit:     &l,
			Triggers:                     []buildv1.BuildTriggerPolicy{{Type: "ConfigChange"}},
			CommonSpec: buildv1.CommonSpec{
				Strategy: buildv1.BuildStrategy{
					Type: "Docker",
					DockerStrategy: &buildv1.DockerBuildStrategy{
						From: &corev1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: "spack-operator-base:spackv0.16.0",
						},
					},
				},
				Source: buildv1.BuildSource{
					Type:       "Dockerfile",
					Dockerfile: baseBuildRecipe,
					ConfigMaps: []buildv1.ConfigMapBuildSource{cmbs, buildLogic},
				},
				Output: buildv1.BuildOutput{
					To: &corev1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: tmp.Spec.ImageStream,
					},
					ImageLabels: []buildv1.ImageLabel{
						// TODO: labels about the arch and other interesting build aspects
						{"built-by", "multiarch-operator"},
					},
				},
				NodeSelector: nil,
			},
		},
		buildv1.BuildConfigStatus{},
	}
	if err := r.Client.Create(ctx, bc); err != nil {
		r.Log.Error(err, "Failed to create the BuildConfig")
		return ctrl.Result{}, err
	}

	//update the build status
	opts := []client.UpdateOption{}
	tmp.Status.State = packagev1alpha1.InitializedStatus
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

func (r *BuildReconciler) deleteBuild(ctx context.Context, spkg *packagev1alpha1.Build) (ctrl.Result, error) {

	r.Log.Info("Deleting package buildConfig", "package", spkg.Name)

	return ctrl.Result{Requeue: false}, nil
}
