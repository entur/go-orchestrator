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
	url              string
	appIDProjects    map[string][]string
	userProjectRoles map[string]map[string][]string
	userGroups       map[string][]string
}

func (s *MockIAMLookupServer) hGCPProjectIDS(w http.ResponseWriter, req *http.Request) {
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

	json.NewEncoder(w).Encode(resBody)
}

func (s *MockIAMLookupServer) hGCPUserHasRoleInProjects(w http.ResponseWriter, req *http.Request) {
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

	json.NewEncoder(w).Encode(resBody)
}

func (s *MockIAMLookupServer) hEntraIDUserGroups(w http.ResponseWriter, req *http.Request) {
	var reqBody EntraIDUserGroupsRequest
	err := json.NewDecoder(req.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	var resBody EntraIDUserGroupsResponse

	resBody.Groups = s.userGroups[reqBody.User]
	json.NewEncoder(w).Encode(resBody)
}

func (s *MockIAMLookupServer) Url() string {
	return s.url
}

// Non-blocking
func (s *MockIAMLookupServer) Start() error {
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
	go s.server.Serve(l)

	return nil
}

func (s *MockIAMLookupServer) Stop() error {
	err := s.server.Close()
	s.url = ""
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

func NewMockIAMLookupServer(opts ...MockIAMLookupServerOption) (*MockIAMLookupServer, error) {
	s := &MockIAMLookupServer{
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
		Handler: mux,
	}

	return s, nil
}
