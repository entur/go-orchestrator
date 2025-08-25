package event

import (
	"context"
	"sync"

	"cloud.google.com/go/pubsub/v2"
	cloudevent "github.com/cloudevents/sdk-go/v2/event"
	"github.com/entur/go-logging"
	"github.com/entur/go-orchestrator"
	"github.com/rs/zerolog"
)

// -----------------------
// Helpers
// -----------------------

type publisherCache struct {
	mu     sync.Mutex
	client *pubsub.Client
	publishers map[string]*pubsub.Publisher
}

func (c *publisherCache) Publisher(name string) *pubsub.Publisher {
	c.mu.Lock()
	defer c.mu.Unlock()

	publisher, ok := c.publishers[name]
	if !ok {
		publisher = c.client.Publisher(name)
		c.publishers[name] = publisher
	}

	return publisher
}

func newPublisherCache(client *pubsub.Client) *publisherCache{
	return &publisherCache{
		client: client,
		publishers: map[string]*pubsub.Publisher{},
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
	cache := newPublisherCache(client)

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
				Msg("Encountered an internal error when processing request")
		}

		res := orchestrator.NewResponse(req.Metadata, result.Code(), result.String())
		publisher := cache.Publisher(req.ResponseTopic)

		err = orchestrator.Respond(ctx, publisher, res)
		if err != nil {
			logger.Error().Err(err).Msg("Encountered an internal error when responding to request")
		}
		return err
	}
}
