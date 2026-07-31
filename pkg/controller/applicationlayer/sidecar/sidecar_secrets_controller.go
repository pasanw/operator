// Copyright (c) 2026 Tigera, Inc. All rights reserved.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package sidecar reconciles the image pull secrets that Application Layer
// sidecar injection needs. When a pod opts into sidecar injection
// (applicationlayer.projectcalico.org/sidecar=true) the mutating webhook injects
// private-registry dikastes/envoy containers into it. Those images can only be
// pulled if the pull secret exists in the pod's namespace. This controller copies
// the Installation pull secrets into every namespace that has a sidecar pod, and
// cleans up copies in namespaces that no longer do. It mirrors the Istio waypoint
// pull-secret controller.
package sidecar

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	operatorv1 "github.com/tigera/operator/api/v1"
	"github.com/tigera/operator/pkg/common"
	"github.com/tigera/operator/pkg/controller/options"
	"github.com/tigera/operator/pkg/controller/utils"
	"github.com/tigera/operator/pkg/ctrlruntime"
	"github.com/tigera/operator/pkg/render"
	"github.com/tigera/operator/pkg/render/common/secret"
)

const (
	// SidecarPodLabel is the pod label that opts a workload into Application
	// Layer sidecar injection (matched by the sidecar mutating webhook).
	SidecarPodLabel = "applicationlayer.projectcalico.org/sidecar"

	// SidecarPullSecretLabel labels the pull-secret copies this controller
	// manages, so they can be found and cleaned up with a single label query
	// (the same rationale as the waypoint controller).
	SidecarPullSecretLabel = "operator.tigera.io/applicationlayer-sidecar-pull-secret"
)

var log = logf.Log.WithName("controller_applicationlayer_sidecar_secrets")

// Add creates the sidecar pull-secret controller and adds it to the Manager.
func Add(mgr manager.Manager, opts options.ControllerOptions) error {
	if !opts.EnterpriseCRDExists {
		return nil
	}

	r := &ReconcileSidecarSecrets{
		Client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
	}

	c, err := ctrlruntime.NewController("applicationlayer-sidecar-secrets-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return fmt.Errorf("failed to create applicationlayer-sidecar-secrets-controller: %w", err)
	}

	// Watch pods that opt into sidecar injection; only their namespaces need the
	// pull secret. The predicate keeps us from reconciling on unrelated pod churn,
	// and every event funnels to a single reconcile that re-derives the full set.
	sidecarPod := predicate.NewPredicateFuncs(func(o client.Object) bool {
		return o.GetLabels()[SidecarPodLabel] == "true"
	})
	if err := c.WatchObject(&corev1.Pod{}, handler.EnqueueRequestsFromMapFunc(enqueueSingleton), sidecarPod); err != nil {
		return fmt.Errorf("applicationlayer-sidecar-secrets-controller failed to watch Pods: %w", err)
	}

	// Watch the ApplicationLayer CR for SidecarInjection enable/disable.
	if err := c.WatchObject(&operatorv1.ApplicationLayer{}, &handler.EnqueueRequestForObject{}); err != nil {
		return fmt.Errorf("applicationlayer-sidecar-secrets-controller failed to watch ApplicationLayer: %w", err)
	}

	// Watch Installation for pull-secret changes.
	if err := utils.AddInstallationWatch(c); err != nil {
		return fmt.Errorf("applicationlayer-sidecar-secrets-controller failed to watch Installation: %w", err)
	}

	// Periodic reconcile as a backstop.
	if err := utils.AddPeriodicReconcile(c, utils.PeriodicReconcileTime, &handler.EnqueueRequestForObject{}); err != nil {
		return fmt.Errorf("applicationlayer-sidecar-secrets-controller failed to create periodic reconcile: %w", err)
	}

	return nil
}

// enqueueSingleton funnels every triggering event to one reconcile request, since
// the reconcile always re-derives the full target-namespace set rather than acting
// on a single object.
func enqueueSingleton(_ context.Context, _ client.Object) []reconcile.Request {
	return []reconcile.Request{{NamespacedName: utils.DefaultEnterpriseInstanceKey}}
}

// ReconcileSidecarSecrets copies the Installation pull secrets into namespaces
// that contain Application Layer sidecar pods and cleans up stale copies.
type ReconcileSidecarSecrets struct {
	client.Client
	scheme *runtime.Scheme
}

func (r *ReconcileSidecarSecrets) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.V(1).Info("Reconciling Application Layer sidecar pull secrets")

	// The feature is active only while the ApplicationLayer CR exists, is not
	// being deleted, and has SidecarInjection enabled. Otherwise every copy is
	// stale and gets cleaned up.
	al := &operatorv1.ApplicationLayer{}
	err := r.Get(ctx, utils.DefaultEnterpriseInstanceKey, al)
	if err != nil && !apierrors.IsNotFound(err) {
		return reconcile.Result{}, err
	}
	active := err == nil &&
		al.DeletionTimestamp.IsZero() &&
		al.Spec.SidecarInjection != nil &&
		*al.Spec.SidecarInjection == operatorv1.SidecarEnabled

	toCreate, toDelete, err := r.pullSecretChanges(ctx, active, reqLogger)
	if err != nil {
		return reconcile.Result{}, err
	}

	hdlr := utils.NewComponentHandler(log, r, r.scheme, nil)
	component := render.NewPassthrough(toCreate, toDelete)
	if err := hdlr.CreateOrUpdateOrDelete(ctx, component, nil); err != nil {
		return reconcile.Result{}, fmt.Errorf("failed to reconcile sidecar pull secrets: %w", err)
	}

	return reconcile.Result{}, nil
}

// pullSecretChanges determines which pull-secret copies must exist (toCreate),
// based on the namespaces that contain sidecar pods, and which existing copies
// are stale (toDelete). When active is false no copies are desired, so every
// existing copy is returned as stale.
func (r *ReconcileSidecarSecrets) pullSecretChanges(ctx context.Context, active bool, reqLogger logr.Logger) (toCreate, toDelete []client.Object, err error) {
	if active {
		_, installationSpec, err := utils.GetInstallationSpec(ctx, r)
		if err != nil {
			if apierrors.IsNotFound(err) {
				reqLogger.V(1).Info("Installation not found")
				return nil, nil, nil
			}
			return nil, nil, err
		}

		pullSecrets, err := utils.GetInstallationPullSecrets(installationSpec, r)
		if err != nil {
			return nil, nil, err
		}

		if len(pullSecrets) > 0 {
			podList := &corev1.PodList{}
			if err := r.List(ctx, podList, client.MatchingLabels{SidecarPodLabel: "true"}); err != nil {
				return nil, nil, fmt.Errorf("failed to list sidecar pods: %w", err)
			}

			targetNamespaces := map[string]bool{}
			for i := range podList.Items {
				ns := podList.Items[i].Namespace
				// The operator's own namespaces already carry the pull secret via
				// the normal render path; only tenant namespaces need a copy.
				if ns != common.OperatorNamespace() && ns != common.CalicoNamespace {
					targetNamespaces[ns] = true
				}
			}

			for ns := range targetNamespaces {
				// Bind the tigera-operator-secrets ClusterRole to the operator SA
				// scoped to this tenant namespace FIRST, so the operator is permitted
				// to write the pull secret here. This mirrors the egress-gateway /
				// log-storage pattern and avoids any cluster-wide secret-write grant.
				rb := render.CreateOperatorSecretsRoleBinding(ns)
				if rb.Labels == nil {
					rb.Labels = map[string]string{}
				}
				rb.Labels[SidecarPullSecretLabel] = "true"
				toCreate = append(toCreate, rb)

				for _, s := range secret.CopyToNamespace(ns, pullSecrets...) {
					if s.Labels == nil {
						s.Labels = map[string]string{}
					}
					s.Labels[SidecarPullSecretLabel] = "true"
					toCreate = append(toCreate, s)
				}
			}
		}
	}

	// Diff the desired set (keyed on namespace/name) against existing managed
	// copies so renamed or no-longer-needed copies are cleaned up.
	desired := map[types.NamespacedName]bool{}
	for _, obj := range toCreate {
		desired[types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()}] = true
	}

	existing := &corev1.SecretList{}
	if err := r.List(ctx, existing, client.MatchingLabels{SidecarPullSecretLabel: "true"}); err != nil {
		return nil, nil, fmt.Errorf("failed to list sidecar pull secrets: %w", err)
	}
	for i := range existing.Items {
		s := &existing.Items[i]
		if !desired[types.NamespacedName{Namespace: s.Namespace, Name: s.Name}] {
			toDelete = append(toDelete, s)
		}
	}

	existingRB := &rbacv1.RoleBindingList{}
	if err := r.List(ctx, existingRB, client.MatchingLabels{SidecarPullSecretLabel: "true"}); err != nil {
		return nil, nil, fmt.Errorf("failed to list sidecar pull-secret rolebindings: %w", err)
	}
	for i := range existingRB.Items {
		rb := &existingRB.Items[i]
		if !desired[types.NamespacedName{Namespace: rb.Namespace, Name: rb.Name}] {
			toDelete = append(toDelete, rb)
		}
	}

	return toCreate, toDelete, nil
}
