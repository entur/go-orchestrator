package orchestrator

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/entur/go-logging"
)

// -----------------------
// Platform Orchestrator
// -----------------------

type APIVersion string // Platform Orchestrator / Sub-Orchestrator APIVersion

const (
	APIVersionOrchestratorResponseV1 APIVersion = "orchestrator.entur.io/request/v1"  // Platform Orchestrator Request
	APIVersionOrchestratorRequestV1  APIVersion = "orchestrator.entur.io/response/v1" // Platform Orchestrator Response
)

type Kind string // Sub-Orchestrator Manifest Kind

type OuterMetadata struct {
	RequestID string `json:"requestId"` // Request ID specified by PO used to identify track the user request
}

type ResultCode string

const (
	ResultCodeSuccess ResultCode = "success" // Sub-Orchestrator succeeded in processing the action
	ResultCodeFailure ResultCode = "failure" // Sub-Orchestrator detected a user failure when processing the action
	ResultCodeNoop    ResultCode = "noop"    // Sub-Orchestrator detected no changes after processing the action
	ResultCodeError   ResultCode = "error"   // Sub-Orchestrator experienced an internal error when processing the action
)

type Resource struct {
	URL string `json:"url"` // 'https://eu-west1.cloudfunctions.net/someresource'
}

type ResourceIAM = Resource

type Resources struct {
	IAM ResourceIAM `json:"iamLookup"`
}

type Action string

const (
	ActionApply       Action = "apply"
	ActionPlan        Action = "plan"
	ActionPlanDestroy Action = "plan_destroy"
	ActionDestroy     Action = "destroy"
)

type RepositoryVisibility string

const (
	RepositoryVisbilityPublic   RepositoryVisibility = "public"
	RepositoryVisbilityInternal RepositoryVisibility = "internal"
	RepositoryVisbilityPrivate  RepositoryVisibility = "private"
)

type Repository struct {
	ID            int                  `json:"id"`            // '123123145'
	Name          string               `json:"name"`          // 'some-remo'
	FullName      string               `json:"fullName"`      // 'entur/some-repo'
	DefaultBranch string               `json:"defaultBranch"` // 'main'
	HtmlURL       string               `json:"htmlUrl"`       // 'https://github.com/entur/some-repo'
	Visibility    RepositoryVisibility `json:"visibility"`    // 'public'
}

type FileChanges struct {
	ContentsURL string `json:"contentsUrl"`
	BlobURL     string `json:"bloblUrl"`
	RawURL      string `json:"rawUrl"`
}

type PullRequestState string

const (
	PullRequestStateOpen   PullRequestState = "open"
	PullRequestStateClosed PullRequestState = "closed"
)

type PullRequest struct {
	ID      int              `json:"id"`    // '123123145'
	State   PullRequestState `json:"state"` // 'open'
	Ref     string           `json:"ref"`
	Title   string           `json:"title"` // 'chore: Added .entur manifests'
	Body    string           `json:"body"`
	Number  int              `json:"number"`
	Labels  []string         `json:"labels"`
	HtmlURL string           `json:"htmlUrl"`
}

type Origin struct {
	FileName    string      `json:"fileName"`
	Repository  Repository  `json:"repository"` // 'https://github.com/entur/some-repo'
	FileChanges FileChanges `json:"fileChanges"`
	PullRequest PullRequest `json:"pullRequest"`
}

type SenderType string

const (
	SenderTypeUser SenderType = "user" // Github user
	SenderTypeBot  SenderType = "bot"  //
)

type RepositoryPermission string

const (
	RepositoryPermissionAdmin    RepositoryPermission = "admin"
	RepositoryPermissionMaintain RepositoryPermission = "maintain"
	RepositoryPermissionWrite    RepositoryPermission = "write"
	RepositoryPermissionTriage   RepositoryPermission = "triage"
	RepositoryPermissionRead     RepositoryPermission = "read"
)

type Sender struct {
	Username   string               `json:"githubLogin"` // 'mockuser'
	Email      string               `json:"githubEmail"` // 'mockuser@entur.org'
	ID         int                  `json:"githubId"`
	Permission RepositoryPermission `json:"githubRepositoryPermission"` // 'admin'
	Type       SenderType           `json:"type"`                       // 'user'
}

type ManifestHeader struct {
	APIVersion APIVersion `json:"apiVersion"` // 'orchestrator.entur.io/mysuborchestrator/v1'
	Kind       Kind       `json:"kind"`       // 'mymanifestkind'
}

type Manifest = json.RawMessage

type Manifests struct {
	Old *Manifest `json:"old"`
	New Manifest  `json:"new"`
}

type Request struct {
	APIVersion    APIVersion    `json:"apiVersion"` // 'orchestrator.entur.io/request/v1'
	Metadata      OuterMetadata `json:"metadata"`
	Resources     Resources     `json:"resources"`
	ResponseTopic string        `json:"responseTopic"`
	Action        Action        `json:"action"`
	Origin        Origin        `json:"origin"`
	Sender        Sender        `json:"sender"`
	Manifest      Manifests     `json:"manifest"`
}

type Response struct {
	APIVersion APIVersion    `json:"apiVersion"` // 'orchestrator.entur.io/response/v1'
	Metadata   OuterMetadata `json:"metadata"`
	ResultCode ResultCode    `json:"result"` // 'success'
	Output     string        `json:"output"`
}

// -----------------------
// Sub-Orchestrator
// -----------------------

// The MiddlewareBefore interface represents the middleware running before every manifest event and/or a specific handler.
type MiddlewareBefore interface {
	MiddlewareBefore(context.Context, Request, *Result) error
}

// The MiddlewareAfter interface represents the middleware running after every manifest event and/or a specific handler.
type MiddlewareAfter interface {
	MiddlewareAfter(context.Context, Request, *Result) error
}

// The ManifestHandler interface represents the logic used for handling a specific APIVersion and Kind.
type ManifestHandler interface {
	// Which APIVersion and Kind this handler operates on
	APIVersion() APIVersion
	Kind() Kind
	// Actions
	Plan(context.Context, Request, *Result) error
	PlanDestroy(context.Context, Request, *Result) error
	Apply(context.Context, Request, *Result) error
	Destroy(context.Context, Request, *Result) error
}

// The Orchestrator interface represents the main configuration of a sub-orchestrator in a Project.
type Orchestrator interface {
	ProjectID() string           // The project this orchestrator is running in
	Handlers() []ManifestHandler // The manifests this orchestrator can handle
}

// The Change interface represents a planned/applied change in the context of a sub-orchestrator.
type Change interface {
	String() string
}

// Internal only struct used to represent simple string changes.
type simpleChange struct {
	text string
}

func (change simpleChange) String() string {
	return change.text
}

type Result struct {
	locked    bool     // If the result has been marked as done and lcoked
	summary   string   // Failure or Success summary
	success   bool     // If the action succeeded or not. A false value indicates a user error
	errs      []error  // The accumulated errors for this result
	creations []Change // A list of resources that are planned/being created.
	updates   []Change // A list of resources that are planned/being updated.
	deletions []Change // A list of resources that are planned/being deleted.
}

// Get all errors that have accumulated.
func (r *Result) Errors() []error {
	return r.errs
}

// Is the result locked for any further changes.
func (r *Result) Locked() bool {
	return r.locked
}

// Mark the result as having succeeded.
func (r *Result) Succeed(summary string) {
	if r.locked {
		r.errs = append(r.errs, logging.NewStackTraceError("attempted to mark a locked result as succeeded"))
	} else {
		r.locked = true
		r.summary = summary
		r.success = true
	}
}

// Mark the result as having failed.
func (r *Result) Fail(summary string) {
	if r.locked {
		r.errs = append(r.errs, logging.NewStackTraceError("attempted to mark a locked result as failed"))
	} else {
		r.locked = true
		r.summary = summary
		r.success = false
	}
}

// Add a new 'create' change to the result.
// Valid change types are:
// * string
// * Stringer/Change interface
func (r *Result) Create(change ...any) {
	if r.locked {
		r.errs = append(r.errs, logging.NewStackTraceError("attempted to add a new 'create' change to a locked result"))
		return
	}

	for _, val := range change {
		switch v := val.(type) {
		case string:
			r.creations = append(r.creations, simpleChange{v})
		case []string:
			for _, str := range v {
				r.creations = append(r.creations, simpleChange{str})
			}
		case Change:
			r.creations = append(r.creations, v)
		case []Change:
			r.creations = append(r.creations, v...)
		default:
			r.errs = append(r.errs, logging.NewStackTraceError("attempted to add a new 'create' change that is not of 'string', '[]string', 'Change' or '[]Change' type"))
		}
	}
}

// Get all current 'create' changes.
func (r *Result) Creations() []Change {
	creations := make([]Change, len(r.creations))
	copy(creations, r.creations)
	return creations
}

// Add a new 'update' change to the result.
// Valid change types are:
// * string
// * Stringer/Change interface
func (r *Result) Update(change ...any) {
	if r.locked {
		r.errs = append(r.errs, logging.NewStackTraceError("attempted to add a new 'update' change to a locked result"))
		return
	}

	for _, val := range change {
		switch v := val.(type) {
		case string:
			r.updates = append(r.updates, simpleChange{v})
		case []string:
			for _, str := range v {
				r.updates = append(r.updates, simpleChange{str})
			}
		case Change:
			r.updates = append(r.updates, v)
		case []Change:
			r.updates = append(r.updates, v...)
		default:
			r.errs = append(r.errs, logging.NewStackTraceError("attempted to add a new 'update' change that is not of 'string', '[]string', 'Change' or '[]Change' type"))
		}
	}
}

// Get all current 'update' changes.
func (r *Result) Updates() []Change {
	updates := make([]Change, len(r.updates))
	copy(updates, r.updates)
	return updates
}

// Add a new 'delete' change to the result.
// Valid change types are:
// * string
// * Stringer/Change interface
func (r *Result) Delete(change ...any) {
	if r.locked {
		r.errs = append(r.errs, logging.NewStackTraceError("attempted to add a new 'delete' change to a locked result"))
		return
	}

	for _, val := range change {
		switch v := val.(type) {
		case string:
			r.deletions = append(r.deletions, simpleChange{v})
		case []string:
			for _, str := range v {
				r.deletions = append(r.deletions, simpleChange{str})
			}
		case Change:
			r.deletions = append(r.deletions, v)
		case []Change:
			r.deletions = append(r.deletions, v...)
		default:
			r.errs = append(r.errs, logging.NewStackTraceError("attempted to add a new 'delete' change that is not of 'string', '[]string', 'Change' or '[]Change' type"))
		}
	}
}

// Get all current 'delete' changes.
func (r *Result) Deletions() []Change {
	deletions := make([]Change, len(r.deletions))
	copy(deletions, r.deletions)
	return deletions
}

// Get the final result code.
func (r *Result) Code() ResultCode {
	if len(r.errs) > 0 || !r.locked {
		return ResultCodeError
	}
	if !r.success {
		return ResultCodeFailure
	}
	if len(r.creations) == 0 && len(r.updates) == 0 && len(r.deletions) == 0 {
		return ResultCodeNoop
	}
	return ResultCodeSuccess
}

// Get the final result string output.
func (r *Result) Output() string {
	if len(r.errs) > 0 || !r.locked {
		return "Internal error"
	}
	if !r.success {
		return r.summary
	}
	if len(r.creations) == 0 && len(r.updates) == 0 && len(r.deletions) == 0 {
		return "No changes"
	}

	var builder strings.Builder

	if r.summary != "" {
		builder.WriteString(r.summary)
		builder.WriteString("\n")
	}

	if len(r.creations) > 0 {
		builder.WriteString("Create:\n")
		for _, create := range r.creations {
			builder.WriteString("+ ")
			builder.WriteString(create.String())
			builder.WriteString("\n")
		}
	}
	if len(r.updates) > 0 {
		builder.WriteString("Update:\n")
		for _, update := range r.updates {
			builder.WriteString("! ")
			builder.WriteString(update.String())
			builder.WriteString("\n")
		}
	}
	if len(r.deletions) > 0 {
		builder.WriteString("Delete:\n")
		for _, delete := range r.deletions {
			builder.WriteString("- ")
			builder.WriteString(delete.String())
			builder.WriteString("\n")
		}
	}

	return builder.String()
}
