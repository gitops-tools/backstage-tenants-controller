package backstage

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestClient_ListTeams(t *testing.T) {
	testServer := newFakeAPI(t)

	c := NewClient(testServer.URL, "test")

	teams, err := c.ListTeams(context.TODO())
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"team-a", "team-b", "team-c", "team-d"}
	if diff := cmp.Diff(want, teams); diff != "" {
		t.Fatalf("failed to get teams:\n%s", diff)
	}
}

func newFakeAPI(t *testing.T) *httptest.Server {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/catalog/entities" {
			// w.Header().Add("Content-Type", "application/json")
			http.ServeFile(w, r, "testdata/groups.json")
			return
		}
		t.Logf("URL Path = %s", r.URL.Path)
		http.Error(w, fmt.Sprintf("%q not found", r.URL.Path), http.StatusNotFound)
	}))
	t.Cleanup(s.Close)
	return s
}
