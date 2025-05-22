package orchestrator

import (
	"context"
	"encoding/json"
	"errors"
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
// Processing
// -----------------------

func Receive(ctx context.Context, so Orchestrator, req Request) Result {
	logger := zerolog.Ctx(ctx)

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

type EventHandler func(context.Context, event.Event) error

func NewEventHandler(so Orchestrator, options ...HandlerOption) EventHandler {
	cfg := &HandlerConfig{}
	for _, opt := range options {
		opt(cfg)
	}

	var parentLogger zerolog.Logger
	if cfg.logger != nil {
		parentLogger = *cfg.logger
	} else {
		parentLogger = logging.New()
	}

	client, _ := pubsub.NewClient(context.Background(), so.ProjectID())
	cache := newTopicCache(client)

	parentLogger.Debug().Msg("Created a new EventHandler")

	return func(ctx context.Context, cloudEvent event.Event) error {
		logger := parentLogger.With().Logger()

		req, err := ParseEvent(cloudEvent)
		if err != nil {
			logger.Error().Err(err).Msg("Encountered an internal error when calling ParseEvent")
			return err
		}

		logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
			return c.Int("gorch_github_user_id", req.Sender.ID).
				Str("gorch_request_id", req.Metadata.RequestID).
				Str("gorch_file_name", req.Origin.FileName).
				Str("gorch_action", string(req.Action))
		})
		ctx = logger.WithContext(ctx)
		logger.Info().Interface("gorch_request", req).Msg("Ready to receive and process request")

		result := Receive(ctx, so, req)
		err = result.errs
		if err != nil {
			logger.Error().Stack().Err(err).
				Interface("gorch_result_creations", result.creations).
				Interface("gorch_result_updates", result.updates).
				Interface("gorch_result_deletions", result.deletions).
				Msg("Encountered an internal error whilst processing request")
		}

		res := NewResponse(req.Metadata, result.Code(), result.String())
		logger.Info().Interface("gorch_response", res).Msg("Ready to send response")

		topic := cache.TopicFullID(req.ResponseTopic)
		err = Respond(ctx, topic, res)
		if err != nil {
			logger.Error().Err(err).Msg("Encountered an internal error whilst responding to request")
		}

		return err
	}
}
