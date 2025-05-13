package orchestrator

import (
	"context"
	"io"

	"cloud.google.com/go/pubsub"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/rs/zerolog"
)

type EventHandlerConfig struct {
	w io.Writer
}

type EventHandlerOption func(*EventHandlerConfig)

func WithCustomLogWriter(w io.Writer) EventHandlerOption {
	return func(c *EventHandlerConfig) {
		c.w = w
	}
}

type EventHandler func(context.Context, event.Event) error

func NewEventHandler[T any](so Orchestrator[T], client *pubsub.Client, options ...EventHandlerOption) EventHandler {
	cfg := &EventHandlerConfig{}
	for _, opt := range options {
		opt(cfg)
	}
	cache := NewTopicCache(client)
	logger := NewLogger(cfg.w)

	return func(ctx context.Context, cloudEvent event.Event) error {
		payload, err := ParseEvent[T](cloudEvent)
		if err != nil {
			logger.Error().Err(err).Msg("ParseEvent failed")
			return err
		}
		logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
			return c.Int("github_user_id", payload.Sender.ID).
				Str("request_id", payload.Metadata.RequestId).
				Str("file_name", payload.Origin.FileName).
				Str("action", string(payload.Action))
		})
		ctx = logger.WithContext(ctx)

		topic := cache.TopicFullID(payload.ResponseTopic)
		return Process(ctx, so, topic, payload)
	}
}
