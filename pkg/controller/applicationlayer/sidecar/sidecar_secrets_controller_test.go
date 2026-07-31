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

package sidecar

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	operatorv1 "github.com/tigera/operator/api/v1"
	"github.com/tigera/operator/pkg/apis"
	"github.com/tigera/operator/pkg/common"
	"github.com/tigera/operator/pkg/controller/utils"
	ctrlrfake "github.com/tigera/operator/pkg/ctrlruntime/client/fake"
	"github.com/tigera/operator/pkg/render"
)

var _ = Describe("Application Layer sidecar pull-secret controller", func() {
	var (
		cli          client.Client
		scheme       *runtime.Scheme
		ctx          context.Context
		r            *ReconcileSidecarSecrets
		installation *operatorv1.Installation
		al           *operatorv1.ApplicationLayer
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		Expect(apis.AddToScheme(scheme, false)).ShouldNot(HaveOccurred())

		ctx = context.Background()
		cli = ctrlrfake.DefaultFakeClientBuilder(scheme).Build()

		r = &ReconcileSidecarSecrets{Client: cli, scheme: scheme}

		installation = &operatorv1.Installation{
			ObjectMeta: metav1.ObjectMeta{Name: "default"},
			Spec:       operatorv1.InstallationSpec{Variant: operatorv1.Calico},
			Status:     operatorv1.InstallationStatus{Variant: operatorv1.Calico},
		}

		enabled := operatorv1.SidecarEnabled
		// Create the CR under the canonical singleton name so that a wrong lookup
		// key in the controller (the original item-3 bug) would leave the CR
		// unfound, make the feature inactive, and fail these copy assertions.
		al = &operatorv1.ApplicationLayer{
			ObjectMeta: metav1.ObjectMeta{Name: utils.DefaultEnterpriseInstanceKey.Name},
			Spec:       operatorv1.ApplicationLayerSpec{SidecarInjection: &enabled},
		}
	})

	createNamespace := func(name string) {
		err := cli.Create(ctx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name}})
		Expect(client.IgnoreAlreadyExists(err)).ShouldNot(HaveOccurred())
	}

	createPullSecret := func(name string) {
		Expect(cli.Create(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: common.OperatorNamespace()},
			Data:       map[string][]byte{".dockerconfigjson": []byte(`{"auths":{}}`)},
			Type:       corev1.SecretTypeDockerConfigJson,
		})).NotTo(HaveOccurred())
	}

	createPod := func(name, ns string, sidecar bool) {
		createNamespace(ns)
		labels := map[string]string{}
		if sidecar {
			labels[SidecarPodLabel] = "true"
		}
		Expect(cli.Create(ctx, &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns, Labels: labels},
			Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "app", Image: "app"}}},
		})).NotTo(HaveOccurred())
	}

	doReconcile := func() (reconcile.Result, error) {
		return r.Reconcile(ctx, reconcile.Request{NamespacedName: utils.DefaultEnterpriseInstanceKey})
	}

	tracked := func() []corev1.Secret {
		l := &corev1.SecretList{}
		Expect(cli.List(ctx, l, client.MatchingLabels{SidecarPullSecretLabel: "true"})).NotTo(HaveOccurred())
		return l.Items
	}

	trackedRB := func() []rbacv1.RoleBinding {
		l := &rbacv1.RoleBindingList{}
		Expect(cli.List(ctx, l, client.MatchingLabels{SidecarPullSecretLabel: "true"})).NotTo(HaveOccurred())
		return l.Items
	}

	It("creates nothing when no pull secrets are configured", func() {
		Expect(cli.Create(ctx, installation)).NotTo(HaveOccurred())
		Expect(cli.Create(ctx, al)).NotTo(HaveOccurred())
		createPod("app-1", "user-ns", true)

		_, err := doReconcile()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(tracked()).To(BeEmpty())
	})

	Context("with pull secrets configured and sidecar injection enabled", func() {
		BeforeEach(func() {
			createPullSecret("tigera-pull-secret")
			installation.Spec.ImagePullSecrets = []corev1.LocalObjectReference{{Name: "tigera-pull-secret"}}
			Expect(cli.Create(ctx, installation)).NotTo(HaveOccurred())
			Expect(cli.Create(ctx, al)).NotTo(HaveOccurred())
		})

		It("copies the pull secret into a sidecar pod's namespace", func() {
			createPod("app-1", "user-ns", true)

			_, err := doReconcile()
			Expect(err).ShouldNot(HaveOccurred())

			s := tracked()
			Expect(s).To(HaveLen(1))
			Expect(s[0].Namespace).To(Equal("user-ns"))
			Expect(s[0].Name).To(Equal("tigera-pull-secret"))
			Expect(s[0].Labels[SidecarPullSecretLabel]).To(Equal("true"))

			// The operator SA cannot write a secret into a tenant namespace without
			// a per-namespace RoleBinding to the tigera-operator-secrets ClusterRole;
			// the controller must emit it alongside the copy (no cluster-wide grant).
			rb := trackedRB()
			Expect(rb).To(HaveLen(1))
			Expect(rb[0].Namespace).To(Equal("user-ns"))
			Expect(rb[0].Name).To(Equal(render.TigeraOperatorSecrets))
			Expect(rb[0].RoleRef.Name).To(Equal(render.TigeraOperatorSecrets))
		})

		It("copies once per namespace and to every sidecar namespace", func() {
			createPod("a1", "ns-a", true)
			createPod("a2", "ns-a", true)
			createPod("b1", "ns-b", true)

			_, err := doReconcile()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(tracked()).To(HaveLen(2))
		})

		It("ignores pods without the sidecar label", func() {
			createPod("plain", "other-ns", false)

			_, err := doReconcile()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(tracked()).To(BeEmpty())
		})

		It("does not copy into the operator or calico-system namespaces", func() {
			createPod("op", common.OperatorNamespace(), true)
			createPod("cs", common.CalicoNamespace, true)

			_, err := doReconcile()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(tracked()).To(BeEmpty())
		})

		It("cleans up the copy when the sidecar pod is gone", func() {
			createPod("app-1", "user-ns", true)
			_, err := doReconcile()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(tracked()).To(HaveLen(1))

			Expect(cli.Delete(ctx, &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "app-1", Namespace: "user-ns"}})).NotTo(HaveOccurred())
			_, err = doReconcile()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(tracked()).To(BeEmpty())
			Expect(trackedRB()).To(BeEmpty())
		})

		It("cleans up when sidecar injection is disabled", func() {
			createPod("app-1", "user-ns", true)
			_, err := doReconcile()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(tracked()).To(HaveLen(1))
			Expect(trackedRB()).To(HaveLen(1))

			disabled := operatorv1.SidecarDisabled
			al.Spec.SidecarInjection = &disabled
			Expect(cli.Update(ctx, al)).NotTo(HaveOccurred())

			_, err = doReconcile()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(tracked()).To(BeEmpty())
			Expect(trackedRB()).To(BeEmpty())
		})
	})
})
