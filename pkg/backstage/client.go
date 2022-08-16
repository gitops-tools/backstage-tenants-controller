package backstage

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
)

// Client is a Backstage client for querying entities.
type Client struct {
	BaseURL   string
	AuthToken string
	client    *http.Client
}

// NewClient creates and returns a client ready for use.
func NewClient(baseURL, auth string) *Client {
	return &Client{
		BaseURL:   baseURL,
		AuthToken: auth,
		client:    http.DefaultClient,
	}
}

// https://demo.backstage.io/api/catalog/entities?filter=kind=group
func (c Client) ListTeams(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/api/catalog/entities?filter=kind=Group", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for URL %q: %w", c.BaseURL, err)
	}
	res, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer res.Body.Close()

	if h := res.Header.Get("Content-Type"); h != "application/json" {
		return nil, fmt.Errorf("did not get JSON response: %q", h)
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	entities := []entity{}
	err = json.Unmarshal(b, &entities)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response body: %w", err)
	}
	teams := []string{}
	for _, v := range entities {
		var spec teamSpec
		if err := json.Unmarshal(v.Spec, &spec); err != nil {
			return nil, fmt.Errorf("failed to parse entity %s: %w", v.Metadata.Name, err)
		}
		if spec.Type == "team" {
			teams = append(teams, v.Metadata.Name)
		}
	}
	sort.Strings(teams)

	return teams, nil
}

type teamSpec struct {
	Type string `json:"type"`
}

type entity struct {
	Metadata struct {
		Name        string `json:"name"`
		Namespace   string `json:"namespace"`
		Description string `json:"description"`
	} `json:"metadata"`
	Spec json.RawMessage
}
