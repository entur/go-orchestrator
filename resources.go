package orchestrator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// See https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body
func request(ctx context.Context, client *http.Client, method string, url string, headers map[string]string, reqBody any, resBody any) (int, error) {
	enc, err := json.Marshal(reqBody)
	if err != nil {
		return 0, fmt.Errorf("http request body couldn't marshal correctly: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(enc))
	if err != nil {
		return 0, fmt.Errorf("http '%s' request preparation to '%s' failed: %w", method, url, err)
	}
	for key, val := range headers {
		req.Header.Set(key, val)
	}

	res, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("http '%s' request to '%s' failed: %w", method, url, err)
	}
	defer res.Body.Close()

	dec := json.NewDecoder(res.Body)
	err = dec.Decode(resBody)
	if err != nil {
		return res.StatusCode, fmt.Errorf("http '%s' response  body from '%s' was unable to be read: %w", method, url, err)
	}

	return res.StatusCode, nil
}

type IAMLookupClient struct {
	client *http.Client
	url    string
}

func (iam *IAMLookupClient) AppProjectIDS(ctx context.Context, appID string) ([]string, error) {
	type AppProjectsRequest struct {
		AppID string `json:"appId"`
	}

	type AppProjectsResponse struct {
		ProjectIDS []string `json:"projects"`
	}

	url := fmt.Sprintf("%s/app/projects/gcp", iam.url)
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	reqBody := AppProjectsRequest{
		AppID: appID,
	}
	resBody := AppProjectsResponse{}

	status, err := request(ctx, iam.client, http.MethodPost, url, headers, reqBody, &resBody)
	if err != nil {
		return nil, err
	}
	if status != 200 && status != 404 {
		return nil, err
	}

	return resBody.ProjectIDS, nil
}

func (iam *IAMLookupClient) UserHasRoleOnProjects(ctx context.Context, email string, role string, ProjectIDS ...string) (bool, error) {
	type UserAccessRequest struct {
		User     string `json:"user"`
		Role     string `json:"role"`
		Resource string `json:"resource"`
	}

	type UserAccessResponse struct {
		HasAccess bool `json:"access"`
	}

	url := fmt.Sprintf("%s/access/gcp", iam.url)
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	reqBody := UserAccessRequest{
		User: email,
		Role: role,
	}
	resBody := UserAccessResponse{}

	for i := range ProjectIDS {
		reqBody.Resource = fmt.Sprintf("projects/%s", ProjectIDS[i])

		status, err := request(ctx, iam.client, http.MethodPost, url, headers, reqBody, &resBody)
		if err != nil {
			return false, err
		}
		if status != 200 {
			return false, nil
		}
	}

	return true, nil
}

func (iam *IAMLookupClient) UserGroups(ctx context.Context, email string) ([]string, error) {
	type UserGroupsRequest struct {
		User string `json:"user"`
	}

	type UserGroupsResponse struct {
		Groups []string `json:"groups"`
	}

	url := fmt.Sprintf("%s/groups/entraid", iam.url)
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	reqBody := UserGroupsRequest{
		User: email,
	}
	resBody := UserGroupsResponse{}

	status, err := request(ctx, iam.client, http.MethodPost, url, headers, reqBody, &resBody)
	if err != nil {
		return nil, err
	}
	if status != 200 {
		return nil, err
	}

	return resBody.Groups, nil
}

func NewIAMLookupClient(client *http.Client, url string) IAMLookupClient {
	return IAMLookupClient{
		client: client,
		url:    url,
	}
}
