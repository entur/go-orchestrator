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

type Changes struct {
	create []string
	update []string
	delete []string
}

func (c *Changes) AddCreate(msg string) {
	c.create = append(c.create, msg)
}

func (c *Changes) AddUpdate(msg string) {
	c.update = append(c.update, msg)
}

func (c *Changes) AddDelete(msg string) {
	c.delete = append(c.delete, msg)
}

func (c *Changes) IsEmpty() bool {
	return len(c.create) == 0 && len(c.update) == 0 && len(c.delete) == 0
}

func (c *Changes) Clear() {
	c.create = c.create[:0]
	c.update = c.update[:0]
	c.delete = c.delete[:0]
}

type Result struct {
	Summary string
	Code    ResultCode
	Changes Changes
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
	if r.Code == ResultCodeSuccess && r.Changes.IsEmpty() {
		r.Code = resultCodeNoop
	}
	var builder strings.Builder
	builder.WriteString(r.Summary)
	builder.WriteString("\n")
	if len(r.Changes.create) > 0 {
		builder.WriteString("Created:\n")
		for _, created := range r.Changes.create {
			builder.WriteString(created)
			builder.WriteString("\n")
		}
	}
	if len(r.Changes.update) > 0 {
		builder.WriteString("Updated:\n")
		for _, updated := range r.Changes.update {
			builder.WriteString(updated)
			builder.WriteString("\n")
		}
	}
	if len(r.Changes.delete) > 0 {
		builder.WriteString("Deleted:\n")
		for _, deleted := range r.Changes.delete {
			builder.WriteString(deleted)
			builder.WriteString("\n")
		}
	}
	// TODO: format
	return Response{
		ApiVersion: "orchestrator.entur.io/response/v1",
		Metadata:   req.Metadata,
		ResultCode: r.Code,
		Output:     base64.StdEncoding.EncodeToString([]byte(builder.String())),
	}
}
