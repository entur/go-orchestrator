package oresources

import (
	"encoding/json"
	"fmt"
	"mime"
	"net"
	"net/http"
	"strings"
	"time"
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

// -----------------------
// Resource Servers
// -----------------------

const defaultReadHeaderTimeout = 10 * time.Second

type MockIAMServer struct {
	server           *http.Server
	port             int
	url              string
	appIDProjects    map[string][]string
	userProjectRoles map[string]map[string][]string
	userGroups       map[string][]string
}

func (s *MockIAMServer) hGCPProjectIDS(w http.ResponseWriter, req *http.Request) {
	var reqBody GCPAppProjectsRequest
	err := json.NewDecoder(req.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	var resBody GCPAppProjectsResponse

	projectIDS, ok := s.appIDProjects[reqBody.AppID]
	resBody.ProjectIDS = projectIDS
	if !ok {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	_ = json.NewEncoder(w).Encode(resBody)
}

func (s *MockIAMServer) hGCPUserHasRoleInProjects(w http.ResponseWriter, req *http.Request) {
	var reqBody GCPUserAccessRequest
	err := json.NewDecoder(req.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(reqBody.Resource, "projects/") {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
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

	_ = json.NewEncoder(w).Encode(resBody)
}

func (s *MockIAMServer) hEntraIDUserGroups(w http.ResponseWriter, req *http.Request) {
	var reqBody EntraIDUserGroupsRequest
	err := json.NewDecoder(req.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	var resBody EntraIDUserGroupsResponse

	resBody.Groups = s.userGroups[reqBody.User]
	_ = json.NewEncoder(w).Encode(resBody)
}

func (s *MockIAMServer) Url() string {
	return s.url
}

// Non-blocking
func (s *MockIAMServer) Start() error {
	if s.url != "" {
		return fmt.Errorf("server is already running")
	}

	var port string
	if s.port == 0 {
		port = ":0"
	} else {
		port = fmt.Sprintf(":%d", s.port)
	}

	l, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}

	s.url = fmt.Sprintf("http://localhost:%d", l.Addr().(*net.TCPAddr).Port)
	go func() {
		_ = s.server.Serve(l)
	}()

	return nil
}

func (s *MockIAMServer) Stop() error {
	err := s.server.Close()
	s.url = ""
	return err
}

type MockIAMServerOption func(*MockIAMServer)

func WithPort(port int) MockIAMServerOption {
	return func(s *MockIAMServer) {
		s.port = port
	}
}

func WithAppIDProjects(appID string, projectIDS []string) MockIAMServerOption {
	return func(s *MockIAMServer) {
		s.appIDProjects[appID] = projectIDS
	}
}

func WithUserProjectRoles(email string, projectID string, roles []string) MockIAMServerOption {
	return func(s *MockIAMServer) {
		project, ok := s.userProjectRoles[projectID]
		if !ok {
			project = map[string][]string{}
			s.userProjectRoles[projectID] = project
		}

		project[email] = roles
	}
}

func WithUserGroups(email string, groups []string) MockIAMServerOption {
	return func(s *MockIAMServer) {
		s.userGroups[email] = groups
	}
}

// NewMockIAMServer returns a new mock server which mimics the functionality of the IAM Lookup resource.
// It can be used along with NewIAMClient for local client -> server testing.
func NewMockIAMServer(opts ...MockIAMServerOption) (*MockIAMServer, error) {
	s := &MockIAMServer{
		appIDProjects:    map[string][]string{},
		userProjectRoles: map[string]map[string][]string{},
		userGroups:       map[string][]string{},
	}

	for _, opt := range opts {
		opt(s)
	}

	if s.port < 0 {
		return nil, fmt.Errorf("the assigned port must be a positive integer")
	}

	mux := http.NewServeMux()
	mux.Handle("POST /app/projects/gcp", enforceJSON(http.HandlerFunc(s.hGCPProjectIDS)))
	mux.Handle("POST /access/gcp", enforceJSON(http.HandlerFunc(s.hGCPUserHasRoleInProjects)))
	mux.Handle("POST /groups/entraid", enforceJSON(http.HandlerFunc(s.hEntraIDUserGroups)))

	s.server = &http.Server{
		ReadHeaderTimeout: defaultReadHeaderTimeout,
		Handler:           mux,
	}

	return s, nil
}
