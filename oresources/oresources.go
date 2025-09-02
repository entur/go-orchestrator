package oresources

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/entur/go-logging"
	"google.golang.org/api/idtoken"
)

// -----------------------
// Helpers
// -----------------------

// See https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body
func request(ctx context.Context, client *http.Client, method string, url string, headers map[string]string, reqBody any, resBody any) (int, error) {
	if client == nil {
		return http.StatusInternalServerError, fmt.Errorf("no client passed to request")
	}
	enc, err := json.Marshal(reqBody)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("http '%s' request body failed to marshal: %w", method, err)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(enc))
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("http '%s' request preparation failed: %w", method, err)
	}

	req.Header.Set("Content-Type", "application/json")
	for key, val := range headers {
		req.Header.Set(key, val)
	}

	res, err := client.Do(req)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("http '%s' request failed: %w", method, err)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.Header.Get("Content-Type") == "application/json" {
		dec := json.NewDecoder(res.Body)
		err = dec.Decode(resBody)
		if err != nil && err != io.EOF {
			return res.StatusCode, fmt.Errorf("http '%s' response failed to decode: %w", method, err)
		}
	}

	return res.StatusCode, nil
}

// -----------------------
// Resource Clients
// -----------------------

const defaultTimeout = 10 * time.Second
const defaultDialerTimeout = 5 * time.Second

type IAMClient struct {
	client *http.Client
	url    string
}

type GCPAppProjectsRequest struct {
	AppID string `json:"appId"`
}

type GCPAppProjectsResponse struct {
	ProjectIDs []string `json:"projects"`
}

// List all of the GCP project ids associated with an app-factory id.
func (iam *IAMClient) GCPAppProjectIDs(ctx context.Context, appID string) ([]string, error) {
	url := fmt.Sprintf("%s/app/projects/gcp", iam.url)
	reqBody := GCPAppProjectsRequest{
		AppID: appID,
	}
	resBody := GCPAppProjectsResponse{}

	status, err := request(ctx, iam.client, http.MethodPost, url, nil, reqBody, &resBody)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK && status != http.StatusNotFound {
		return nil, err
	}

	return resBody.ProjectIDs, nil
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
func (iam *IAMClient) GCPUserHasRoleInProjects(ctx context.Context, email string, role string, projectIDs ...string) (bool, error) {
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
		if status != http.StatusOK {
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
func (iam *IAMClient) EntraIDUserGroups(ctx context.Context, email string) ([]string, error) {
	url := fmt.Sprintf("%s/groups/entraid", iam.url)
	reqBody := EntraIDUserGroupsRequest{
		User: email,
	}
	resBody := EntraIDUserGroupsResponse{}

	status, err := request(ctx, iam.client, http.MethodPost, url, nil, reqBody, &resBody)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, err
	}

	return resBody.Groups, nil
}

type IAMClientOption = idtoken.ClientOption

// NewIAMClient returns a http client which can be used against the IAM Lookup Resource.
// It can also be used along with NewMockIAMServer for local client -> server testing.
func NewIAMClient(ctx context.Context, url string, opts ...IAMClientOption) (*IAMClient, error) {
	logger := logging.Ctx(ctx)

	client, err := idtoken.NewClient(ctx, url, opts...)
	if err != nil {
		errStr := err.Error()
		if !strings.HasPrefix(errStr, "idtoken: unsupported credentials type") && !strings.HasPrefix(errStr, "google: could not find default credentials") {
			return nil, fmt.Errorf("unable to create iam client: %w", err)
		}

		logger.Debug().Msg("Unable to discover idtoken credentials, defaulting to http.Client for IAM")
		client = &http.Client{
			Timeout: defaultTimeout,
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout: defaultDialerTimeout,
				}).Dial,
			},
		}
	}

	return &IAMClient{
		client: client,
		url:    url,
	}, nil
}
