package orchestrator

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"sync"

	"cloud.google.com/go/pubsub/v2"
	cloudevent "github.com/cloudevents/sdk-go/v2/event"
	"github.com/entur/go-logging"
	"github.com/rs/zerolog"
)

// -----------------------
// Cloud Event
// -----------------------

type PubSubMessageAttributes struct{}

type PubSubMessage struct {
	ID          string                  `json:"messageId"`
	PublishTime string                  `json:"publishTime"`
	Attributes  PubSubMessageAttributes `json:"attributes"`
	Data        []byte                  `json:"data"`
}

type CloudEventData struct {
	Subscription string
	Message      PubSubMessage
}

func UnmarshalCloudEvent(e cloudevent.Event, v any) error {
	var data CloudEventData
	err := e.DataAs(&data)
	if err != nil {
		return err
	}

	return json.Unmarshal(data.Message.Data, v)
}

// -----------------------
// Handler
// -----------------------

type HandlerConfig struct {
	client    *pubsub.Client
	clientSet bool
	logger    *zerolog.Logger
}

type HandlerOption func(*HandlerConfig)

func WithCustomLogger(logger zerolog.Logger) HandlerOption {
	return func(c *HandlerConfig) {
		c.logger = &logger
	}
}

func WithCustomPubSubClient(client *pubsub.Client) HandlerOption {
	return func(c *HandlerConfig) {
		c.client = client
		c.clientSet = true
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

	var client *pubsub.Client
	if cfg.clientSet {
		client = cfg.client
	} else {
		client, _ = pubsub.NewClient(context.Background(), so.ProjectID())
	}

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
		err = errors.Join(result.errs...)

		logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
			return c.Interface("gorch_result_summary", result.summary).
				Interface("gorch_result_creations", result.creations).
				Interface("gorch_result_updates", result.updates).
				Interface("gorch_result_deletions", result.deletions)
		})

		if client == nil {
			logger.Warn().Msg("Pubsub client is set to null, no responses will be sent")
		} else {
			mu.Lock()
			topic := req.ResponseTopic
			publisher, ok := publishers[topic]
			if !ok {
				publisher = client.Publisher(topic)
				publishers[topic] = publisher
			}
			mu.Unlock()

			var res = Response{
				ApiVersion: ApiVersionOrchestratorResponseV1,
				Metadata:   req.Metadata,
				ResultCode: result.Code(),
				Output:     base64.StdEncoding.EncodeToString([]byte(result.Output())),
			}

			err = errors.Join(err, respond(ctx, publisher, &res))
		}

		if err != nil {
			logger.Error().Err(err).Msg("Encountered an internal error during processing")
		}
		return err
	}
}
