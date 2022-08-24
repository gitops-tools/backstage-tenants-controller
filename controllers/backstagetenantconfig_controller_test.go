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
	"path/filepath"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	tenantsv1 "github.com/gitops-tools/backstage-tenants-controller/api/v1alpha1"
	"github.com/gitops-tools/backstage-tenants-controller/pkg/backstage"
	"github.com/google/go-cmp/cmp"
)

func TestAPIs(t *testing.T) {
	testEnv := &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	cfg, err := testEnv.Start()
	if err != nil {
		t.Fatal(err)
	}

	if err = tenantsv1.AddToScheme(scheme.Scheme); err != nil {
		t.Fatal(err)
	}

	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme.Scheme})
	if err != nil {
		t.Fatal(err)
	}

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{Scheme: scheme.Scheme})
	if err != nil {
		t.Fatal(err)
	}
	reconciler := &BackstageTenantConfigReconciler{
		Client: k8sClient,
		Scheme: scheme.Scheme,
	}
	if err := reconciler.SetupWithManager(mgr); err != nil {
		t.Fatal(err)
	}

	t.Run("successfully querying the API", func(t *testing.T) {
		ctx := context.TODO()
		cfg := newTestConfig()
		if err := k8sClient.Create(ctx, cfg); err != nil {
			t.Fatal(err)
		}
		defer cleanupResource(t, k8sClient, cfg)

		res, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cfg)})
		if err != nil {
			t.Fatal(err)
		}

		if res.RequeueAfter != cfg.Spec.Interval.Duration {
			t.Fatalf("got RequeueAfter %v, want %v", res.RequeueAfter, cfg.Spec.Interval)
		}

		updated := &tenantsv1.BackstageTenantConfig{}
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(cfg), updated); err != nil {
			t.Fatal(err)
		}
		want := []string{"team-a", "team-b", "team-c", "team-d"}
		if diff := cmp.Diff(want, updated.Status.TeamNames); diff != "" {
			t.Fatalf("team names not loaded:\n%s", diff)
		}
		// Because this is hitting Backstage directly, we can't really know what
		// the current Etag is.
		if updated.Status.LastEtag == "" {
			t.Fatal("expected Status.LastEtag to not be empty")
		}
	})

	t.Run("successfully reconciling Namespaces", func(t *testing.T) {
		ctx := context.TODO()
		cfg := newTestConfig()
		if err := k8sClient.Create(ctx, cfg); err != nil {
			t.Fatal(err)
		}
		defer cleanupResource(t, k8sClient, cfg)

		res, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cfg)})
		if err != nil {
			t.Fatal(err)
		}

		if res.RequeueAfter != cfg.Spec.Interval.Duration {
			t.Fatalf("got RequeueAfter %v, want %v", res.RequeueAfter, cfg.Spec.Interval)
		}

		updated := &tenantsv1.BackstageTenantConfig{}
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(cfg), updated); err != nil {
			t.Fatal(err)
		}
		nsList := &corev1.NamespaceList{}
		if err := k8sClient.List(ctx, nsList, client.HasLabels([]string{"tenants.gitops.pro/team"})); err != nil {
			t.Fatal(err)
		}
		nsNames := []string{}
		for _, v := range nsList.Items {
			nsNames = append(nsNames, v.GetName())
		}
		for _, v := range updated.Status.TeamNames {
			if !sliceContains(v, nsNames) {
				t.Errorf("name %v, namespace missing", v)
			}
		}
	})

	t.Run("successfully reconciling ServiceAccounts", func(t *testing.T) {
		ctx := context.TODO()
		cfg := newTestConfig()
		if err := k8sClient.Create(ctx, cfg); err != nil {
			t.Fatal(err)
		}
		defer cleanupResource(t, k8sClient, cfg)

		res, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cfg)})
		if err != nil {
			t.Fatal(err)
		}

		if res.RequeueAfter != cfg.Spec.Interval.Duration {
			t.Fatalf("got RequeueAfter %v, want %v", res.RequeueAfter, cfg.Spec.Interval)
		}

		updated := &tenantsv1.BackstageTenantConfig{}
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(cfg), updated); err != nil {
			t.Fatal(err)
		}
		saList := &corev1.ServiceAccountList{}
		if err := k8sClient.List(ctx, saList, client.HasLabels([]string{"tenants.gitops.pro/team"})); err != nil {
			t.Fatal(err)
		}
		accountNames := []string{}
		for _, v := range saList.Items {
			accountNames = append(accountNames, v.GetName())
		}
		for _, v := range updated.Status.TeamNames {
			if !sliceContains(v, accountNames) {
				t.Errorf("name %v, service account missing", v)
			}
		}
	})

	t.Run("updating the inventory of created resources", func(t *testing.T) {
		ctx := context.TODO()
		cfg := newTestConfig()
		if err := k8sClient.Create(ctx, cfg); err != nil {
			t.Fatal(err)
		}
		defer cleanupResource(t, k8sClient, cfg)

		res, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cfg)})
		if err != nil {
			t.Fatal(err)
		}

		if res.RequeueAfter != cfg.Spec.Interval.Duration {
			t.Fatalf("got RequeueAfter %v, want %v", res.RequeueAfter, cfg.Spec.Interval)
		}

		updated := &tenantsv1.BackstageTenantConfig{}
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(cfg), updated); err != nil {
			t.Fatal(err)
		}
		if l := len(updated.Status.TenantInventory); l == 0 {
			t.Fatalf("got no tenant inventory items")
		}
	})

	t.Run("querying with current Etag", func(t *testing.T) {
		ctx := context.TODO()
		cfg := newTestConfig()
		if err := k8sClient.Create(ctx, cfg); err != nil {
			t.Fatal(err)
		}
		defer cleanupResource(t, k8sClient, cfg)

		// initialise the config with the current etag for the data.
		// theoretically, this could fail if there was a change to the upstream
		// data between here and the reconcile call.
		bc := backstage.NewClient(cfg.Spec.BaseURL, "")
		_, err := bc.ListTeams(ctx)
		cfg.Status.LastEtag = bc.LastEtag
		if err := k8sClient.Status().Update(ctx, cfg); err != nil {
			t.Fatal(err)
		}

		res, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cfg)})
		if err != nil {
			t.Fatal(err)
		}

		if res.RequeueAfter != cfg.Spec.Interval.Duration {
			t.Fatalf("got RequeueAfter %v, want %v", res.RequeueAfter, cfg.Spec.Interval)
		}

		// Because the Etag matches, we get no teams and so the teams don't get
		// stored.
		updated := &tenantsv1.BackstageTenantConfig{}
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(cfg), updated); err != nil {
			t.Fatal(err)
		}
		if len(updated.Status.TeamNames) != 0 {
			t.Fatalf("want no teams, got %v", updated.Status.TeamNames)
		}
	})

	if err := testEnv.Stop(); err != nil {
		t.Fatal(err)
	}
}

func cleanupResource(t *testing.T, cl client.Client, obj client.Object) {
	if err := cl.Delete(context.TODO(), obj); err != nil {
		t.Fatal(err)
	}
}

func newTestConfig() *tenantsv1.BackstageTenantConfig {
	return &tenantsv1.BackstageTenantConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testing",
			Namespace: "default",
		},
		Spec: tenantsv1.BackstageTenantConfigSpec{
			BaseURL:  "https://demo.backstage.io/",
			Interval: metav1.Duration{Duration: 5 * time.Second},
		},
	}
}

func sliceContains[T comparable](want T, elements []T) bool {
	for _, v := range elements {
		if v == want {
			return true
		}
	}
	return false
}
