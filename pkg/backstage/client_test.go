package backstage

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestClient_ListTeams(t *testing.T) {
	testServer := newFakeAPI(t, "/api/catalog/entities", url.Values{
		"filter": []string{"kind=Group"},
	}, "testdata/groups.json", "Bearer testing")

	c := NewClient(testServer.URL, "testing")

	teams, err := c.ListTeams(context.TODO())
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"team-a", "team-b", "team-c", "team-d"}
	if diff := cmp.Diff(want, teams); diff != "" {
		t.Fatalf("failed to get teams:\n%s", diff)
	}
}

func newFakeAPI(t *testing.T, urlPath string, values url.Values, filename, authToken string) *httptest.Server {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if v := r.Header.Get("Authorization"); v != authToken {
			http.Error(w, fmt.Sprintf("not authorized, got %q, want %q", v, authToken), http.StatusForbidden)
			return
		}
		if r.URL.Path == urlPath && reflect.DeepEqual(r.URL.Query(), values) {
			http.ServeFile(w, r, filename)
			return
		}
		t.Logf("URL Path = %v, query = %v", r.URL.Path, r.URL.Query())
		http.Error(w, fmt.Sprintf("%q not found", r.URL.Path), http.StatusNotFound)
	}))
	t.Cleanup(s.Close)
	return s
}
