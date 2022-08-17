package backstage

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"path"
	"sort"
)

// Client is a Backstage client for querying entities.
type Client struct {
	// BaseURL is the base URL for the Backstage API.
	BaseURL string
	// This is a Bearer tokem for when the Backstage requires auth.
	AuthToken string
	// This is updated with the Etag when a request returns an Etag.
	LastEtag string
	client   *http.Client
}

// NewClient creates and returns a client ready for use.
func NewClient(BaseURL, auth string) *Client {
	return &Client{
		BaseURL:   BaseURL,
		AuthToken: auth,
		client:    http.DefaultClient,
	}
}

// ListTeams lists Groups of Type team
//
// https://backstage.io/docs/features/software-catalog/software-catalog-api
//
// https://backstage.io/docs/features/software-catalog/descriptor-format#kind-group
//
// e.g. https://demo.backstage.io/api/catalog/entities?filter=kind=group
func (c *Client) ListTeams(ctx context.Context) ([]string, error) {
	entities, err := c.queryEntities(ctx, map[string]string{
		"kind": "Group",
	})
	if err != nil {
		return nil, err
	}
	if entities == nil {
		return nil, nil
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

func (c *Client) queryEntities(ctx context.Context, filter map[string]string) ([]entity, error) {
	apiURL, err := entitiesURL(c.BaseURL, filter)
	if err != nil {
		return nil, fmt.Errorf("calculating API URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for URL %q: %w", c.BaseURL, err)
	}
	if c.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.AuthToken)
	}
	if c.LastEtag != "" {
		req.Header.Set("If-None-Match", c.LastEtag)
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer res.Body.Close()

	if c.LastEtag != "" && res.StatusCode == http.StatusNotModified {
		return nil, nil
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status %v", res.StatusCode)
	}

	mediatype, _, err := mime.ParseMediaType(res.Header.Get("Content-Type"))
	if err != nil {
		return nil, fmt.Errorf("parsing Content-Type %q: %w", mediatype, err)
	}
	if mediatype != "application/json" {
		return nil, fmt.Errorf("unexpected Content-Type %q", mediatype)
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}
	entities := []entity{}
	err = json.Unmarshal(b, &entities)
	if err != nil {
		return nil, fmt.Errorf("parsing response body: %w", err)
	}
	c.LastEtag = res.Header.Get("Etag")

	return entities, nil
}

func entitiesURL(base string, filters map[string]string) (string, error) {
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("parsing Backstage API base %q: %w", base, err)
	}
	values := parsed.Query()
	for k, v := range filters {
		values.Add("filter", fmt.Sprintf("%s=%s", k, v))
	}

	parsed.RawQuery = values.Encode()
	parsed.Path = path.Join(parsed.Path, "/api/catalog/entities")

	return parsed.String(), nil
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
