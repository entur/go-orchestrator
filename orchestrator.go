package orchestrator

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"cloud.google.com/go/pubsub"
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
	ResultCodeSuccess ResultCode = "success" // Sub-Orchestrator succeeded in processing the action
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

type OrchestratorMiddlewareBefore interface {
	MiddlewareBefore(context.Context, Request, *Result) error
}

type OrchestratorMiddlewareAfter interface {
	MiddlewareAfter(context.Context, Request, *Result) error
}

type Result struct {
	done      bool     // If the result has been marked as done
	errs      error    // The accumulated errors for this result
	summary   string   // Failure or Success summary
	success   bool     // If the action succeeded or not. A false value indicates a user error
	creations []string // A list of resources that are planned/being created.
	updates   []string // A list of resources that are planned/being updated.
	deletions []string // A list of resources that are planned/being deleted.
}

func (r *Result) AccumulatedError() error {
	return r.errs
}

func (r *Result) Done(summary string, success bool) {
	if r.done {
		r.errs = errors.Join(r.errs, logging.NewStackTraceError("attempted to mark an already finished result as done"))
	} else {
		r.done = true
		r.summary = summary
		r.success = success
	}
}

func (r *Result) Create(change ...string) {
	if r.done {
		r.errs = errors.Join(r.errs, logging.NewStackTraceError("attempted to add a create change to an already finished result"))
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
		r.errs = errors.Join(r.errs, logging.NewStackTraceError("attempted to add an update change to an already finished result"))
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
		r.errs = errors.Join(r.errs, logging.NewStackTraceError("attempted to add a delete change to an already finished result"))
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
	if r.errs != nil || !r.done {
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

func (r *Result) String() string {
	if r.errs != nil || !r.done {
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

// -----------------------
// Processing
// -----------------------

type contextCache struct {
	values map[string]any
}

func (c contextCache) Get(key string) any {
	v, ok := c.values[key]
	if !ok {
		return nil
	}
	return v
}

func (c contextCache)  Set(key string, value any) {
	c.values[key] = value
}

func newContextCache() contextCache {
	return contextCache{
		values: map[string]any{},
	}
}

type ctxKey struct{}

// Retrieve the cache attached to the current request context
func CtxCache(ctx context.Context) contextCache {	
	v := ctx.Value(ctxKey{})
	if v == nil {
		return newContextCache()
	}
	c, _ := v.(contextCache)
	return c
}

func Receive(ctx context.Context, so Orchestrator, req Request) Result {
	logger := logging.Ctx(ctx)
	logger.Info().Interface("gorch_request", req).Msg("Received and processing request")

	var result Result
	var header ManifestHeader

	err := json.Unmarshal(req.Manifest.New, &header)
	if err != nil {
		err = fmt.Errorf("unable to unmarshal ManifestHeader: %w", err)
	} else {
		match := false

		for _, h := range so.Handlers() {
			if header.ApiVersion == h.ApiVersion() && header.Kind == h.Kind() {
				logger.Debug().Msgf("Found ManifestHandler %s %s", header.ApiVersion, header.Kind)
				match = true
				ctx = context.WithValue(ctx, ctxKey{}, newContextCache())

				before, ok := so.(OrchestratorMiddlewareBefore)
				if ok {
					logger.Debug().Msg("Executing MiddlewareBefore handler")
					err = before.MiddlewareBefore(ctx, req, &result)
					if err != nil {
						err = fmt.Errorf("so middleware (before): %w", err)
						break
					}
					if result.done {
						break
					}
				}

				logger.Debug().Msgf("Executing ManifestHandler %s %s %s", header.ApiVersion, header.Kind, req.Action)
				switch req.Action {
				case ActionApply:
					err = h.Apply(ctx, req, &result)
				case ActionPlan:
					err = h.Plan(ctx, req, &result)
				case ActionPlanDestroy:
					err = h.PlanDestroy(ctx, req, &result)
				case ActionDestroy:
					err = h.Destroy(ctx, req, &result)
				default:
					err = fmt.Errorf("invalid action")
				}

				if err != nil {
					err = fmt.Errorf("ManifestHandler %s %s %s: %w", header.ApiVersion, header.Kind, req.Action, err)
					break
				}

				after, ok := so.(OrchestratorMiddlewareAfter)
				if ok {
					logger.Debug().Msg("Executing MiddlewareAfter handler")
					err = after.MiddlewareAfter(ctx, req, &result)
					if err != nil {
						err = fmt.Errorf("so middleware (after): %w", err)
						break
					}
					if result.done {
						break
					}
				}

				if !result.done {
					err = fmt.Errorf("forgot to call .Done() in handler %s %s %s", header.ApiVersion, header.Kind, req.Action)
				}

				break
			}
		}

		if !match {
			err = fmt.Errorf("no matching ManifestHandler for %s %s", header.ApiVersion, header.Kind)
		}
	}

	result.errs = errors.Join(result.errs, err)
	return result
}

func Respond(ctx context.Context, topic *pubsub.Topic, res Response) error {
	logger := logging.Ctx(ctx)
	logger.Info().Interface("gorch_response", res).Msg("Sending response")

	if topic == nil {
		return fmt.Errorf("no topic set, unable to respond")
	}

	enc, err := json.Marshal(res)
	if err != nil {
		return err
	}

	result := topic.Publish(ctx, &pubsub.Message{
		Data: enc,
	})
	_, err = result.Get(ctx)
	return err
}
