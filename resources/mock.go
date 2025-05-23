package resources

import (
	"encoding/json"
	"fmt"
	"mime"
	"net"
	"net/http"
	"strings"
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
func (s *MockIAMLookupServer) Serve() (string, error) {
	if s.up {
		return "", fmt.Errorf("server is already running")
	}

	portStr := ":0"
	if s.port > 0 {
		portStr = fmt.Sprintf(":%d", s.port)
	}

	l, err := net.Listen("tcp", portStr)
	if err != nil {
		return "", err
	}

	go s.server.Serve(l)

	s.port = l.Addr().(*net.TCPAddr).Port
	return fmt.Sprintf("http://localhost:%d", s.port), nil
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
