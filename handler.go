package orchestrator

import (
	"context"

	"cloud.google.com/go/pubsub"
	logging "github.com/entur/go-logging"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/rs/zerolog"
)

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
	cache := NewTopicCache(client)

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
