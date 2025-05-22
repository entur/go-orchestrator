package orchestrator

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"

	"github.com/entur/go-logging"
)

// -----------------------
// Platform Orchestrator
// -----------------------

type ApiVersion string

type Kind string

type OuterMetadata struct {
	RequestID string `json:"requestId"`
}

type ResultCode string

const (
	ResultCodeSuccess ResultCode = "success" // Sub-Orchestrator succeded in processing the action
	ResultCodeFailure ResultCode = "failure" // Sub-Orchestrator detected a user failure when processing the action
	ResultCodeNoop    ResultCode = "noop"    // Sub-Orchestrator detected no changes after processing the action
	ResultCodeError   ResultCode = "error"   // Sub-Orchestrator experienced an internal error when processing the action
)

type Output string

type Resource struct {
	Url string `json:"url"`
}

type ResourceIAMLookup = Resource

type Resources struct {
	IAM ResourceIAMLookup `json:"iamLookup"`
}

type Action string

const (
	ActionApply       Action = "apply"
	ActionPlan        Action = "plan"
	ActionPlanDestroy Action = "plan_destroy"
	ActionDestroy     Action = "destroy"
)

type GitRepository struct {
	HtmlUrl string `json:"htmlUrl"`
}

type Origin struct {
	FileName   string        `json:"fileName"`
	Repository GitRepository `json:"repository"`
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

type Manifest = json.RawMessage

type Manifests struct {
	Old *Manifest `json:"old"`
	New Manifest  `json:"new"`
}

type Request struct {
	ApiVersion    string        `json:"apiVersion"`
	Metadata      OuterMetadata `json:"metadata"`
	Resources     Resources     `json:"resources"`
	ResponseTopic string        `json:"responseTopic"`
	Action        Action        `json:"action"`
	Origin        Origin        `json:"origin"`
	Sender        Sender        `json:"sender"`
	Manifest      Manifests     `json:"manifest"`
}

type Response struct {
	ApiVersion string        `json:"apiVersion"`
	Metadata   OuterMetadata `json:"metadata"`
	ResultCode ResultCode    `json:"result"`
	Output     string        `json:"output"`
}

func NewResponse(metadata OuterMetadata, code ResultCode, msg string) Response {
	return Response{
		ApiVersion: "orchestrator.entur.io/response/v1",
		Metadata:   metadata,
		ResultCode: code,
		Output:     base64.StdEncoding.EncodeToString([]byte(msg)),
	}
}

// -----------------------
// Sub Orchestrator
// -----------------------

type ActionHandler = func(context.Context, Request, *ResponseResult) error

type ManifestHandler interface {
	// Which ApiVersion and Kind this handler correlates with
	ApiVersion() ApiVersion
	Kind() Kind
	// Actions
	Plan(context.Context, Request, *ResponseResult) error
	PlanDestroy(context.Context, Request, *ResponseResult) error
	Apply(context.Context, Request, *ResponseResult) error
	Destroy(context.Context, Request, *ResponseResult) error
}

type Orchestrator interface {
	ProjectID() string
	Handlers() []ManifestHandler
}

type OrchestratorMiddlewareBefore interface {
	MiddlewareBefore(context.Context, Request, *ResponseResult) error
}

type OrchestratorMiddlewareAfter interface {
	MiddlewareAfter(context.Context, Request, *ResponseResult) error
}

type ResponseResult struct {
	lock     bool
	mistakes error

	summary   string   // Your failure or success summary.
	success   bool     // If the action succeeded or not. A false value indicates a user error
	creations []string // A list of resources that are planned/being created.
	updates   []string // A list of resources that are planned/being updated.
	deletions []string // A list of resources that are planned/being deleted.
}

func (r *ResponseResult) Succeeded() bool {
	return r.success
}

func (r *ResponseResult) HasChanges() bool {
	return len(r.creations) > 0 || len(r.updates) > 0 && len(r.deletions) > 0
}

func (r *ResponseResult) Done(summary string, success bool) {
	if r.lock {
		r.mistakes = errors.Join(r.mistakes, logging.NewStackTraceError("already done"))
	} else {
		r.lock = true
		r.summary = summary
		r.success = success
	}
}

func (r *ResponseResult) Create(change ...string) {
	if r.lock {
		r.mistakes = errors.Join(r.mistakes, logging.NewStackTraceError("already done"))
	} else {
		r.creations = append(r.creations, change...)
	}
}

func (r *ResponseResult) Creations() []string {
	creations := make([]string, len(r.creations))
	copy(creations, r.creations)
	return creations
}

func (r *ResponseResult) Update(change ...string) {
	if r.lock {
		r.mistakes = errors.Join(r.mistakes, logging.NewStackTraceError("already done"))
	} else {
		r.updates = append(r.updates, change...)
	}
}

func (r *ResponseResult) Updates() []string {
	updates := make([]string, len(r.updates))
	copy(updates, r.updates)
	return updates
}

func (r *ResponseResult) Delete(change ...string) {
	if r.lock {
		r.mistakes = errors.Join(r.mistakes, logging.NewStackTraceError("already done"))
	} else {
		r.deletions = append(r.deletions, change...)
	}
}

func (r *ResponseResult) Deletions() []string {
	deletions := make([]string, len(r.deletions))
	copy(deletions, r.deletions)
	return deletions
}

func (r *ResponseResult) String() string {
	if !r.Succeeded() {
		return r.summary
	}
	if !r.HasChanges() {
		return "No changes detected"
	}

	var builder strings.Builder

	builder.WriteString(r.summary)
	builder.WriteString("\n")
	if len(r.creations) > 0 {
		builder.WriteString("Created:\n")
		for _, created := range r.creations {
			builder.WriteString("+ ")
			builder.WriteString(created)
			builder.WriteString("\n")
		}
	}
	if len(r.updates) > 0 {
		builder.WriteString("Updated:\n")
		for _, updated := range r.updates {
			builder.WriteString("! ")
			builder.WriteString(updated)
			builder.WriteString("\n")
		}
	}
	if len(r.deletions) > 0 {
		builder.WriteString("Deleted:\n")
		for _, deleted := range r.deletions {
			builder.WriteString("- ")
			builder.WriteString(deleted)
			builder.WriteString("\n")
		}
	}

	return builder.String()
}
