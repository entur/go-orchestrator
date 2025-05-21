package orchestrator

import (
	"context"
	"encoding/base64"
	"strings"

	"google.golang.org/api/idtoken"
)

// -----------------------
// Platform Orchestrator
// -----------------------

type ApiVersion string

type Kind string

type Metadata struct {
	ID string `json:"id"`
}

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

func (resource *ResourceIAMLookup) ToClient() IAMLookupClient {
	client, _ := idtoken.NewClient(context.Background(), resource.Url)
	return NewIAMLookupClient(client, resource.Url)
}

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

type Manifests[T any] struct {
	Old *T `json:"old"`
	New T  `json:"new"`
}

type Response struct {
	ApiVersion string        `json:"apiVersion"`
	Metadata   OuterMetadata `json:"metadata"`
	ResultCode ResultCode    `json:"result"`
	Output     string        `json:"output"`
}

type Request[T any] struct {
	ApiVersion    string        `json:"apiVersion"`
	Metadata      OuterMetadata `json:"metadata"`
	Resources     Resources     `json:"resources"`
	ResponseTopic string        `json:"responseTopic"`
	Action        Action        `json:"action"`
	Origin        Origin        `json:"origin"`
	Sender        Sender        `json:"sender"`
	Manifest      Manifests[T]  `json:"manifest"`
}

func (req Request[T]) ToResponse(code ResultCode, msg string) Response {
	return Response{
		ApiVersion: "orchestrator.entur.io/response/v1",
		Metadata:   req.Metadata,
		ResultCode: code,
		Output:     base64.StdEncoding.EncodeToString([]byte(msg)),
	}
}

// -----------------------
// Sub Orchestrator
// -----------------------

type Orchestrator[T any] interface {
	ProjectID() string
	Plan(context.Context, Request[T]) (Result, error)
	PlanDestroy(context.Context, Request[T]) (Result, error)
	Apply(context.Context, Request[T]) (Result, error)
	Destroy(context.Context, Request[T]) (Result, error)
}

type Result struct {
	Summary   string   // Your failure or success summary.
	Success   bool     // If the action succeeded or not. A false value indicates a user error
	Creations []string // A list of resources that are planned/being created.
	Updates   []string // A list of resources that are planned/being updated.
	Deletions []string // A list of resources that are planned/being deleted.
}

func (r *Result) String() string {
	if !r.Success {
		return r.Summary
	}
	if len(r.Creations) == 0 && len(r.Updates) == 0 && len(r.Deletions) == 0 {
		return "No changes detected"
	}

	var builder strings.Builder

	builder.WriteString(r.Summary)
	builder.WriteString("\n")
	if len(r.Creations) > 0 {
		builder.WriteString("Created:\n")
		for _, created := range r.Creations {
			builder.WriteString("+ ")
			builder.WriteString(created)
			builder.WriteString("\n")
		}
	}
	if len(r.Updates) > 0 {
		builder.WriteString("Updated:\n")
		for _, updated := range r.Updates {
			builder.WriteString("! ")
			builder.WriteString(updated)
			builder.WriteString("\n")
		}
	}
	if len(r.Deletions) > 0 {
		builder.WriteString("Deleted:\n")
		for _, deleted := range r.Deletions {
			builder.WriteString("- ")
			builder.WriteString(deleted)
			builder.WriteString("\n")
		}
	}

	return builder.String()
}
