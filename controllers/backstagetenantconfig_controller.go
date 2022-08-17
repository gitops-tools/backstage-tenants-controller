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
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	tenantsv1 "github.com/gitops-tools/backstage-tenants-controller/api/v1alpha1"
	"github.com/gitops-tools/backstage-tenants-controller/pkg/backstage"
)

// BackstageTenantConfigReconciler reconciles a BackstageTenantConfig object
type BackstageTenantConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=tenants.gitops.pro,resources=backstagetenantconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tenants.gitops.pro,resources=backstagetenantconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tenants.gitops.pro,resources=backstagetenantconfigs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *BackstageTenantConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	// TODO: metrics!

	cfg := &tenantsv1.BackstageTenantConfig{}
	if err := r.Get(ctx, req.NamespacedName, cfg); err != nil {
		// TODO: check for deleted etc.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// TODO: Auth if available!
	bc := backstage.NewClient(cfg.Spec.BaseURL, "")
	names, err := bc.ListTeams(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}

	// TODO: Flux patchHelper?
	patch, err := json.Marshal(map[string]interface{}{
		"status": map[string]interface{}{
			"teamNames": names,
			"lastEtag":  bc.LastEtag,
		},
	})
	if err != nil {
		return ctrl.Result{}, err
	}
	if err := r.Status().Patch(ctx, cfg, client.RawPatch(types.MergePatchType, patch)); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to update config: %w", err)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *BackstageTenantConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tenantsv1.BackstageTenantConfig{}).
		Complete(r)
}
