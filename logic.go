package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"

	"cloud.google.com/go/pubsub/v2"
	"github.com/entur/go-logging"
)

// -----------------------
// Internal
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

func (c contextCache) Set(key string, value any) {
	c.values[key] = value
}

func newContextCache() contextCache {
	return contextCache{
		values: map[string]any{},
	}
}

type ctxKey struct{}

func process(ctx context.Context, so Orchestrator, h ManifestHandler, req *Request, res *Result) error {
	var err error

	ctx = context.WithValue(ctx, ctxKey{}, newContextCache())
	logger := logging.Ctx(ctx)

	project := so.ProjectID()
	version := h.ApiVersion()
	kind := h.Kind()
	action := req.Action

	before, ok := so.(MiddlewareBefore)
	if ok {
		logger.Debug().Msgf("Executing Orchestrator MiddlewareBefore (%s)", project)
		err = before.MiddlewareBefore(ctx, *req, res)
		if err != nil {
			return fmt.Errorf("orchestrator middleware (before): %w", err)
		}
		if res.locked {
			return nil
		}
	}

	before, ok = h.(MiddlewareBefore)
	if ok {
		logger.Debug().Msgf("Executing ManifestHandler MiddlewareBefore (%s, %s, %s)", version, kind, action)
		err = before.MiddlewareBefore(ctx, *req, res)
		if err != nil {
			return fmt.Errorf("manifesthandler middleware (before): %w", err)
		}
		if res.locked {
			return nil
		}
	}

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

	after, ok := h.(MiddlewareAfter)
	if ok {
		logger.Debug().Msgf("Executing ManifestHandler MiddlewareAfter (%s, %s, %s)", version, kind, action)
		err = after.MiddlewareAfter(ctx, *req, res)
		if err != nil {
			return fmt.Errorf("manifesthandler middleware (after): %w", err)
		}
		if res.locked {
			return nil
		}
	}

	after, ok = so.(MiddlewareAfter)
	if ok {
		logger.Debug().Msgf("Executing Orchestrator MiddlewareAfter (%s)", project)
		err = after.MiddlewareAfter(ctx, *req, res)
		if err != nil {
			return fmt.Errorf("orchestrator middleware (after): %w", err)
		}
		if res.locked {
			return nil
		}
	}

	if !res.locked {
		return fmt.Errorf("forgot to call .Done() in manifest handler (%s, %s, %s)", version, kind, action)
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

func Process(ctx context.Context, so Orchestrator, req *Request) Result {
	logger := logging.Ctx(ctx)
	logger.Debug().Interface("gorch_request", req).Msg("Processing request")

	var result Result
	var header ManifestHeader

	err := json.Unmarshal(req.Manifest.New, &header)
	if err != nil {
		err = fmt.Errorf("unable to unmarshal ManifestHeader: %w", err)
	} else {
		match := false

		for _, h := range so.Handlers() {
			if header.ApiVersion == h.ApiVersion() && header.Kind == h.Kind() {
				logger.Debug().Msgf("Found ManifestHandler (%s, %s)", header.ApiVersion, header.Kind)
				err = process(ctx, so, h, req, &result)
				match = true
				break
			}
		}

		if !match {
			err = fmt.Errorf("no matching ManifestHandler for (%s, %s)", header.ApiVersion, header.Kind)
		}
	}

	if err != nil {
		result.errs = append(result.errs, err)
	}

	return result
}

// Retrieve the value cache attached to the current request context
func Ctx(ctx context.Context) contextCache {
	v := ctx.Value(ctxKey{})
	if v == nil {
		return newContextCache()
	}
	c, _ := v.(contextCache)
	return c
}
