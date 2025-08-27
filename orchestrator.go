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

type ApiVersion string

const (
	ApiVersionOrchestratorResponseV1 ApiVersion = "orchestrator.entur.io/request/v1"
	ApiVersionOrchestratorRequestV1  ApiVersion = "orchestrator.entur.io/response/v1"
)

type Kind string

type OuterMetadata struct {
	RequestID string `json:"requestId"`
}

type ResultCode string

const (
	ResultCodeSuccess ResultCode = "success" // Sub-Orchestrator succeeded in processing the action
	ResultCodeFailure ResultCode = "failure" // Sub-Orchestrator detected a user failure when processing the action
	ResultCodeNoop    ResultCode = "noop"    // Sub-Orchestrator detected no changes after processing the action
	ResultCodeError   ResultCode = "error"   // Sub-Orchestrator experienced an internal error when processing the action
)

type Output string

type Resource struct {
	Url string `json:"url"`
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
	ID            int                  `json:"id"`            // E.g. '123123145'
	Name          string               `json:"name"`          // E.g. 'some-remo'
	FullName      string               `json:"fullName"`      // E.g. 'entur/some-repo'
	DefaultBranch string               `json:"defaultBranch"` // E.g. 'main'
	HtmlUrl       string               `json:"htmlUrl"`       // E.g. 'https://github.com/entur/some-repo'
	Visibility    RepositoryVisibility `json:"visibility"`    // E.g. 'public'
}

type Origin struct {
	FileName   string     `json:"fileName"`
	Repository Repository `json:"repository"`
}

type SenderType string

const (
	SenderTypeUser SenderType = "user"
	SenderTypeBot  SenderType = "bot"
)

type Sender struct {
	Email string     `json:"githubEmail"`
	ID    int        `json:"githubId"`
	Type  SenderType `json:"type"`
}

type ManifestHeader struct {
	ApiVersion ApiVersion `json:"apiVersion"`
	Kind       Kind       `json:"kind"`
}

type Manifest = json.RawMessage

type Manifests struct {
	Old *Manifest `json:"old"`
	New Manifest  `json:"new"`
}

type Request struct {
	ApiVersion    ApiVersion    `json:"apiVersion"`
	Metadata      OuterMetadata `json:"metadata"`
	Resources     Resources     `json:"resources"`
	ResponseTopic string        `json:"responseTopic"`
	Action        Action        `json:"action"`
	Origin        Origin        `json:"origin"`
	Sender        Sender        `json:"sender"`
	Manifest      Manifests     `json:"manifest"`
}

type Response struct {
	ApiVersion ApiVersion    `json:"apiVersion"`
	Metadata   OuterMetadata `json:"metadata"`
	ResultCode ResultCode    `json:"result"`
	Output     string        `json:"output"`
}

// -----------------------
// Sub Orchestrator
// -----------------------

type Middleware = func(context.Context, Request, *Result) error

type MiddlewareBefore interface {
	MiddlewareBefore(context.Context, Request, *Result) error
}

type MiddlewareAfter interface {
	MiddlewareAfter(context.Context, Request, *Result) error
}

type ManifestHandler interface {
	// Which ApiVersion and Kind this handler correlates with
	ApiVersion() ApiVersion
	Kind() Kind
	// Actions
	Plan(context.Context, Request, *Result) error
	PlanDestroy(context.Context, Request, *Result) error
	Apply(context.Context, Request, *Result) error
	Destroy(context.Context, Request, *Result) error
}

type Orchestrator interface {
	ProjectID() string           // The project this orchestrator is running in
	Handlers() []ManifestHandler // The manifests this orchestrator can handle
}

type Result struct {
	done      bool     // If the result has been marked as done
	summary   string   // Failure or Success summary
	success   bool     // If the action succeeded or not. A false value indicates a user error
	errs      []error    // The accumulated errors for this result
	creations []string // A list of resources that are planned/being created.
	updates   []string // A list of resources that are planned/being updated.
	deletions []string // A list of resources that are planned/being deleted.
}

func (r *Result) Errors() []error {
	return r.errs
}

func (r *Result) IsDone() bool {
	return r.done
}

func (r *Result) Done(summary string, success bool) {
	if r.done {
		r.errs = append(r.errs, logging.NewStackTraceError("attempted to mark an already finished result as done"))
	} else {
		r.done = true
		r.summary = summary
		r.success = success
	}
}

func (r *Result) Create(change ...string) {
	if r.done {
		r.errs = append(r.errs, logging.NewStackTraceError("attempted to add a create change to an already finished result"))
	} else {
		r.creations = append(r.creations, change...)
	}
}

func (r *Result) Creations() []string {
	creations := make([]string, len(r.creations))
	copy(creations, r.creations)
	return creations
}

func (r *Result) Update(change ...string) {
	if r.done {
		r.errs = append(r.errs, logging.NewStackTraceError("attempted to add an update change to an already finished result"))
	} else {
		r.updates = append(r.updates, change...)
	}
}

func (r *Result) Updates() []string {
	updates := make([]string, len(r.updates))
	copy(updates, r.updates)
	return updates
}

func (r *Result) Delete(change ...string) {
	if r.done {
		r.errs = append(r.errs, logging.NewStackTraceError("attempted to add a delete change to an already finished result"))
	} else {
		r.deletions = append(r.deletions, change...)
	}
}

func (r *Result) Deletions() []string {
	deletions := make([]string, len(r.deletions))
	copy(deletions, r.deletions)
	return deletions
}

func (r *Result) Code() ResultCode {
	if len(r.errs) > 0 || !r.done {
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

func (r *Result) Output() string {
	if len(r.errs) > 0 || !r.done {
		return "Internal error"
	}
	if !r.success {
		return r.summary
	}
	if len(r.creations) == 0 && len(r.updates) == 0 && len(r.deletions) == 0 {
		return "No changes"
	}

	var builder strings.Builder

	builder.WriteString(r.summary)
	builder.WriteString("\n")
	if len(r.creations) > 0 {
		builder.WriteString("Create:\n")
		for _, create := range r.creations {
			builder.WriteString("+ ")
			builder.WriteString(create)
			builder.WriteString("\n")
		}
	}
	if len(r.updates) > 0 {
		builder.WriteString("Update:\n")
		for _, update := range r.updates {
			builder.WriteString("! ")
			builder.WriteString(update)
			builder.WriteString("\n")
		}
	}
	if len(r.deletions) > 0 {
		builder.WriteString("Delete:\n")
		for _, delete := range r.deletions {
			builder.WriteString("- ")
			builder.WriteString(delete)
			builder.WriteString("\n")
		}
	}

	return builder.String()
}
