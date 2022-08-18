package controllers

import (
	"context"
	"testing"

	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/gitops-tools/backstage-tenants-controller/pkg/backstage"
)

// ReconcileServiceAccounts ensures there is a ServiceAccount for each team.
//
// ServiceAccounts for removed teams will be removed.
func ReconcileServiceAccounts(ctx context.Context, cl client.Client, teams []backstage.Team) error {

	return nil
}

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
}

// func TestReconcileServiceAccounts_pruning(t *testing.T) {
// 	acct := &corev1.ServiceAccount{
// 		TypeMeta: metav1.TypeMeta{"v1", "ServiceAccount"},
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name: "testing", Namespace: "testing",
// 			Labels: map[string]string{
// 				"gitops.pro/backstage-team": "testing",
// 			},
// 		},
// 	}
// 	cl := newFakeClient(t, acct)
// 	teams := []backstage.Team{
// 		{
// 			Name:      "team-a",
// 			Namespace: "default",
// 		},
// 	}
// }

func newFakeClient(t *testing.T, objs ...runtime.Object) client.Client {
	t.Helper()
	scheme := runtime.NewScheme()
	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(objs...).
		Build()
}
