package tenancy

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gitops-tools/backstage-tenants-controller/pkg/backstage"
	"github.com/gitops-tools/pkg/sets"
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
// TODO: logger!
func ReconcileServiceAccounts(ctx context.Context, cl client.Client, teams []backstage.Team) error {
	existingSAs, err := existingTeamServiceAccounts(ctx, cl)
	if err != nil {
		return err
	}

	newSAs := []corev1.ServiceAccount{}
	for i := range teams {
		sa := corev1.ServiceAccount{
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
		newSAs = append(newSAs, sa)
	}

	for _, sa := range newSAs {
		existing := &corev1.ServiceAccount{}
		err := cl.Get(ctx, client.ObjectKeyFromObject(&sa), existing)
		if err != nil {
			if !errors.IsNotFound(err) {
				return fmt.Errorf("checking for existing ServiceAccount: %w", err)
			}
			if err := cl.Create(ctx, &sa); err != nil {
				return fmt.Errorf("creating new ServiceAccount: %w", err)
			}
			continue
		}

		if !equality.Semantic.DeepDerivative(sa.GetLabels(), existing.GetLabels()) {
			existing.SetLabels(sa.GetLabels())
			if err := cl.Update(ctx, existing); err != nil {
				return fmt.Errorf("updating existing ServiceAccount: %w", err)
			}
		}
	}

	namesToRemove := setFromServiceAccounts(existingSAs).Difference(setFromServiceAccounts(newSAs))
	for _, sa := range namesToRemove.List() {
		if err := cl.Delete(ctx, &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: sa.Name, Namespace: sa.Namespace}}); err != nil {
			return fmt.Errorf("pruning ServiceAccount: %w", err)
		}
	}

	return nil
}

func setFromServiceAccounts(sas []corev1.ServiceAccount) sets.Set[types.NamespacedName] {
	nameSet := sets.New[types.NamespacedName]()
	for _, sa := range sas {
		nameSet.Insert(client.ObjectKeyFromObject(&sa))
	}

	return nameSet
}

func existingTeamServiceAccounts(ctx context.Context, cl client.Client) ([]corev1.ServiceAccount, error) {
	existingSAs := &corev1.ServiceAccountList{}
	if err := cl.List(context.TODO(), existingSAs, client.HasLabels([]string{"tenants.gitops.pro/team"})); err != nil {
		return nil, fmt.Errorf("listing existing ServiceAccounts: %w", err)
	}

	return existingSAs.Items, nil
}
