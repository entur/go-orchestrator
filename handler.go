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