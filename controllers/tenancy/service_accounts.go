package controllers

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gitops-tools/backstage-tenants-controller/pkg/backstage"
)

const (
	managedByLabel  = "app.kubernetes.io/managed-by"
	teamLabel       = "tenants.gitops.pro/team"
	controllerName  = "backstage-tenants-controller"
	controllerLabel = "app.kubernetes.io/created-by"
)

// ReconcileServiceAccounts ensures there is a ServiceAccount for each team.
//
// ServiceAccounts for removed teams will be removed.
func ReconcileServiceAccounts(ctx context.Context, cl client.Client, teams []backstage.Team) error {
	for i := range teams {
		sa := &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      teams[i].Name,
				Namespace: teams[i].Name,
				// TODO: shift this to a function
				Labels: map[string]string{
					managedByLabel:  "backstage",
					teamLabel:       teams[i].Name,
					controllerLabel: controllerName,
				},
			},
		}

		if err := cl.Create(ctx, sa); err != nil {
			return fmt.Errorf("creating ServiceAccount: %w", err)
		}
	}

	return nil
}
