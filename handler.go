package orchestrator

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"cloud.google.com/go/pubsub/v2"
	cloudevent "github.com/cloudevents/sdk-go/v2/event"
	"github.com/entur/go-logging"
	"github.com/rs/zerolog"
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
		if res.done {
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
		if res.done {
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
		if res.done {
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
		if res.done {
			return nil
		}
	}

	if !res.done {
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

// -----------------------
// Handlers
// -----------------------

type HandlerConfig struct {
	logger *zerolog.Logger
}

type HandlerOption func(*HandlerConfig)

func WithCustomLogger(logger zerolog.Logger) HandlerOption {
	return func(c *HandlerConfig) {
		c.logger = &logger
	}
}

func NewCloudEventHandler(so Orchestrator, opts ...HandlerOption) func(context.Context, cloudevent.Event) error {
	cfg := &HandlerConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	var parentLogger zerolog.Logger
	if cfg.logger != nil {
		parentLogger = *cfg.logger
	} else {
		parentLogger = logging.New()
	}

	client, _ := pubsub.NewClient(context.Background(), so.ProjectID())
	/*
		TODO: Still need to figure out what to do here
		if err != nil {
			errStr := err.Error()
			if !strings.HasPrefix(errStr, "pubsub(publisher): credentials: could not find default credentials.") {
				parentLogger.Panic().Err(err).Msg("Failed to create underlying pubsub client")
			}

			//option.WithCredentialsJSON([]byte(`{"type": "external_account", "audience": "test", "subject_token_type": "test"}`)),


			os.Setenv("PUBSUB_EMULATOR_HOST", )

			client, err = pubsub.NewClient(context.Background(), "",
				option.WithoutAuthentication(),
				option.WithTelemetryDisabled(),
				internaloption.SkipDialSettingsValidation(),
				option.WithGRPCDialOption(
					grpc.WithTransportCredentials(insecure.NewCredentials()),
				),
			)

				option.WithAuthCredentials(auth.NewCredentials(&auth.CredentialsOptions{
					JSON: []byte(`{"type": "external_account", "audience": "test", "subject_token_type": "test"}`),
				})),

			)
			if err != nil {
				return func(ctx context.Context, e cloudevent.Event) error {
					return err
				}
			}
		}
	*/

	publishers := map[string]*pubsub.Publisher{}
	mu := sync.Mutex{}

	parentLogger.Debug().Msg("Created a new CloudEventHandler")
	return func(ctx context.Context, e cloudevent.Event) error {
		logger := parentLogger.With().Logger()
		
		var req Request

		err := UnmarshalCloudEvent(e, &req)
		if err != nil {
			logger.Error().Err(err).Msg("Encountered an internal error when unmarshalling CloudEvent")
			return err
		}

		logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
			return c.Int("gorch_github_user_id", req.Sender.ID).
				Str("gorch_request_id", req.Metadata.RequestID).
				Str("gorch_file_name", req.Origin.FileName).
				Str("gorch_action", string(req.Action))
		})

		ctx = logger.WithContext(ctx)
		result := Process(ctx, so, &req)

		logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
			return c.Interface("gorch_result_summary", result.summary).
				Interface("gorch_result_creations", result.creations).
				Interface("gorch_result_updates", result.updates).
				Interface("gorch_result_deletions", result.deletions)
		})

		if errs := result.Errors(); len(errs) > 0 {
			err = errors.Join(errs...)
			logger.Error().Stack().Err(err).Msg("Encountered an internal error whilst processing request")
		}

		mu.Lock()
		topic := req.ResponseTopic
		publisher, ok := publishers[topic]
		if !ok {
			publisher = client.Publisher(topic)
			publishers[topic] = publisher
		}
		mu.Unlock()

		var res Response = Response{
			ApiVersion: ApiVersionOrchestratorResponseV1,
			Metadata:   req.Metadata,
			ResultCode: result.Code(),
			Output:     base64.StdEncoding.EncodeToString([]byte(result.Output())),
		}

		err = respond(ctx, publisher, &res)
		if err != nil {
			logger.Error().Err(err).Msg("Encountered an internal error whilst responding to request")
		}
		return err
	}
}