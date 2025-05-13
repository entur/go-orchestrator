package orchestrator

import (
	"context"
	"encoding/base64"
	"strings"
)

// Manifest Types
type ApiVersion string
type Kind string
type Metadata struct {
	ID string `json:"id"`
}

type Orchestrator[T any] interface {
	Plan(context.Context, Request[T]) (Result, error)
	PlanDestroy(context.Context, Request[T]) (Result, error)
	Apply(context.Context, Request[T]) (Result, error)
	Destroy(context.Context, Request[T]) (Result, error)
}

type Action string

const (
	ActionApply       Action = "apply"
	ActionPlan        Action = "plan"
	ActionPlanDestroy Action = "plan_destroy"
	ActionDestroy     Action = "destroy"
)

type Manifests[T any] struct {
	Old *T `json:"old"`
	New T  `json:"new"`
}

type OuterMetadata struct {
	RequestID string `json:"request_id"`
}
type Resource struct {
	Url string `json:"url"`
}

type Resources struct {
	IAM Resource `json:"iamLookup"`
}

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

type Output string
type ResultCode string

// The possible results of the sub-orchestrator response
const (
	ResultCodeSuccess ResultCode = "success"
	ResultCodeFailure ResultCode = "failure"
	resultCodeNoop    ResultCode = "noop"
	resultCodeError   ResultCode = "error"
)

type Response struct {
	ApiVersion string        `json:"apiVersion"`
	Metadata   OuterMetadata `json:"metadata"`
	ResultCode ResultCode    `json:"result"`
	Output     string        `json:"output"`
}

func (r *Result) NoChanges() bool {
	return len(r.Creations) == 0 && len(r.Updates) == 0 && len(r.Deletions) == 0
}

func (r *Result) String() string {
	var builder strings.Builder
	builder.WriteString(r.Summary)
	builder.WriteString("\n")
	if len(r.Creations) > 0 {
		builder.WriteString("Created:\n")
		for _, created := range r.Creations {
			builder.WriteString(created)
			builder.WriteString("\n")
		}
	}
	if len(r.Updates) > 0 {
		builder.WriteString("Updated:\n")
		for _, updated := range r.Updates {
			builder.WriteString(updated)
			builder.WriteString("\n")
		}
	}
	if len(r.Deletions) > 0 {
		builder.WriteString("Deleted:\n")
		for _, deleted := range r.Deletions {
			builder.WriteString(deleted)
			builder.WriteString("\n")
		}
	}
	return builder.String()
}

type Result struct {
	Summary   string
	Code      ResultCode
	Creations []string
	Updates   []string
	Deletions []string
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

func (req Request[T]) ToResponse(r Result) Response {
	if r.Code == ResultCodeSuccess && r.NoChanges() {
		r.Code = resultCodeNoop
	}
	return Response{
		ApiVersion: "orchestrator.entur.io/response/v1",
		Metadata:   req.Metadata,
		ResultCode: r.Code,
		Output:     base64.StdEncoding.EncodeToString([]byte(r.String())),
	}
}
