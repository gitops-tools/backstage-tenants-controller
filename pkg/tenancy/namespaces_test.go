package tenancy

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gitops-tools/backstage-tenants-controller/pkg/backstage"
	"github.com/google/go-cmp/cmp"
)

func TestReconcileNamespaces(t *testing.T) {
	cl := newFakeClient(t)
	teams := []backstage.Team{
		{
			Name:      "team-a",
			Namespace: "default",
		},
	}

	err := ReconcileNamespaces(context.TODO(), cl, teams)
	if err != nil {
		t.Fatal(err)
	}

	nsList := &corev1.NamespaceList{}
	if err := cl.List(context.TODO(), nsList, client.HasLabels([]string{"tenants.gitops.pro/team"})); err != nil {
		t.Fatal(err)
	}

	want := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "team-a",
			Namespace:       "team-a",
			ResourceVersion: "1",
			Labels: map[string]string{
				"app.kubernetes.io/created-by": "backstage-tenants-controller",
				"app.kubernetes.io/managed-by": "backstage",
				"tenants.gitops.pro/team":      "team-a",
			},
		},
	}

	if diff := cmp.Diff(want, nsList.Items[0]); diff != "" {
		t.Fatalf("didn't create service account correctly:\n%s", diff)
	}
}

func TestReconcileNamespaces_preexisting(t *testing.T) {
	existing := &corev1.Namespace{
		// No labels on the pre-existing one.
		ObjectMeta: metav1.ObjectMeta{
			Name:      "team-b",
			Namespace: "team-b",
		},
	}
	cl := newFakeClient(t, existing)
	teams := []backstage.Team{
		{
			Name:      "team-b",
			Namespace: "default",
		},
	}

	err := ReconcileNamespaces(context.TODO(), cl, teams)
	if err != nil {
		t.Fatal(err)
	}

	want := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "team-b",
			Namespace: "team-b",
			Labels: map[string]string{
				"app.kubernetes.io/created-by": "backstage-tenants-controller",
				"app.kubernetes.io/managed-by": "backstage",
				"tenants.gitops.pro/team":      "team-b",
			},
		},
	}

	updated := &corev1.Namespace{}
	if err := cl.Get(context.TODO(), client.ObjectKeyFromObject(existing), updated); err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(want, updated, ignoreResourceVersion()); diff != "" {
		t.Fatalf("failed to update Namespace:\n%s", diff)
	}
}

func TestReconcileNamespaces_pruning(t *testing.T) {
	existing := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "team-b",
			Namespace: "team-b",
			Labels: map[string]string{
				"app.kubernetes.io/created-by": "backstage-tenants-controller",
				"app.kubernetes.io/managed-by": "backstage",
				"tenants.gitops.pro/team":      "team-b",
			},
		},
	}
	cl := newFakeClient(t, existing)
	teams := []backstage.Team{
		{
			Name:      "team-a",
			Namespace: "default",
		},
	}

	err := ReconcileNamespaces(context.TODO(), cl, teams)
	if err != nil {
		t.Fatal(err)
	}

	nsList := &corev1.NamespaceList{}
	if err := cl.List(context.TODO(), nsList, client.HasLabels([]string{"tenants.gitops.pro/team"})); err != nil {
		t.Fatal(err)
	}

	if l := len(nsList.Items); l != 1 {
		t.Fatalf("got %d Namespaces, want 1 (%v)", l, resourceNames(t, nsList))
	}
}
