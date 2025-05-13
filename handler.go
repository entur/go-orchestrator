package orchestrator

import (
	"context"

	logging "github.com/entur/go-logging"

	"cloud.google.com/go/pubsub"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/rs/zerolog"
)

type EventHandlerConfig struct {
	logger *zerolog.Logger
}

type EventHandlerOption func(*EventHandlerConfig)

func WithCustomLogger(logger zerolog.Logger) EventHandlerOption {
	return func(c *EventHandlerConfig) {
		c.logger = &logger
	}
}

type EventHandler func(context.Context, event.Event) error

func NewEventHandler[T any](so Orchestrator[T], client *pubsub.Client, options ...EventHandlerOption) EventHandler {
	cfg := &EventHandlerConfig{}
	for _, opt := range options {
		opt(cfg)
	}
	
	var pLogger zerolog.Logger
	if cfg.logger != nil {
		pLogger = *cfg.logger
	} else {
		pLogger = logging.New()
	}

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
