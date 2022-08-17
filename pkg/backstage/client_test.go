// Simple Backstage.io client for fetching entity data.
package backstage

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/gitops-tools/backstage-tenants-controller/test"
	"github.com/google/go-cmp/cmp"
)

func TestClient_ListTeams(t *testing.T) {
	testServer := newFakeAPI(t, "/api/catalog/entities", url.Values{
		"filter": []string{"kind=Group"},
	}, "testdata/groups.json", "Bearer testing", "")
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

func TestClient_sets_etag(t *testing.T) {
	testEtag := `"W/\"3f1b-lVgqz+2vTy1JnXsjdTi/dLfvQVE\""`
	testServer := newFakeAPI(t, "/api/catalog/entities", url.Values{
		"filter": []string{"kind=Group"},
	}, "testdata/groups.json", "Bearer testing", testEtag)
	c := NewClient(testServer.URL, "testing")

	_, err := c.ListTeams(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	if c.LastEtag != testEtag {
		t.Fatalf("got etag %q, want %q", c.LastEtag, testEtag)
	}
}

func TestClient_sending_etag(t *testing.T) {
	testEtag := `"W/\"3f1b-lVgqz+2vTy1JnXsjdTi/dLfvQVE\""`
	testServer := newFakeAPI(t, "/api/catalog/entities", url.Values{
		"filter": []string{"kind=Group"},
	}, "testdata/groups.json", "Bearer testing", testEtag)
	c := NewClient(testServer.URL, "testing")
	c.LastEtag = testEtag

	teams, err := c.ListTeams(context.TODO())
	if err != nil {
		t.Fatal(err)
	}
	if teams != nil {
		t.Fatalf("expected no teams when etag matched, got %v", teams)
	}

}

func TestClient_unauthenticated(t *testing.T) {
	testServer := newFakeAPI(t, "/api/catalog/entities", url.Values{
		"filter": []string{"kind=Group"},
	}, "testdata/groups.json", "", "")
	c := NewClient(testServer.URL, "")

	teams, err := c.ListTeams(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"team-a", "team-b", "team-c", "team-d"}
	if diff := cmp.Diff(want, teams); diff != "" {
		t.Fatalf("failed to get teams:\n%s", diff)
	}
}

func TestClient_unauthorized(t *testing.T) {
	testServer := newFakeAPI(t, "/api/catalog/entities", url.Values{
		"filter": []string{"kind=Group"},
	}, "testdata/groups.json", "Bearer password", "")
	c := NewClient(testServer.URL, "")

	_, err := c.ListTeams(context.TODO())
	// TODO: change this to a specific error type.
	test.AssertErrorMatch(t, "unexpected response status 403", err)
}

func TestClient_bad_base_url(t *testing.T) {
	c := NewClient("%%", "")

	_, err := c.ListTeams(context.TODO())
	test.AssertErrorMatch(t, `calculating API URL: parsing Backstage API base "%%"`, err)
}

func TestClient_invalid_base_url(t *testing.T) {
	c := NewClient("http:///localhost:9000", "")

	_, err := c.ListTeams(context.TODO())
	test.AssertErrorMatch(t, `failed to execute request: Get "http://`, err)
}

func TestClient_bad_content(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("No Content"))
	}))

	c := NewClient(s.URL, "")

	_, err := c.ListTeams(context.TODO())
	test.AssertErrorMatch(t, `unexpected Content-Type "text/plain"`, err)
}

func TestClient_bad_response(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("No Content"))
	}))

	c := NewClient(s.URL, "")

	_, err := c.ListTeams(context.TODO())
	test.AssertErrorMatch(t, `parsing response body: invalid character 'N'`, err)
}

func TestClient_timeout(t *testing.T) {
	// https://github.com/google/go-github/blob/f2d99f17ead8dd906d8598ac43f99996b647a614/github/github.go#L647
}

func newFakeAPI(t *testing.T, urlPath string, values url.Values, filename, authToken, etag string) *httptest.Server {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if v := r.Header.Get("Authorization"); v != authToken {
			http.Error(w, fmt.Sprintf("not authorized, got %q, want %q", v, authToken), http.StatusForbidden)
			return
		}
		if v := r.Header.Get("If-None-Match"); v != "" {
			if v != etag {
				http.Error(w, fmt.Sprintf("invalid Etag %q", v), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNotModified)
			return
		}
		if etag != "" {
			w.Header().Set("Etag", etag)
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
