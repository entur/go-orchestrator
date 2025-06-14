package event

import (
	"context"
	"strings"
	"sync"

	"cloud.google.com/go/pubsub"
	cloudevent "github.com/cloudevents/sdk-go/v2/event"
	"github.com/entur/go-logging"
	"github.com/entur/go-orchestrator"
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

type HandlerConfig struct {
	logger *zerolog.Logger
}

type HandlerOption func(*HandlerConfig)

func WithCustomLogger(logger zerolog.Logger) HandlerOption {
	return func(c *HandlerConfig) {
		c.logger = &logger
	}
}

type EventHandler func(context.Context, cloudevent.Event) error

func NewEventHandler(so orchestrator.Orchestrator, opts ...HandlerOption) EventHandler {
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
	cache := newTopicCache(client)

	parentLogger.Debug().Msg("Created a new EventHandler")
	return func(ctx context.Context, e cloudevent.Event) error {
		logger := parentLogger.With().Logger()

		req, err := ParseEvent(e)
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

		result := orchestrator.Receive(ctx, so, *req)
		err = result.AccumulatedError()
		if err != nil {
			logger.Error().Stack().Err(err).
				Interface("gorch_result_creations", result.Creations()).
				Interface("gorch_result_updates", result.Updates()).
				Interface("gorch_result_deletions", result.Deletions()).
				Msg("Encountered an internal error whilst processing request")
		}

		res := orchestrator.NewResponse(req.Metadata, result.Code(), result.String())
		topic := cache.TopicFullID(req.ResponseTopic)

		err = orchestrator.Respond(ctx, topic, res)
		if err != nil {
			logger.Error().Err(err).Msg("Encountered an internal error whilst responding to request")
		}
		return err
	}
}
