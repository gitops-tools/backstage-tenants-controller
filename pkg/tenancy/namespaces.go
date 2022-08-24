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

// ReconcileNamespaces ensures there is a Namespace for each team.
//
// Namespaces for removed teams will be removed.
// TODO: logger!
func ReconcileNamespaces(ctx context.Context, cl client.Client, teams []backstage.Team) error {
	existingNSs, err := existingTeamNamespaces(ctx, cl)
	if err != nil {
		return err
	}

	newNSs := []corev1.Namespace{}
	for i := range teams {
		ns := corev1.Namespace{
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
		newNSs = append(newNSs, ns)
	}

	for _, ns := range newNSs {
		existing := &corev1.Namespace{}
		err := cl.Get(ctx, client.ObjectKeyFromObject(&ns), existing)
		if err != nil {
			if !errors.IsNotFound(err) {
				return fmt.Errorf("checking for existing Namespace: %w", err)
			}
			if err := cl.Create(ctx, &ns); err != nil {
				return fmt.Errorf("creating new Namespace: %w", err)
			}
			continue
		}

		if !equality.Semantic.DeepDerivative(ns.GetLabels(), existing.GetLabels()) {
			existing.SetLabels(ns.GetLabels())
			if err := cl.Update(ctx, existing); err != nil {
				return fmt.Errorf("updating existing Namespace: %w", err)
			}
		}
	}

	namesToRemove := setFromNamespaces(existingNSs).Difference(setFromNamespaces(newNSs))
	for _, ns := range namesToRemove.List() {
		if err := cl.Delete(ctx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns.Name, Namespace: ns.Namespace}}); err != nil {
			return fmt.Errorf("pruning Namespace: %w", err)
		}
	}

	return nil
}

func setFromNamespaces(ns []corev1.Namespace) sets.Set[types.NamespacedName] {
	nameSet := sets.New[types.NamespacedName]()
	for _, ns := range ns {
		nameSet.Insert(client.ObjectKeyFromObject(&ns))
	}

	return nameSet
}

func existingTeamNamespaces(ctx context.Context, cl client.Client) ([]corev1.Namespace, error) {
	existingNSs := &corev1.NamespaceList{}
	if err := cl.List(context.TODO(), existingNSs, client.HasLabels([]string{"tenants.gitops.pro/team"})); err != nil {
		return nil, fmt.Errorf("listing existing Namespaces: %w", err)
	}

	return existingNSs.Items, nil

}
