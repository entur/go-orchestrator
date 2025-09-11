package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"cloud.google.com/go/pubsub/v2"
	"github.com/entur/go-logging"
)

// -----------------------
// Internal
// -----------------------

type ctxKey struct{}

func process(ctx context.Context, so Orchestrator, h ManifestHandler, req *Request, res *Result) error {
	var err error

	ctx = context.WithValue(ctx, ctxKey{}, ContextCache{
		values: map[string]any{},
	})
	logger := logging.Ctx(ctx)

	project := so.ProjectID()
	version := h.APIVersion()
	kind := h.Kind()
	action := req.Action

	before, ok := so.(MiddlewareBefore)
	if ok {
		logger.Debug().Msgf("Executing Orchestrator MiddlewareBefore (%s)", project)
		err = before.MiddlewareBefore(ctx, *req, res)
		if err != nil {
			return fmt.Errorf("orchestrator middleware (before): %w", err)
		}
	}

	before, ok = h.(MiddlewareBefore)
	if ok {
		logger.Debug().Msgf("Executing ManifestHandler MiddlewareBefore (%s, %s, %s)", version, kind, action)
		err = before.MiddlewareBefore(ctx, *req, res)
		if err != nil {
			return fmt.Errorf("manifesthandler middleware (before): %w", err)
		}
	}

	if !res.locked {
		logger.Debug().Msgf("Executing ManifestHandler (%s, %s, %s)", version, kind, action)
		switch req.Action {
		case ActionApply:
			err = h.Apply(ctx, *req, res)
		case ActionPlan:
			err = h.Plan(ctx, *req, res)
		case ActionPlanDestroy:
			err = h.PlanDestroy(ctx, *req, res)
		case ActionDestroy:
			err = h.Destroy(ctx, *req, res)
		default:
			err = fmt.Errorf("invalid action")
		}

		if err != nil {
			return fmt.Errorf("manifesthandler (%s, %s, %s): %w", version, kind, action, err)
		}
	} else {
		logger.Debug().Msgf("Skipping Executing ManifestHandler (%s, %s, %s) since result has already been set in middleware", version, kind, action)
	}

	after, ok := h.(MiddlewareAfter)
	if ok {
		logger.Debug().Msgf("Executing ManifestHandler MiddlewareAfter (%s, %s, %s)", version, kind, action)
		err = after.MiddlewareAfter(ctx, *req, res)
		if err != nil {
			return fmt.Errorf("manifesthandler middleware (after): %w", err)
		}
	}

	after, ok = so.(MiddlewareAfter)
	if ok {
		logger.Debug().Msgf("Executing Orchestrator MiddlewareAfter (%s)", project)
		err = after.MiddlewareAfter(ctx, *req, res)
		if err != nil {
			return fmt.Errorf("orchestrator middleware (after): %w", err)
		}
	}

	if !res.locked {
		return fmt.Errorf("forgot to call .Succeed(msg) or .Fail(msg) in manifest handler (%s, %s, %s)", version, kind, action)
	}

	return nil
}

func respond(ctx context.Context, publisher *pubsub.Publisher, res *Response) error {
	logger := logging.Ctx(ctx)
	logger.Debug().Interface("gorch_response", res).Msg("Sending response")

	enc, err := json.Marshal(res)
	if err != nil {
		return err
	}

	publishResult := publisher.Publish(ctx, &pubsub.Message{
		Data: enc,
	})
	_, err = publishResult.Get(ctx)
	return err
}

// -----------------------
// Core
// -----------------------

func Process(ctx context.Context, so Orchestrator, req *Request) *Result {
	logger := logging.Ctx(ctx)
	logger.Debug().Interface("gorch_request", req).Msg("Processing request")

	var header ManifestHeader
	result := &Result{}

	err := json.Unmarshal(req.Manifest.New, &header)
	if err != nil {
		err = fmt.Errorf("unable to unmarshal ManifestHeader: %w", err)
	} else {
		handlers := so.Handlers()

		// Loop through all manifest handlers in the so, and run the first one with a matching APIVersion and Kind.
		match := false
		for _, handler := range handlers {
			if header.APIVersion == handler.APIVersion() && header.Kind == handler.Kind() {
				logger.Debug().Msgf("Found ManifestHandler (%s, %s)", header.APIVersion, header.Kind)
				err = process(ctx, so, handler, req, result)
				match = true
				break
			}
		}

		// If we couldn't find a match, mark the result as having failed, and provide the user with a list of possible valid alternatives
		if !match {
			suggestions := make([]string, 0, len(handlers))
			for _, handler := range handlers {
				suggestion := fmt.Sprintf("apiVersion: %s\nkind: %s", handler.APIVersion(), handler.Kind())
				suggestions = append(suggestions, suggestion)
			}

			msg := fmt.Sprintf("The manifest apiVersion '%s' and kind '%s' is not valid. Perhaps you actually intended to use one of the following value combinations instead:\n%s", strings.Join(suggestions, "\n"))
			result.Fail(msg)
		}
	}

	if err != nil {
		result.errs = append(result.errs, err)
	}

	return result
}

type ContextCache struct {
	values map[string]any
}

func (c ContextCache) Get(key string) any {
	v, ok := c.values[key]
	if !ok {
		return nil
	}
	return v
}

func (c ContextCache) Set(key string, value any) {
	c.values[key] = value
}

// Retrieve the value cache attached to the current request context
func Ctx(ctx context.Context) ContextCache {
	v := ctx.Value(ctxKey{})
	if v == nil {
		return ContextCache{
			values: map[string]any{},
		}
	}

	//nolint:revive
	c, _ := v.(ContextCache)
	return c
}
