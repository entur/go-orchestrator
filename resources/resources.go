package resources

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"google.golang.org/api/idtoken"
)

// -----------------------
// Helpers
// -----------------------

// See https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body
func request(ctx context.Context, client *http.Client, method string, url string, headers map[string]string, reqBody any, resBody any) (int, error) {
	if client == nil {
		return 0, fmt.Errorf("no client passed to request")
	}
	enc, err := json.Marshal(reqBody)
	if err != nil {
		return 0, fmt.Errorf("http request body failed to marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(enc))
	if err != nil {
		return 0, fmt.Errorf("http '%s' request preparation failed: %w", method, err)
	}

	req.Header.Set("Content-Type", "application/json")
	for key, val := range headers {
		req.Header.Set(key, val)
	}

	res, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("http '%s' request failed: %w", method, err)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	dec := json.NewDecoder(res.Body)
	err = dec.Decode(resBody)
	if err != nil {
		return res.StatusCode, fmt.Errorf("http '%s' response failed to read: %w", method, err)
	}

	return res.StatusCode, nil
}

// -----------------------
// Resource Clients
// -----------------------

type IAMLookupClient struct {
	client *http.Client
	url    string
}

type GCPAppProjectsRequest struct {
	AppID string `json:"appId"`
}

type GCPAppProjectsResponse struct {
	ProjectIDS []string `json:"projects"`
}

// List all of the GCP project ids associated with an app-factory id.
func (iam *IAMLookupClient) GCPAppProjectIDS(ctx context.Context, appID string) ([]string, error) {
	url := fmt.Sprintf("%s/app/projects/gcp", iam.url)
	reqBody := GCPAppProjectsRequest{
		AppID: appID,
	}
	resBody := GCPAppProjectsResponse{}

	status, err := request(ctx, iam.client, http.MethodPost, url, nil, reqBody, &resBody)
	if err != nil {
		return nil, err
	}
	if status != 200 && status != 404 {
		return nil, err
	}

	return resBody.ProjectIDS, nil
}

type GCPUserAccessRequest struct {
	User     string `json:"user"`
	Role     string `json:"role"`
	Resource string `json:"resource"`
}

type GCPUserAccessResponse struct {
	HasAccess bool `json:"access"`
}

// Check if the user (email) has the specified Sub-Orchestrator role in *all* of the given GCP projects.
func (iam *IAMLookupClient) GCPUserHasRoleInProjects(ctx context.Context, email string, role string, projectIDs ...string) (bool, error) {
	url := fmt.Sprintf("%s/access/gcp", iam.url)
	reqBody := GCPUserAccessRequest{
		User: email,
		Role: role,
	}
	resBody := GCPUserAccessResponse{}

	for i := range projectIDs {
		reqBody.Resource = fmt.Sprintf("projects/%s", projectIDs[i])

		status, err := request(ctx, iam.client, http.MethodPost, url, nil, reqBody, &resBody)
		if err != nil {
			return false, err
		}
		if status != 200 {
			return false, nil
		}
	}

	return true, nil
}

type EntraIDUserGroupsRequest struct {
	User string `json:"user"`
}

type EntraIDUserGroupsResponse struct {
	Groups []string `json:"groups"`
}

// List all of the entra id groups (without the @ suffix) that a user (email) belongs to.
func (iam *IAMLookupClient) EntraIDUserGroups(ctx context.Context, email string) ([]string, error) {
	url := fmt.Sprintf("%s/groups/entraid", iam.url)
	reqBody := EntraIDUserGroupsRequest{
		User: email,
	}
	resBody := EntraIDUserGroupsResponse{}

	status, err := request(ctx, iam.client, http.MethodPost, url, nil, reqBody, &resBody)
	if err != nil {
		return nil, err
	}
	if status != 200 {
		return nil, err
	}

	return resBody.Groups, nil
}

type IAMLookupClientOptions = idtoken.ClientOption

func NewIAMLookupClient(url string, options ...IAMLookupClientOptions) IAMLookupClient {
	client, err := idtoken.NewClient(context.Background(), url, options...)
	if err != nil && err.Error() == "idtoken: unsupported credentials type" {
		client = http.DefaultClient
	}

	return IAMLookupClient{
		client: client,
		url:    url,
	}
}
