package tenancy

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/gitops-tools/backstage-tenants-controller/pkg/backstage"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestReconcileServiceAccounts(t *testing.T) {
	cl := newFakeClient(t)
	teams := []backstage.Team{
		{
			Name:      "team-a",
			Namespace: "default",
		},
	}

	err := ReconcileServiceAccounts(context.TODO(), cl, teams)
	if err != nil {
		t.Fatal(err)
	}
	want := &corev1.ServiceAccount{
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
	loaded := &corev1.ServiceAccount{}
	if err := cl.Get(context.TODO(), types.NamespacedName{Name: "team-a", Namespace: "team-a"}, loaded); err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(want, loaded); diff != "" {
		t.Errorf("didn't create service account correctly:\n%s", diff)
	}
}

func TestReconcileServiceAccounts_preexisting(t *testing.T) {
	existing := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
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

	err := ReconcileServiceAccounts(context.TODO(), cl, teams)
	if err != nil {
		t.Fatal(err)
	}

	want := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
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

	updated := &corev1.ServiceAccount{}
	if err := cl.Get(context.TODO(), client.ObjectKeyFromObject(existing), updated); err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(want, updated, ignoreResourceVersion()); diff != "" {
		t.Errorf("failed to update ServiceAccount:\n%s", diff)
	}
}

func TestReconcileServiceAccounts_pruning(t *testing.T) {
	existing := &corev1.ServiceAccount{
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

	err := ReconcileServiceAccounts(context.TODO(), cl, teams)
	if err != nil {
		t.Fatal(err)
	}

	saList := &corev1.ServiceAccountList{}
	if err := cl.List(context.TODO(), saList, client.HasLabels([]string{"tenants.gitops.pro/team"})); err != nil {
		t.Fatal(err)
	}

	if l := len(saList.Items); l != 1 {
		t.Errorf("got %d ServiceAccounts, want 1 (%v)", l, resourceNames(t, saList))
	}
}

func resourceNames(t *testing.T, list runtime.Object) []string {
	t.Helper()
	names := []string{}
	err := meta.EachListItem(list, func(obj runtime.Object) error {
		names = append(names, obj.(client.Object).GetName())
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	return names
}

func newFakeClient(t *testing.T, objs ...runtime.Object) client.Client {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(objs...).
		Build()
}

func ignoreResourceVersion() cmp.Option {
	return cmpopts.IgnoreFields(metav1.ObjectMeta{}, "ResourceVersion")
}
