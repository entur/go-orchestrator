package orchestrator

import (
	"context"
	"encoding/base64"
)

// Manifest Types
type ApiVersion string
type Kind string
type Metadata struct {
	Id string `json:"id"`
}

type Orchestrator[T any] interface {
	Plan(context.Context, Request[T]) (Result, error)
	PlanDestroy(context.Context, Request[T]) (Result, error)
	Apply(context.Context, Request[T]) (Result, error)
	Destroy(context.Context, Request[T]) (Result, error)
}

type Action string

const (
	Apply       Action = "apply"
	Plan        Action = "plan"
	PlanDestroy Action = "plan_destroy"
	Destroy     Action = "destroy"
)

type Manifests[T any] struct {
	Old *T `json:"old"`
	New T  `json:"new"`
}

type OuterMetadata struct {
	RequestId string `json:"request_id"`
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
	Success ResultCode = "success"
	Failure ResultCode = "failure"
	Noop    ResultCode = "noop"
	Error   ResultCode = "error"
)

type Response struct {
	ApiVersion string        `json:"apiVersion"`
	Metadata   OuterMetadata `json:"metadata"`
	ResultCode ResultCode    `json:"result"`
	Output     string        `json:"output"`
}

type Result struct {
	Code   ResultCode
	Output string
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
	return Response{
		ApiVersion: "orchestrator.entur.io/response/v1",
		Metadata:   req.Metadata,
		ResultCode: r.Code,
		Output:     base64.StdEncoding.EncodeToString([]byte(r.Output)),
	}
}
