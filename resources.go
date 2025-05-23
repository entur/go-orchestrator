package orchestrator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime"
	"net"
	"net/http"
	"strings"

	"google.golang.org/api/idtoken"
)

// -----------------------
// Helpers
// -----------------------

func enforceJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentT := r.Header.Get("Content-Type")
		if contentT != "" {
			mt, _, err := mime.ParseMediaType(contentT)
			if err != nil {
				http.Error(w, "Malformed Content-Type header", http.StatusBadRequest)
				return
			}

			if mt != "application/json" {
				http.Error(w, "Content-Type header must be application/json", http.StatusUnsupportedMediaType)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

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

type MockIAMLookupServer struct {
	server           *http.Server
	port             int
	up               bool
	appIDProjects    map[string][]string
	userProjectRoles map[string]map[string][]string
	userGroups       map[string][]string
}

func (s *MockIAMLookupServer) hGCPProjectIDS(w http.ResponseWriter, req *http.Request) {
	var reqBody GCPAppProjectsRequest
	err := json.NewDecoder(req.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	var resBody GCPAppProjectsResponse

	projectIDS, ok := s.appIDProjects[reqBody.AppID]
	resBody.ProjectIDS = projectIDS
	if !ok {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(resBody)
}

func (s *MockIAMLookupServer) hGCPUserHasRoleInProjects(w http.ResponseWriter, req *http.Request) {
	var reqBody GCPUserAccessRequest
	err := json.NewDecoder(req.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(reqBody.Resource, "projects/") {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	reqBody.Resource = strings.TrimPrefix(reqBody.Resource, "projects/")

	w.Header().Set("Content-Type", "application/json")
	var resBody GCPUserAccessResponse

	project, ok := s.userProjectRoles[reqBody.Resource]
	if ok {
		userRoles := project[reqBody.Role]
		for _, role := range userRoles {
			if role == reqBody.Role {
				resBody.HasAccess = true
				break
			}
		}
	}

	json.NewEncoder(w).Encode(resBody)
}

func (s *MockIAMLookupServer) hEntraIDUserGroups(w http.ResponseWriter, req *http.Request) {
	var reqBody EntraIDUserGroupsRequest
	err := json.NewDecoder(req.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	var resBody EntraIDUserGroupsResponse

	resBody.Groups = s.userGroups[reqBody.User]
	json.NewEncoder(w).Encode(resBody)
}

// Non-blocking
func (s *MockIAMLookupServer) Serve() (ResourceIAMLookup, error) {
	var resource ResourceIAMLookup

	if s.up {
		return resource, fmt.Errorf("TODO")
	}

	portStr := ":0"
	if s.port > 0 {
		portStr = fmt.Sprintf(":%d", s.port)
	}

	l, err := net.Listen("tcp", portStr)
	if err != nil {
		return resource, err
	}

	go s.server.Serve(l)

	s.port = l.Addr().(*net.TCPAddr).Port
	resource.Url = fmt.Sprintf("http://localhost:%d", s.port)

	return resource, nil
}

func (s *MockIAMLookupServer) Close() error {
	err := s.server.Close()
	s.up = false
	return err
}

type MockIAMLookupServerOption func(*MockIAMLookupServer)

func WithPort(port int) MockIAMLookupServerOption {
	return func(s *MockIAMLookupServer) {
		s.port = port
	}
}

func WithAppIDProjects(appID string, projectIDS []string) MockIAMLookupServerOption {
	return func(s *MockIAMLookupServer) {
		s.appIDProjects[appID] = projectIDS
	}
}

func WithUserProjectRoles(email string, projectID string, roles []string) MockIAMLookupServerOption {
	return func(s *MockIAMLookupServer) {
		project, ok := s.userProjectRoles[projectID]
		if !ok {
			project = map[string][]string{}
			s.userProjectRoles[projectID] = project
		}

		project[email] = roles
	}
}

func WithUserGroups(email string, groups []string) MockIAMLookupServerOption {
	return func(s *MockIAMLookupServer) {
		s.userGroups[email] = groups
	}
}

func NewMockIAMLookupServer(options ...MockIAMLookupServerOption) *MockIAMLookupServer {
	s := &MockIAMLookupServer{
		appIDProjects:    map[string][]string{},
		userProjectRoles: map[string]map[string][]string{},
		userGroups:       map[string][]string{},
	}

	for _, opt := range options {
		opt(s)
	}

	mux := http.NewServeMux()
	mux.Handle("POST /app/projects/gcp", enforceJSON(http.HandlerFunc(s.hGCPProjectIDS)))
	mux.Handle("POST /access/gcp", enforceJSON(http.HandlerFunc(s.hGCPUserHasRoleInProjects)))
	mux.Handle("POST /groups/entraid", enforceJSON(http.HandlerFunc(s.hEntraIDUserGroups)))

	s.server = &http.Server{
		Handler: mux,
	}

	return s
}
