package orchestrator

import (
	"context"
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
		topics := make([]string, 0, num)
		for _, topic := range topics {
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

func NewEventHandler[T any](so Orchestrator[T], options ...HandlerOption) EventHandler {
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

	return func(ctx context.Context, cloudEvent event.Event) error {
		logger := pLogger.With().Logger()
		payload, err := ParseEvent[T](cloudEvent)
		if err != nil {
			logger.Error().Err(err).Msg("ParseEvent failed")
			return err
		}

		logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
			return c.Int("gorch_github_user_id", payload.Sender.ID).
				Str("gorch_request_id", payload.Metadata.RequestID).
				Str("gorch_file_name", payload.Origin.FileName).
				Str("gorch_action", string(payload.Action))
		})
		ctx = logger.WithContext(ctx)

		topic := cache.TopicFullID(payload.ResponseTopic)
		return Process(ctx, so, topic, payload)
	}
}
