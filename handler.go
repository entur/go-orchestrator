package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"cloud.google.com/go/pubsub"
	logging "github.com/entur/go-logging"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/rs/zerolog"
)

// -----------------------
// Helpers
// -----------------------

type ManifestHeader struct {
	ApiVersion ApiVersion `json:"apiVersion"`
	Kind       Kind       `json:"kind"`
}

type topicCache struct {
	mu     sync.Mutex
	client *pubsub.Client
	topics map[string]*pubsub.Topic
}

func (c *topicCache) Topics() []*pubsub.Topic {
	c.mu.Lock()
	defer c.mu.Unlock()

	var topics []*pubsub.Topic

	num := len(c.topics)
	if num > 0 {
		topics = make([]*pubsub.Topic, 0, num)
		for _, topic := range c.topics {
			topics = append(topics, topic)
		}
	}

	return topics
}

func (c *topicCache) Topic(projectID string, topicID string) *pubsub.Topic {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := projectID + topicID

	topic, ok := c.topics[key]
	if !ok {
		topic = c.client.TopicInProject(topicID, projectID)
		c.topics[key] = topic
	}

	return topic
}

func (c *topicCache) TopicFullID(id string) *pubsub.Topic {
	if !strings.HasPrefix(id, "projects/") {
		return nil
	}

	i := strings.Index(id[9:], "/")
	if i == -1 {
		return nil
	}

	projectID := id[9 : 9+i]
	topicID := id[strings.LastIndex(id, "/")+1:]

	return c.Topic(projectID, topicID)
}

func newTopicCache(client *pubsub.Client) *topicCache {
	return &topicCache{
		client: client,
		topics: map[string]*pubsub.Topic{},
	}
}

// -----------------------
// Handlers
// -----------------------

func HandleRequest(ctx context.Context, so Orchestrator, req Request) (res ResponseResult, err error) {
	logger := zerolog.Ctx(ctx)
	var header ManifestHeader
	err = json.Unmarshal(req.Manifest.New, &header)
	if err != nil {
		return res, err
	}
	for _, h := range so.Handlers() {
		if header.ApiVersion == h.ApiVersion() && header.Kind == h.Kind() {
			logger.Info().Msg(fmt.Sprintf("Found handler for %s %s", h.ApiVersion(), h.Kind()))
			before, ok := so.(OrchestratorMiddlewareBefore)
			if ok {
				logger.Info().Msg("Executing MiddlewareBefore")
				err = before.MiddlewareBefore(ctx, req, &res)
				if err != nil {
					// TODO: Wrap error
					return
				}
			}

			switch req.Action {
			case ActionApply:
				err = h.Apply(ctx, req, &res)
			case ActionPlan:
				err = h.Plan(ctx, req, &res)
			case ActionPlanDestroy:
				err = h.PlanDestroy(ctx, req, &res)
			case ActionDestroy:
				err = h.Destroy(ctx, req, &res)
			default:
				err = fmt.Errorf("TODO")
			}

			if err != nil {
				logger.Error().Err(err).Msg(fmt.Sprintf("Could not perform handler action %s", req.Action))
				// TODO: Wrap error
				return
			}
			logger.Info().Msg(fmt.Sprintf("Performed %s on %s %s", req.Action, h.ApiVersion(), h.Kind()))

			after, ok := so.(OrchestratorMiddlewareAfter)
			if ok {
				// TODO: Wrap error
				logger.Info().Msg("Executing MiddlewareBefore")
				err = after.MiddlewareAfter(ctx, req, &res)
			}
			return
		}
	}

	// TODO: better error?
	err = fmt.Errorf("found no matching handler")
	return
}

type HandlerConfig struct {
	logger *zerolog.Logger
}

type HandlerOption func(*HandlerConfig)

func WithCustomLogger(logger zerolog.Logger) HandlerOption {
	return func(c *HandlerConfig) {
		c.logger = &logger
	}
}

type EventHandler func(context.Context, event.Event) error

func NewEventHandler(so Orchestrator, options ...HandlerOption) EventHandler {
	cfg := &HandlerConfig{}
	for _, opt := range options {
		opt(cfg)
	}

	var pLogger zerolog.Logger
	if cfg.logger != nil {
		pLogger = *cfg.logger
	} else {
		pLogger = logging.New()
	}

	client, _ := pubsub.NewClient(context.Background(), so.ProjectID())
	cache := newTopicCache(client)
	pLogger.Info().Msg("Created a new EventHandler")

	return func(ctx context.Context, cloudEvent event.Event) error {
		logger := pLogger.With().Logger()

		req, err := ParseEvent(cloudEvent)
		if err != nil {
			logger.Error().Err(err).Msg("ParseEvent failed")
			return err
		}
		logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
			return c.Int("gorch_github_user_id", req.Sender.ID).
				Str("gorch_request_id", req.Metadata.RequestID).
				Str("gorch_file_name", req.Origin.FileName).
				Str("gorch_action", string(req.Action))
		})
		ctx = logger.WithContext(ctx)

		logger.Info().Interface("req", req).Msg("Handling request")

		result, err := HandleRequest(ctx, so, req)

		// TODO:
		// Cleanup all of this an/or split it into a new function.
		// Preferably the altter if we are making other event handlers
		var code ResultCode
		var msg string

		if err != nil || result.mistakes != nil {
			if result.mistakes != nil {
				logger.Error().Stack().Err(result.mistakes).Msg("")
			}
			if err != nil {
				logger.Error().Err(err).Interface("gorch_result", result).Msg(msg)
			}

			msg = "An internal error occured"
			code = ResultCodeError
		} else {
			msg = result.String()

			if !result.Succeeded() {
				code = ResultCodeFailure
			} else if !result.HasChanges() {
				code = ResultCodeNoop
			} else {
				code = ResultCodeSuccess
			}
		}

		res := NewResponse(req.Metadata, code, msg)

		logger.Info().Interface("res", res).Msg("Got response")

		enc, err := json.Marshal(res)
		if err != nil {
			return err
		}

		topic := cache.TopicFullID(req.ResponseTopic)
		if topic == nil {
			return fmt.Errorf("no topic set, cannot respond")
		}
		pubres := topic.Publish(ctx, &pubsub.Message{
			Data: enc,
		})
		_, err = pubres.Get(ctx)
		return err
	}
}
