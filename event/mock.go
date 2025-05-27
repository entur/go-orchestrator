package event

import (
	"encoding/base64"
	"encoding/json"

	cloudevent "github.com/cloudevents/sdk-go/v2/event"
	"github.com/entur/go-orchestrator"
)

const MockUserEmail = "mockuser@entur.io"

type MockEventOption func(*orchestrator.Request)

func NewMockEvent(manifest any, sender orchestrator.SenderType, action orchestrator.Action, opts ...MockEventOption) (*cloudevent.Event, error) {
	newManifest, err := json.Marshal(manifest)
	if err != nil {
		return nil, err
	}

	req := &orchestrator.Request{
		ApiVersion: "orchestrator.entur.io/request/v1",
		Metadata: orchestrator.OuterMetadata{
			RequestID: "mockid",
		},
		Sender: orchestrator.Sender{
			Type: sender,
		},
		Action:        action,
		ResponseTopic: "topic",
		Manifest: orchestrator.Manifests{
			Old: nil,
			New: newManifest,
		},
	}
	if sender == orchestrator.SenderTypeUser {
		req.Sender.Email = MockUserEmail
	}

	for _, opt := range opts {
		opt(req)
	}
	reqdata, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, base64.StdEncoding.EncodedLen(len(reqdata)))
	base64.StdEncoding.Encode(buf, reqdata)
	data, err := json.Marshal(&CloudEventData{
		Message: PubSubMessage{
			Data:        reqdata,
			ID:          "id",
			PublishTime: "time",
			Attributes:  PubSubMessageAttributes{},
		},
		Subscription: "sub",
	})
	if err != nil {
		return nil, err
	}

	e := cloudevent.New(cloudevent.CloudEventsVersionV03)
	e.DataEncoded = data
	return &e, nil
}
